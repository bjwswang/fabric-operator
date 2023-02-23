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
	"fmt"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
)

// Update defines a list of elements that we detect spec updates on
type Update struct {
	specUpdated          bool
	ordererCreate        bool
	ordererStatusUpdated bool
}

func (u *Update) SpecUpdated() bool {
	return u.specUpdated
}

func (u *Update) OrdererCreate() bool {
	return u.ordererCreate
}
func (u *Update) OrdererStatusUpdated() bool {
	return u.ordererStatusUpdated
}

// GetUpdateStackWithTrues is a helper method to print updates that have been detected
func (u *Update) GetUpdateStackWithTrues() string {
	stack := ""

	if u.specUpdated {
		stack += "specUpdated "
	}

	if u.ordererCreate {
		stack += "ordererCreate"
	}

	if u.ordererStatusUpdated {
		stack += "ordererStatusUpdated"
	}

	if len(stack) == 0 {
		stack = "emptystack "
	}

	return stack
}

// GetUpdateStatus with index 0
func (r *ReconcileNetwork) GetUpdateStatus(instance *current.Network) *Update {
	return r.GetUpdateStatusAtElement(instance, 0)
}

func (r *ReconcileNetwork) GetUpdateStatusAtElement(instance *current.Network, index int) *Update {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	update := Update{}
	_, ok := r.update[instance.GetName()]
	if !ok {
		return &update
	}

	if len(r.update[instance.GetName()]) >= 1 {
		update = r.update[instance.GetName()][index]
	}

	return &update
}

func (r *ReconcileNetwork) PushUpdate(instance string, update Update) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.update[instance] = AppendUpdateIfMissing(r.update[instance], update)
}

func (r *ReconcileNetwork) PopUpdate(instance string) *Update {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	update := Update{}
	if len(r.update[instance]) >= 1 {
		update = r.update[instance][0]
		if len(r.update[instance]) == 1 {
			r.update[instance] = []Update{}
		} else {
			r.update[instance] = r.update[instance][1:]
		}
	}

	return &update
}

func AppendUpdateIfMissing(updates []Update, update Update) []Update {
	for _, u := range updates {
		if u == update {
			return updates
		}
	}
	return append(updates, update)
}

func GetUpdateStack(allUpdates map[string][]Update) string {
	stack := ""

	for orderer, updates := range allUpdates {
		currentStack := ""
		for index, update := range updates {
			currentStack += fmt.Sprintf("{ %s}", update.GetUpdateStackWithTrues())
			if index != len(updates)-1 {
				currentStack += " , "
			}
		}
		stack += fmt.Sprintf("%s: [ %s ] ", orderer, currentStack)
	}

	return stack
}
