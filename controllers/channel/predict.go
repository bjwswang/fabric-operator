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
	"fmt"
	"reflect"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/go-test/deep"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func (r *ReconcileChannel) CreateFunc(e event.CreateEvent) bool {
	channel := e.Object.(*current.Channel)
	log.Info(fmt.Sprintf("Create event detected for channel '%s'", channel.GetName()))

	update := Update{}

	if channel.HasType() {
		log.Info(fmt.Sprintf("Operator restart detected, running update flow on existing channel '%s'", channel.GetName()))

		// Get the spec state of the resource before the operator went down, this
		// will be used to compare to see if the spec of resources has changed
		cm, err := r.GetSpecState(channel)
		if err != nil {
			log.Info(fmt.Sprintf("Failed getting saved channel spec '%s', triggering create: %s", channel.GetName(), err.Error()))
			return true
		}

		specBytes := cm.BinaryData["spec"]
		existingChannel := &current.Channel{}
		err = yaml.Unmarshal(specBytes, &existingChannel.Spec)
		if err != nil {
			log.Info(fmt.Sprintf("Unmarshal failed for saved channel spec '%s', triggering create: %s", channel.GetName(), err.Error()))
			return true
		}

		diff := deep.Equal(channel.Spec, existingChannel.Spec)
		if diff != nil {
			log.Info(fmt.Sprintf("Channel '%s' spec was updated while operator was down", channel.GetName()))
			log.Info(fmt.Sprintf("Difference detected: %v", diff))
			update.specUpdated = true
		}

		added, removed := current.DifferMembers(channel.Spec.Members, existingChannel.Spec.Members)
		if len(added) != 0 || len(removed) != 0 {
			log.Info(fmt.Sprintf("Channel '%s' members was updated while operator was down", channel.GetName()))
			log.Info(fmt.Sprintf("Difference detected: added members %v", added))
			log.Info(fmt.Sprintf("Difference detected: removed members %v", removed))
			update.memberUpdated = true
		}

		log.Info(fmt.Sprintf("Create event triggering reconcile for updating Channel '%s'", channel.GetName()))
		r.PushUpdate(channel.GetName(), update)
		return true
	}

	update.specUpdated = true
	update.memberUpdated = true
	r.PushUpdate(channel.GetName(), update)

	return true
}

func (r *ReconcileChannel) UpdateFunc(e event.UpdateEvent) bool {
	oldChan := e.ObjectOld.(*current.Channel)
	newChan := e.ObjectNew.(*current.Channel)
	log.Info(fmt.Sprintf("Update event detected for channel '%s'", oldChan.GetName()))

	update := Update{}

	if reflect.DeepEqual(oldChan.Spec, newChan.Spec) {
		return false
	}

	update.specUpdated = true

	added, removed := current.DifferMembers(oldChan.GetMembers(), newChan.GetMembers())
	if len(added) != 0 || len(removed) != 0 {
		log.Info(fmt.Sprintf("Difference detected: added members %v", added))
		log.Info(fmt.Sprintf("Difference detected: removed members %v", removed))
		update.memberUpdated = true
	}

	r.PushUpdate(oldChan.GetName(), update)

	log.Info(fmt.Sprintf("Spec update triggering reconcile on Channel custom resource %s: update [ %+v ]", oldChan.Name, update.GetUpdateStackWithTrues()))

	return true
}

func (r *ReconcileChannel) PeerCreateFunc(e event.CreateEvent) bool {
	return false
}

func (r *ReconcileChannel) PeerUpdateFunc(e event.UpdateEvent) bool {
	return false
}

func (r *ReconcileChannel) PeerDeleteFunc(e event.DeleteEvent) bool {
	return false
}
