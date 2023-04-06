/*
 * Copyright contributors to the Hyperledger Fabric Operator project
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 * 	  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package network

import (
	"context"
	"fmt"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	config "github.com/IBM-Blockchain/fabric-operator/operatorconfig"
	"github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources"
	resourcemanager "github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources/manager"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/common"
	"github.com/IBM-Blockchain/fabric-operator/pkg/operatorerrors"
	bcrbac "github.com/IBM-Blockchain/fabric-operator/pkg/rbac"
	"github.com/IBM-Blockchain/fabric-operator/pkg/user"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("base_network")

//go:generate counterfeiter -o mocks/update.go -fake-name Update . Update

type Update interface {
	SpecUpdated() bool
	MemberUpdated() bool
	OrdererCreate() bool
	OrdererStatusUpdated() bool
}

//go:generate counterfeiter -o mocks/override.go -fake-name Override . Override

type Override interface {
	Orderer(v1.Object, *current.IBPOrderer, resources.Action) error
}

//go:generate counterfeiter -o mocks/basenetwork.go -fake-name Network . Network

type Network interface {
	PreReconcileChecks(instance *current.Network, update Update) error
	Initialize(instance *current.Network, update Update) error
	ReconcileManagers(instance *current.Network, update Update) error
	CheckStates(instance *current.Network, update Update) (common.Result, error)
	Reconcile(instance *current.Network, update Update) (common.Result, error)
}

var _ Network = (*BaseNetwork)(nil)

const (
	KIND = "NETWORK"
)

type BaseNetwork struct {
	Client controllerclient.Client
	Scheme *runtime.Scheme

	Config *config.Config

	Override Override

	RBACManager    *bcrbac.Manager
	OrdererManager resources.Manager
}

func New(client controllerclient.Client, scheme *runtime.Scheme, config *config.Config, o Override) *BaseNetwork {
	base := &BaseNetwork{
		Client:   client,
		Scheme:   scheme,
		Config:   config,
		Override: o,
	}

	base.CreateManagers()

	return base
}

// TODO: leave this due to we might need managers in the future
// - configmap manager
func (network *BaseNetwork) CreateManagers() {
	network.RBACManager = bcrbac.NewRBACManager(network.Client, nil)
	override := network.Override
	mgr := resourcemanager.New(network.Client, network.Scheme)
	network.OrdererManager = mgr.CreateOrdererManager("", override.Orderer, network.GetLabels, network.Config.NetworkInitConfig.OrdererFile)
}

// Reconcile on Network upon Update
func (network *BaseNetwork) Reconcile(instance *current.Network, update Update) (common.Result, error) {
	var err error

	if err = network.PreReconcileChecks(instance, update); err != nil {
		return common.Result{}, errors.Wrap(err, "failed on prereconcile checks")
	}

	if err = network.Initialize(instance, update); err != nil {
		return common.Result{}, operatorerrors.Wrap(err, operatorerrors.NetworkInitializationFailed, "failed to initialize network")
	}
	if err = network.ReconcileUser(instance, update); err != nil {
		return common.Result{}, operatorerrors.Wrap(err, operatorerrors.NetworkInitializationFailed, "failed to reconcileUser")
	}
	if err = network.ReconcileManagers(instance, update); err != nil {
		return common.Result{}, errors.Wrap(err, "failed to reconcile managers")
	}

	return network.CheckStates(instance, update)
}

// PreReconcileChecks on Network upon Update
func (network *BaseNetwork) PreReconcileChecks(instance *current.Network, update Update) error {
	log.Info(fmt.Sprintf("PreReconcileChecks on Network %s", instance.GetName()))

	if !instance.HasOrder() {
		return errors.New("network's order is empty")
	}

	// Federation & Member check
	if !instance.HasFederation() {
		return errors.New("network's federation is empty")
	}

	if !instance.HasMembers() {
		return errors.New("network's members is empty")
	}

	// Federation status must be at `Activated`
	federation, err := network.GetFederation(instance)
	if err != nil {
		return errors.Wrap(err, "failed to find the dependent federation")
	}

	if federation.Status.Type != current.FederationActivated {
		return errors.Errorf("the dependent federation %s is not activated yet", federation.GetName())
	}

	// Network only can contain members inherited from Federation
	added, _ := current.DifferMembers(federation.GetMembers(), instance.GetMembers())
	if len(added) != 0 {
		return errors.Errorf("network %s contains members %v which not in Federation %s", instance.GetName(), added, federation.GetName())
	}

	return nil
}

// Initialize on Network upon Update
func (network *BaseNetwork) Initialize(instance *current.Network, update Update) error {
	return nil
}

// ReconcileManagers on Network upon Update
func (network *BaseNetwork) ReconcileManagers(instance *current.Network, update Update) error {
	var err error
	if update.MemberUpdated() {
		err = network.RBACManager.Reconcile(bcrbac.Network, instance, bcrbac.ResourceUpdate)
		if err != nil {
			return err
		}
	}
	log.Info(fmt.Sprintf("ReconcileManagers on Order %s", instance.GetName()))
	return network.OrdererManager.Reconcile(instance, false)
}

// CheckStates on Network
func (network *BaseNetwork) CheckStates(instance *current.Network, update Update) (common.Result, error) {
	status := instance.Status.CRStatus
	if !instance.HasType() {
		status.Type = current.Created
		status.Status = current.True
	}

	if update.OrdererStatusUpdated() && instance.Status.Type != current.Error {
		orderer := &current.IBPOrderer{}
		orderer.Name = instance.GetOrdererName()
		orderer.Namespace = instance.GetOrdererNamespace()
		if err := network.Client.Get(context.TODO(), client.ObjectKeyFromObject(orderer), orderer); err == nil {
			return common.Result{Status: &orderer.Status.CRStatus}, nil
		}
	}

	return common.Result{
		Status: &status,
	}, nil
}

// GetLabels from instance.GetLabels
func (network *BaseNetwork) GetLabels(instance v1.Object) map[string]string {
	return instance.GetLabels()
}

func (network *BaseNetwork) GetFederation(instance *current.Network) (*current.Federation, error) {
	federation := &current.Federation{}

	err := network.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Federation}, federation)
	if err != nil {
		return nil, err
	}
	return federation, nil
}

// ReconcileUser on Network upon Update
func (network *BaseNetwork) ReconcileUser(instance *current.Network, update Update) (err error) {
	if !network.Config.OrganizationInitConfig.IAMEnabled || !update.OrdererCreate() {
		return nil
	}
	org := &current.Organization{ObjectMeta: v1.ObjectMeta{Name: instance.GetInitiatorMember()}}
	if err = network.Client.Get(context.TODO(), client.ObjectKeyFromObject(org), org); err != nil {
		return err
	}
	targetUser := org.Spec.Admin
	size := instance.Spec.OrderSpec.ClusterSize
	enrollIDs := make([]string, size)
	for i := range enrollIDs {
		enrollIDs[i] = fmt.Sprintf("%s%d", instance.Name, i)
	}
	err = user.ReconcileMultiple(network.Client, targetUser, org.Name, user.ORDERER, user.Add, enrollIDs...)
	if err != nil {
		return err
	}
	return nil
}
