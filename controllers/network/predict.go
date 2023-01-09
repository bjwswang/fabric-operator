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
	"reflect"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	bcrbac "github.com/IBM-Blockchain/fabric-operator/pkg/rbac"
	"github.com/IBM-Blockchain/fabric-operator/pkg/user"
	"github.com/go-test/deep"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func (r *ReconcileNetwork) CreateFunc(e event.CreateEvent) bool {
	network := e.Object.(*current.Network)

	log.Info(fmt.Sprintf("Create event detected for network '%s'", network.GetName()))
	update := Update{}

	if network.HasType() {
		log.Info(fmt.Sprintf("Operator restart detected, running update flow on existing network '%s'", network.GetName()))

		// Get the spec state of the resource before the operator went down, this
		// will be used to compare to see if the spec of resources has changed
		cm, err := r.GetSpecState(network)
		if err != nil {
			log.Info(fmt.Sprintf("Failed getting saved fedeation spec '%s', triggering create: %s", network.GetName(), err.Error()))
			return true
		}

		specBytes := cm.BinaryData["spec"]
		existingNet := &current.Network{}
		err = yaml.Unmarshal(specBytes, &existingNet.Spec)
		if err != nil {
			log.Info(fmt.Sprintf("Unmarshal failed for saved network spec '%s', triggering create: %s", network.GetName(), err.Error()))
			return true
		}

		diff := deep.Equal(network.Spec, existingNet.Spec)
		if diff != nil {
			log.Info(fmt.Sprintf("Network '%s' spec was updated while operator was down", network.GetName()))
			log.Info(fmt.Sprintf("Difference detected: %v", diff))
			update.specUpdated = true
		}

		added, removed := current.DifferMembers(network.Spec.Members, existingNet.Spec.Members)
		if len(added) != 0 || len(removed) != 0 {
			log.Info(fmt.Sprintf("Network '%s' members was updated while operator was down", network.GetName()))
			log.Info(fmt.Sprintf("Difference detected: added members %v", added))
			log.Info(fmt.Sprintf("Difference detected: removed members %v", removed))
			update.memberUpdated = true
		}

		log.Info(fmt.Sprintf("Create event triggering reconcile for updating Network '%s'", network.GetName()))
		r.PushUpdate(network.GetName(), update)
		return true
	}

	update.specUpdated = true
	update.memberUpdated = true
	update.ordererCreate = true

	r.PushUpdate(network.GetName(), update)

	return true
}

func (r *ReconcileNetwork) UpdateFunc(e event.UpdateEvent) bool {
	oldNet := e.ObjectOld.(*current.Network)
	newNet := e.ObjectNew.(*current.Network)
	log.Info(fmt.Sprintf("Update event detected for network '%s'", oldNet.GetName()))
	update := Update{}

	if reflect.DeepEqual(oldNet.Spec, newNet.Spec) {
		return false
	}

	update.specUpdated = true

	added, removed := current.DifferMembers(oldNet.GetMembers(), newNet.GetMembers())
	if len(added) != 0 || len(removed) != 0 {
		log.Info(fmt.Sprintf("Difference detected: added members %v", added))
		log.Info(fmt.Sprintf("Difference detected: removed members %v", removed))
		update.memberUpdated = true
	}

	r.PushUpdate(oldNet.GetName(), update)

	log.Info(fmt.Sprintf("Spec update triggering reconcile on Network custom resource %s: update [ %+v ]", oldNet.Name, update.GetUpdateStackWithTrues()))

	return true
}

func (r *ReconcileNetwork) DeleteFunc(e event.DeleteEvent) bool {
	network := e.Object.(*current.Network)
	err := r.rbacManager.Reconcile(bcrbac.Network, network, bcrbac.ResourceDelete)
	if err != nil {
		log.Error(err, "failed to sync rbac uppon proposal delete")
	}
	if !r.Config.OrganizationInitConfig.IAMEnabled {
		return false
	}
	org := &current.Organization{}
	if err = r.client.Get(context.TODO(), types.NamespacedName{Name: network.GetInitiatorMember().Name, Namespace: network.GetInitiatorMember().Namespace}, org); err != nil {
		log.Error(err, "failed to get org when network delete")
		return false
	}
	targetUser := org.Spec.Admin
	for i := 0; i < network.Spec.OrderSpec.ClusterSize; i++ {
		enrollID := fmt.Sprintf("%s%d", network.Name, i)
		err = user.Reconcile(r.client, targetUser, org.Name, enrollID, user.ORDERER, user.Remove)
		if err != nil {
			log.Error(err, "failed to reconcile user when network delete")
		}
	}
	return false
}
