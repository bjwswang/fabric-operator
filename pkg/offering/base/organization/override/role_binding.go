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
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/common"
	bcrbac "github.com/IBM-Blockchain/fabric-operator/pkg/rbac"
	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (o *Override) AdminRoleBinding(object v1.Object, rb *rbacv1.RoleBinding, action resources.Action) error {
	instance := object.(*current.Organization)
	switch action {
	case resources.Create, resources.Update:
		return o.SyncAdminRoleBinding(instance, rb)
	}
	return nil
}

// Sync Admin RoleBinding based on Organization.Spec.Admin
func (o *Override) SyncAdminRoleBinding(instance *current.Organization, rb *rbacv1.RoleBinding) error {
	rb.Name = bcrbac.GetRoleBinding(instance.GetNamespaced(), bcrbac.Admin).Name
	rb.Namespace = instance.GetUserNamespace()

	// Only one `Admin`
	rb.Subjects = []rbacv1.Subject{
		common.GetDefaultSubject(instance.Spec.Admin, instance.Namespace, o.SubjectKind),
	}

	rb.RoleRef = bcrbac.RoleRef(instance.GetNamespaced(), bcrbac.Admin)

	rb.OwnerReferences = []v1.OwnerReference{bcrbac.OwnerReference(bcrbac.Organization, instance)}

	return nil
}

func (o *Override) ClientRoleBinding(object v1.Object, rb *rbacv1.RoleBinding, action resources.Action) error {
	instance := object.(*current.Organization)
	switch action {
	case resources.Create, resources.Update:
		return o.SyncClientRoleBinding(instance, rb)
	}
	return nil
}

// TODO: Sync Client RoleBinding based on Organization.Spec.Clients
func (o *Override) SyncClientRoleBinding(instance *current.Organization, rb *rbacv1.RoleBinding) error {
	rb.Name = bcrbac.GetRoleBinding(instance.GetNamespaced(), bcrbac.Client).Name
	return nil
}
