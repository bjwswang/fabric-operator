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
	"errors"
	"sync"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/controllers/mocks"
	mockedreconcile "github.com/IBM-Blockchain/fabric-operator/controllers/network/mocks"
	"github.com/IBM-Blockchain/fabric-operator/operatorconfig"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/common"
	"github.com/IBM-Blockchain/fabric-operator/pkg/rbac"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("ReconcileFederation", func() {

	var (
		client      *mocks.Client
		reconciler  *ReconcileNetwork
		k8soffering *mockedreconcile.NetworkReconcile
		request     reconcile.Request
		network     *current.Network
	)

	BeforeEach(func() {
		network = &current.Network{
			ObjectMeta: metav1.ObjectMeta{
				Name: "network-sample",
			},
			Spec: current.NetworkSpec{
				Federation: "fedeartion-sample",
				Initiator:  "org1",
				OrderSpec:  current.IBPOrdererSpec{},
			},
		}
		client = &mocks.Client{
			GetStub: func(ctx context.Context, nn types.NamespacedName, o k8sclient.Object) error {
				switch obj := o.(type) {
				case *current.Network:
					obj.Name = network.Name
					obj.Namespace = network.Namespace
					obj.Spec = network.Spec
					obj.Labels = map[string]string{
						NETWORK_INITIATOR_LABEL: "org1",
					}
				}
				return nil
			},
		}
		k8soffering = &mockedreconcile.NetworkReconcile{}
		reconciler = &ReconcileNetwork{
			client:   client,
			Offering: k8soffering,
			Config: &operatorconfig.Config{
				Operator: operatorconfig.Operator{
					Namespace: "operator-system",
				},
			},
			update:      make(map[string][]Update),
			mutex:       &sync.Mutex{},
			rbacManager: rbac.NewRBACManager(client, nil),
		}

		request = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: network.GetNamespace(),
				Name:      network.GetName(),
			},
		}
	})

	It("reconcile break due to network not found", func() {
		client.GetStub = func(ctx context.Context, nn types.NamespacedName, o k8sclient.Object) error {
			switch o.(type) {
			case *current.Network:
				return k8serrors.NewNotFound(
					schema.GroupResource{Group: "ibp.com", Resource: "networks"},
					"network-sample",
				)
			}
			return nil
		}

		result, err := reconciler.Reconcile(context.Background(), request)
		Expect(err).To(BeNil())
		Expect(result.Requeue).To(BeFalse())

	})

	It("reconcile failed due to unexpected k8s error", func() {
		client.GetStub = func(ctx context.Context, nn types.NamespacedName, o k8sclient.Object) error {
			return k8serrors.NewTimeoutError("api server timeout", 10)
		}

		_, err := reconciler.Reconcile(context.Background(), request)
		Expect(k8serrors.IsNotFound(err)).To(BeFalse())
	})

	Context("reconcile failed due to offdering.reconcile failed", func() {
		BeforeEach(func() {
			reconciler.update = map[string][]Update{
				network.GetName(): {
					{specUpdated: true},
				},
			}

			k8soffering.ReconcileReturns(common.Result{}, errors.New("reconcile faield"))
		})
		It("set error status succ", func() {
			_, err := reconciler.Reconcile(context.TODO(), request)
			Expect(err.Error()).To(ContainSubstring("encountered error"))
		})
		It("set error status failed", func() {
			client.PatchStatusReturns(errors.New("patch error"))
			_, err := reconciler.Reconcile(context.TODO(), request)
			Expect(err.Error()).To(ContainSubstring("patch error"))
		})
	})

	Context("reconcile succ", func() {
		BeforeEach(func() {
			reconciler.update = map[string][]Update{
				network.GetName(): {
					{specUpdated: true},
				},
			}
			k8soffering.ReconcileReturns(common.Result{
				Status: &current.CRStatus{
					Type: current.Created,
				},
				Result: reconcile.Result{
					Requeue: false,
				},
			}, nil)
		})
		It("set status failed", func() {
			client.PatchStatusReturns(errors.New("patch error"))
			_, err := reconciler.Reconcile(context.TODO(), request)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("patch error"))
		})

		It("reconcile result contains requeue:true", func() {
			k8soffering.ReconcileReturns(common.Result{
				Status: &current.CRStatus{
					Type: current.Created,
				},
				Result: reconcile.Result{
					Requeue: true,
				},
			}, nil)
			result, err := reconciler.Reconcile(context.TODO(), request)
			Expect(err).To(BeNil())
			Expect(result.Requeue).To(BeTrue())
		})

		It("normal exit", func() {
			result, err := reconciler.Reconcile(context.TODO(), request)
			Expect(err).To(BeNil())
			Expect(result.Requeue).To(BeFalse())
		})
	})

})
