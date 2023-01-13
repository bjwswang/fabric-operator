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

package channel

import (
	"context"
	"fmt"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	config "github.com/IBM-Blockchain/fabric-operator/operatorconfig"
	chaninit "github.com/IBM-Blockchain/fabric-operator/pkg/initializer/channel"
	"github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/common"
	"github.com/IBM-Blockchain/fabric-operator/pkg/operatorerrors"
	bcrbac "github.com/IBM-Blockchain/fabric-operator/pkg/rbac"
	"github.com/IBM-Blockchain/fabric-operator/version"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("base_channel")

//go:generate counterfeiter -o mocks/update.go -fake-name Update . Update

type Update interface {
	SpecUpdated() bool
	MemberUpdated() bool
}

//go:generate counterfeiter -o mocks/override.go -fake-name Override . Override

type Override interface{}

//go:generate counterfeiter -o mocks/basechannel.go -fake-name Channel . Channel

type Channel interface {
	PreReconcileChecks(instance *current.Channel, update Update) error
	Initialize(instance *current.Channel, update Update) error
	ReconcileManagers(instance *current.Channel, update Update) error
	CheckStates(instance *current.Channel, update Update) (common.Result, error)
	Reconcile(instance *current.Channel, update Update) (common.Result, error)
}

var _ Channel = (*BaseChannel)(nil)

const (
	KIND = "CHANNEL"
)

type BaseChannel struct {
	Client controllerclient.Client
	Scheme *runtime.Scheme

	Config *config.Config

	Override Override

	Initializer *chaninit.Initializer

	RBACManager *bcrbac.Manager
}

func New(client controllerclient.Client, scheme *runtime.Scheme, config *config.Config, o Override) *BaseChannel {
	base := &BaseChannel{
		Client:   client,
		Scheme:   scheme,
		Config:   config,
		Override: o,
	}

	base.Initializer = chaninit.New(client, scheme, config.ChannelInitConfig)

	base.CreateManagers()

	return base
}

func (channel *BaseChannel) CreateManagers() {
	channel.RBACManager = bcrbac.NewRBACManager(channel.Client, nil)
}

// Reconcile on Channel upon Update
func (channel *BaseChannel) Reconcile(instance *current.Channel, update Update) (common.Result, error) {
	var err error

	if err = channel.PreReconcileChecks(instance, update); err != nil {
		return common.Result{}, errors.Wrap(err, "failed on prereconcile checks")
	}

	if err = channel.Initialize(instance, update); err != nil {
		return common.Result{}, operatorerrors.Wrap(err, operatorerrors.ChannelInitializationFailed, "failed to initialize channel")
	}

	if err = channel.ReconcileManagers(instance, update); err != nil {
		return common.Result{}, errors.Wrap(err, "failed to reconcile managers")
	}

	return channel.CheckStates(instance, update)
}

// PreReconcileChecks on Channel upon Update
func (channel *BaseChannel) PreReconcileChecks(instance *current.Channel, update Update) error {
	var err error
	log.Info(fmt.Sprintf("PreReconcileChecks on Channel %s", instance.GetName()))

	if !instance.HasNetwork() {
		return errors.New("channel's network is empty")
	}

	if !instance.HashMembers() {
		return errors.New("channel has no members")
	}

	// make sure channel members is the subset of network's members
	network := &current.Network{}
	err = channel.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Network}, network)
	if err != nil {
		return errors.Wrap(err, "get channel's network")
	}
	if network.Status.Type != current.Created {
		return errors.Errorf("network %s not created yet", network.GetName())
	}
	members := make(map[string]struct{})
	for _, m := range network.GetMembers() {
		members[m.Name] = struct{}{}
	}
	for _, m := range instance.GetMembers() {
		_, ok := members[m.Name]
		if !ok {
			return errors.Errorf("channel member %s not a network member", m.Name)
		}
	}

	return nil
}

// Initialize on Channel upon Update
func (baseChan *BaseChannel) Initialize(instance *current.Channel, update Update) error {
	err := baseChan.Initializer.CreateOrUpdateChannel(instance)
	if err != nil {
		return err
	}
	return nil
}

// ReconcileManagers on Channel upon Update
func (baseChan *BaseChannel) ReconcileManagers(instance *current.Channel, update Update) error {
	var err error

	// set channel's owner reference to its network
	err = baseChan.SetOwnerReference(instance, update)
	if err != nil {
		return err
	}

	// reconcile channel member's rbac
	err = baseChan.ReconcileRBAC(instance, update)
	if err != nil {
		return err
	}
	return nil
}

func (baseChan *BaseChannel) SetOwnerReference(instance *current.Channel, update Update) error {
	var err error

	network := &current.Network{}
	err = baseChan.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Network}, network)
	if err != nil {
		return errors.Wrap(err, "get channel's network")
	}
	ownerReference := bcrbac.OwnerReference(bcrbac.Network, network)

	var exist bool
	for _, reference := range instance.OwnerReferences {
		if reference.UID == ownerReference.UID {
			exist = true
			break
		}
	}
	if !exist {
		instance.OwnerReferences = []v1.OwnerReference{bcrbac.OwnerReference(bcrbac.Network, network)}

		err = baseChan.Client.Update(context.TODO(), instance)
		if err != nil {
			return err
		}
	}

	return nil
}

// ReconcileRBAC will sync current channel to every member's AdminClusterRole
func (baseChan *BaseChannel) ReconcileRBAC(instance *current.Channel, update Update) error {
	if update.MemberUpdated() && baseChan.Config.OrganizationInitConfig.IAMEnabled {
		err := baseChan.RBACManager.Reconcile(bcrbac.Channel, instance, bcrbac.ResourceUpdate)
		if err != nil {
			return err
		}
	}
	return nil
}

// CheckStates on Channel
func (baseChan *BaseChannel) CheckStates(instance *current.Channel, update Update) (common.Result, error) {
	return common.Result{
		Status: &current.CRStatus{
			Type:    current.ChannelCreated,
			Version: version.Operator,
		},
	}, nil
}

// GetLabels from instance.GetLabels
func (baseChan *BaseChannel) GetLabels(instance v1.Object) map[string]string {
	return instance.GetLabels()
}
