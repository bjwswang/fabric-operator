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
	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	cmocks "github.com/IBM-Blockchain/fabric-operator/controllers/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("RBACManager tests", func() {
	var mockKubeClient *cmocks.Client

	BeforeEach(func() {
		mockKubeClient = &cmocks.Client{}
	})

	It("manager reconfile", func() {
		overrides := make(map[Resource]Synchronizer)
		overrides[Federation] = EmptySynchronizer
		mgr := NewRBACManager(mockKubeClient, overrides)

		instance := &current.Federation{
			ObjectMeta: v1.ObjectMeta{
				Name: "fed-sample",
			},
		}

		err := mgr.Reconcile(Resource("notresource"), instance, ResourceCreate)
		Expect(err).To(Equal(ErrResouceHasNoSynchronizer))

		err = mgr.Reconcile(Federation, instance, ResourceCreate)
		Expect(err).To(BeNil())
	})
})
