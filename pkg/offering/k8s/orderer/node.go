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

package k8sorderer

import (
	"context"
	"fmt"
	"strings"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	commoninit "github.com/IBM-Blockchain/fabric-operator/pkg/initializer/common"
	k8sclient "github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources"
	resourcemanager "github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources/manager"
	baseorderer "github.com/IBM-Blockchain/fabric-operator/pkg/offering/base/orderer"
	baseoverride "github.com/IBM-Blockchain/fabric-operator/pkg/offering/base/orderer/override"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/common"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/k8s/orderer/override"
	"github.com/IBM-Blockchain/fabric-operator/pkg/operatorerrors"
	"github.com/IBM-Blockchain/fabric-operator/pkg/util"
	"github.com/IBM-Blockchain/fabric-operator/version"
	"github.com/pkg/errors"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Override interface {
	baseorderer.Override
	Ingress(v1.Object, *networkingv1.Ingress, resources.Action) error
	Ingressv1beta1(v1.Object, *networkingv1beta1.Ingress, resources.Action) error
}

var _ baseorderer.IBPOrderer = &Node{}

type Node struct {
	*baseorderer.Node

	IngressManager        resources.Manager
	Ingressv1beta1Manager resources.Manager

	Override Override
}

func NewNode(basenode *baseorderer.Node) *Node {
	node := &Node{
		Node: basenode,
		Override: &override.Override{
			Override: baseoverride.Override{
				Name:   basenode.Name,
				Client: basenode.Client,
				Config: basenode.Config,
			},
		},
	}
	node.CreateManagers()
	return node
}

func (n *Node) CreateManagers() {
	override := n.Override
	resourceManager := resourcemanager.New(n.Client, n.Scheme)
	n.IngressManager = resourceManager.CreateIngressManager("", override.Ingress, n.GetLabels, n.Config.OrdererInitConfig.IngressFile)
	n.Ingressv1beta1Manager = resourceManager.CreateIngressv1beta1Manager("", override.Ingressv1beta1, n.GetLabels, n.Config.OrdererInitConfig.Ingressv1beta1File)
}

func (n *Node) Reconcile(instance *current.IBPOrderer, update baseorderer.Update) (common.Result, error) {
	var err error
	var status *current.CRStatus

	log.Info(fmt.Sprintf("Reconciling node instance '%s' ... update: %+v", instance.Name, update))

	versionSet, err := n.SetVersion(instance)
	if err != nil {
		return common.Result{}, errors.Wrap(err, fmt.Sprintf("failed updating CR '%s' to version '%s'", instance.Name, version.Operator))
	}
	if versionSet {
		log.Info("Instance version updated, requeuing request...")
		return common.Result{
			Result: reconcile.Result{
				Requeue: true,
			},
			OverrideUpdateStatus: true,
		}, nil
	}

	instanceUpdated, err := n.PreReconcileChecks(instance, update)
	if err != nil {
		return common.Result{}, errors.Wrap(err, "failed pre reconcile checks")
	}
	externalEndpointUpdated := n.UpdateExternalEndpoint(instance)

	if instanceUpdated || externalEndpointUpdated {
		log.Info(fmt.Sprintf("Updating instance after pre reconcile checks: %t, updating external endpoint: %t",
			instanceUpdated, externalEndpointUpdated))

		err = n.Client.Patch(context.TODO(), instance, nil, k8sclient.PatchOption{
			Resilient: &k8sclient.ResilientPatch{
				Retry:    3,
				Into:     &current.IBPOrderer{},
				Strategy: client.MergeFrom,
			},
		})
		if err != nil {
			return common.Result{}, errors.Wrap(err, "failed to update instance")
		}

		log.Info("Instance updated during reconcile checks, request will be requeued...")
		return common.Result{
			Result: reconcile.Result{
				Requeue: true,
			},
			Status: &current.CRStatus{
				Type:    current.Initializing,
				Reason:  "Setting default values for either zone, region, and/or external endpoint",
				Message: "Operator has updated spec with defaults as part of initialization",
			},
			OverrideUpdateStatus: true,
		}, nil
	}

	err = n.Initialize(instance, update)
	if err != nil {
		return common.Result{}, operatorerrors.Wrap(err, operatorerrors.OrdererInitilizationFailed, "failed to initialize orderer node")
	}

	if update.OrdererTagUpdated() {
		if err := n.ReconcileFabricOrdererMigration(instance); err != nil {
			return common.Result{}, operatorerrors.Wrap(err, operatorerrors.FabricOrdererMigrationFailed, "failed to migrate fabric orderer versions")
		}
	}

	if update.MigrateToV2() {
		if err := n.FabricOrdererMigrationV2_0(instance); err != nil {
			return common.Result{}, operatorerrors.Wrap(err, operatorerrors.FabricOrdererMigrationFailed, "failed to migrate fabric orderer to version v2.x")
		}
	}

	if update.MigrateToV24() {
		if err := n.FabricOrdererMigrationV2_4(instance); err != nil {
			return common.Result{}, operatorerrors.Wrap(err, operatorerrors.FabricOrdererMigrationFailed, "failed to migrate fabric orderer to version v2.4.x")
		}
	}

	err = n.ReconcileManagers(instance, update, nil)
	if err != nil {
		return common.Result{}, errors.Wrap(err, "failed to reconcile managers")
	}

	err = n.UpdateConnectionProfile(instance)
	if err != nil {
		return common.Result{}, errors.Wrap(err, "failed to create connection profile")
	}

	err = n.CheckStates(instance)
	if err != nil {
		return common.Result{}, errors.Wrap(err, "failed to check and restore state")
	}

	err = n.UpdateParentStatus(instance)
	if err != nil {
		return common.Result{}, errors.Wrap(err, "failed to update parent's status")
	}

	status, result, err := n.CustomLogic(instance, update)
	if err != nil {
		return common.Result{}, errors.Wrap(err, "failed to run custom offering logic		")
	}
	if result != nil {
		log.Info(fmt.Sprintf("Finished reconciling '%s' with Custom Logic result", instance.GetName()))
		return *result, nil
	}

	if update.EcertUpdated() {
		log.Info("Ecert was updated")
		// Request deployment restart for tls cert update
		err = n.Restart.ForCertUpdate(commoninit.ECERT, instance)
		if err != nil {
			return common.Result{}, errors.Wrap(err, "failed to update restart config")
		}
	}

	if update.TLSCertUpdated() {
		log.Info("TLS cert was updated")
		// Request deployment restart for ecert update
		err = n.Restart.ForCertUpdate(commoninit.TLS, instance)
		if err != nil {
			return common.Result{}, errors.Wrap(err, "failed to update restart config")
		}
	}

	if update.MSPUpdated() {
		if err = n.UpdateMSPCertificates(instance); err != nil {
			return common.Result{}, errors.Wrap(err, "failed to update certificates passed in MSP spec")
		}
	}

	if err := n.HandleActions(instance, update); err != nil {
		return common.Result{}, err
	}

	if err := n.HandleRestart(instance, update); err != nil {
		return common.Result{}, err
	}

	return common.Result{
		Status: status,
	}, nil
}

func (n *Node) ReconcileManagers(instance *current.IBPOrderer, updated baseorderer.Update, genesisBlock []byte) error {
	err := n.Node.ReconcileManagers(instance, updated, genesisBlock)
	if err != nil {
		return err
	}

	update := updated.SpecUpdated()

	err = n.ReconcileIngressManager(instance, update)
	if err != nil {
		return errors.Wrap(err, "failed Ingress reconciliation")
	}

	return nil
}

func (n *Node) ReconcileIngressManager(instance *current.IBPOrderer, update bool) error {
	if n.Config.Operator.Globals.AllowKubernetesEighteen == "true" {
		// check k8s version
		version, err := util.GetServerVersion()
		if err != nil {
			return err
		}
		if strings.Compare(version.Minor, "19") < 0 { // v1beta
			err = n.Ingressv1beta1Manager.Reconcile(instance, update)
			if err != nil {
				return errors.Wrap(err, "failed Ingressv1beta1 reconciliation")
			}
		} else {
			err = n.IngressManager.Reconcile(instance, update)
			if err != nil {
				return errors.Wrap(err, "failed Ingress reconciliation")
			}
		}
	} else {
		err := n.IngressManager.Reconcile(instance, update)
		if err != nil {
			return errors.Wrap(err, "failed Ingress reconciliation")
		}
	}
	return nil
}
