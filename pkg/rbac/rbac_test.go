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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("RBAC default settings", func() {
	Context("Role/RoleBinding/ClusterRole/ClusterRoleBinding default settings", func() {
		var instance types.NamespacedName
		BeforeEach(func() {
			instance = types.NamespacedName{
				Name:      "org1",
				Namespace: "org1",
			}
		})
		It("Role default setting", func() {
			role := GetRole(instance, Admin)
			Expect(role.Name).To(Equal("org1-blockchain:admin-role"))
			Expect(role.Namespace).To(Equal(instance.Namespace))
			role = GetRole(instance, Client)
			Expect(role.Name).To(Equal("org1-blockchain:client-role"))
			Expect(role.Namespace).To(Equal(instance.Namespace))
		})
		It("RoleBinding default setting", func() {
			roleBinding := GetRoleBinding(instance, Admin)
			Expect(roleBinding.Name).To(Equal("org1-blockchain:admin-rolebinding"))
			Expect(roleBinding.Namespace).To(Equal(instance.Namespace))
			roleBinding = GetRoleBinding(instance, Client)
			Expect(roleBinding.Name).To(Equal("org1-blockchain:client-rolebinding"))
			Expect(roleBinding.Namespace).To(Equal(instance.Namespace))
		})
		It("RoleRef", func() {
			rf := RoleRef(instance, Admin)
			Expect(rf.Name).To(Equal(GetRole(instance, Admin).Name))
		})
		It("ClusterRole default setting", func() {
			clusterRole := GetClusterRole(instance, Admin)
			Expect(clusterRole.Name).To(Equal("org1-blockchain:admin-clusterrole"))
			clusterRole = GetClusterRole(instance, Client)
			Expect(clusterRole.Name).To(Equal("org1-blockchain:client-clusterrole"))
		})
		It("ClusterRoleBinding default setting", func() {
			clusterRoleBinding := GetClusterRoleBinding(instance, Admin)
			Expect(clusterRoleBinding.Name).To(Equal("org1-blockchain:admin-clusterrolebinding"))
			clusterRoleBinding = GetClusterRoleBinding(instance, Client)
			Expect(clusterRoleBinding.Name).To(Equal("org1-blockchain:client-clusterrolebinding"))
		})
		It("ClusterRoleRef", func() {
			crf := ClusterRoleRef(instance, Admin)
			Expect(crf.Name).To(Equal(GetClusterRole(instance, Admin).Name))
		})
	})

	Context("PolicyRule checks", func() {
		var rules []rbacv1.PolicyRule
		var rule rbacv1.PolicyRule
		var instance v1.Object
		var instance2 v1.Object
		BeforeEach(func() {
			instance = &current.Organization{
				ObjectMeta: v1.ObjectMeta{
					Name: "org1",
				},
			}
			instance2 = &current.Federation{
				ObjectMeta: v1.ObjectMeta{
					Name: "federation-sample",
				},
			}
			rules = []rbacv1.PolicyRule{PolicyRule(Federation, []v1.Object{instance2}, []Verb{Get})}
			rule = PolicyRule(Organization, []v1.Object{instance}, []Verb{Get})
		})
		It("policy rule string", func() {
			rule := PolicyRule(Organization, []v1.Object{instance}, []Verb{Get})
			Expect(PolicyRuleString(rule)).To(Equal("organizations.org1"))
		})

		It("check policy rule", func() {
			index, found := CheckPolicyRule(rules, rule)
			Expect(found).To(BeFalse())
			Expect(index).To(Equal(0))

			rules = append(rules, rule)
			index, found = CheckPolicyRule(rules, rule)
			Expect(found).To(BeTrue())
			Expect(index).To(Equal(1))
		})
	})
})
