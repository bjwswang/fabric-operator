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

package override

import (
	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources"
	bcrbac "github.com/IBM-Blockchain/fabric-operator/pkg/rbac"
	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (o *Override) AdminClusterRole(object v1.Object, cr *rbacv1.ClusterRole, action resources.Action) error {
	instance := object.(*current.Organization)
	switch action {
	case resources.Create, resources.Update:
		return o.SyncAdminClusterRole(instance, cr)
	}

	return nil
}

func (o *Override) SyncAdminClusterRole(instance *current.Organization, cr *rbacv1.ClusterRole) error {
	namespaced := bcrbac.GetClusterRole(instance.GetNamespaced(), bcrbac.Admin)
	cr.Name = namespaced.Name
	cr.Namespace = namespaced.Namespace

	// update/patch/delete
	cr.Rules = append(cr.Rules, bcrbac.PolicyRule(bcrbac.Organization, []v1.Object{instance}, []bcrbac.Verb{bcrbac.Update, bcrbac.Patch, bcrbac.Delete}))

	cr.OwnerReferences = []v1.OwnerReference{bcrbac.OwnerReference(bcrbac.Organization, instance)}

	return nil
}

func (o *Override) ClientClusterRole(object v1.Object, cr *rbacv1.ClusterRole, action resources.Action) error {
	instance := object.(*current.Organization)
	switch action {
	case resources.Create, resources.Update:
		return o.SyncClientClusterRole(instance, cr)
	}
	return nil
}

func (o *Override) SyncClientClusterRole(instance *current.Organization, cr *rbacv1.ClusterRole) error {
	namespaced := bcrbac.GetClusterRole(instance.GetNamespaced(), bcrbac.Client)
	cr.Name = namespaced.Name
	cr.Namespace = namespaced.Namespace
	cr.OwnerReferences = []v1.OwnerReference{bcrbac.OwnerReference(bcrbac.Organization, instance)}
	return nil
}
