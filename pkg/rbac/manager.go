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

package rbac

import (
	"errors"

	"github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ErrResouceHasNoSynchronizer = errors.New("resource has no synchronizer defined")
)

// ResourceAction defines possible actions on Resource
type ResourceAction int

const (
	ResourceCreate ResourceAction = iota
	ResourceUpdate
	ResourceDelete
)

// Manager help to reconcile uppon resource actions
type Manager struct {
	Client controllerclient.Client

	synchronizers map[Resource]Synchronizer
}

// NewRBACManager initialize a Manager instance to reconcile on Resources' rbac
func NewRBACManager(client controllerclient.Client, overrides map[Resource]Synchronizer) *Manager {
	synchronizers := make(map[Resource]Synchronizer)
	for resource, synchronizer := range defaultSynchronizers {
		synchronizers[resource] = synchronizer
	}
	for resource, synchronizer := range overrides {
		synchronizers[resource] = synchronizer
	}
	return &Manager{
		Client:        client,
		synchronizers: synchronizers,
	}
}

// Reconcile on resource uppon action(create/update/delete)
func (s *Manager) Reconcile(resouce Resource, instance v1.Object, action ResourceAction) error {
	synchronizer, ok := s.synchronizers[resouce]
	if !ok {
		return ErrResouceHasNoSynchronizer
	}
	return synchronizer(s.Client, instance, action)
}
