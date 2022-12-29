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

func (o *Override) AdminRole(object v1.Object, role *rbacv1.Role, action resources.Action) error {
	instance := object.(*current.Organization)
	switch action {
	case resources.Create, resources.Update:
		return o.SyncAdminRole(instance, role)
	}

	return nil
}

func (o *Override) SyncAdminRole(instance *current.Organization, role *rbacv1.Role) error {
	role.Name = bcrbac.GetRole(instance.GetNamespaced(), bcrbac.Admin).Name
	role.Namespace = instance.GetUserNamespace()
	role.OwnerReferences = []v1.OwnerReference{bcrbac.OwnerReference(bcrbac.Organization, instance)}
	return nil
}

func (o *Override) ClientRole(object v1.Object, role *rbacv1.Role, action resources.Action) error {
	instance := object.(*current.Organization)
	switch action {
	case resources.Create, resources.Update:
		return o.SyncClientRole(instance, role)
	}
	return nil
}

func (o *Override) SyncClientRole(instance *current.Organization, role *rbacv1.Role) error {
	role.Name = bcrbac.GetRole(instance.GetNamespaced(), bcrbac.Client).Name
	role.Namespace = instance.GetUserNamespace()

	role.OwnerReferences = []v1.OwnerReference{bcrbac.OwnerReference(bcrbac.Organization, instance)}
	return nil
}
