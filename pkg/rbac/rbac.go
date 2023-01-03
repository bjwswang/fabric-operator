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
	"fmt"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type RoleType string

const (
	Admin  RoleType = "admin"
	Client RoleType = "client"
)

func (rt RoleType) String() string {
	return string(rt)
}

func Roles() []string {
	return []string{Admin.String(), Client.String()}
}

const (
	// AdminSuffix used to keep same format with pkg/manager's rolo/clusterrole reconcile
	AdminSuffix = "blockchain:admin"
	// ClientSuffix used to keep same format with pkg/manager's rolo/clusterrole reconcile
	ClientSuffix = "blockchain:client"
)

// Role/RoleBinding on Namespaced scope resources

// GetRole returns namespaced Role info by instance(organziation) and role type
func GetRole(instance types.NamespacedName, role RoleType) types.NamespacedName {
	if role == Admin {
		return types.NamespacedName{Name: fmt.Sprintf("%s-%s-role", instance.Name, AdminSuffix), Namespace: instance.Namespace}
	}
	return types.NamespacedName{Name: fmt.Sprintf("%s-%s-role", instance.Name, ClientSuffix), Namespace: instance.Namespace}
}

// GetRoleBinding returns namespaced RoleBinding info by instance(organziation) and role type
func GetRoleBinding(instance types.NamespacedName, role RoleType) types.NamespacedName {
	if role == Admin {
		return types.NamespacedName{Name: fmt.Sprintf("%s-%s-rolebinding", instance.Name, AdminSuffix), Namespace: instance.Namespace}
	}
	return types.NamespacedName{Name: fmt.Sprintf("%s-%s-rolebinding", instance.Name, ClientSuffix), Namespace: instance.Namespace}
}

// RoleRef build a rbacv1.RoleRef by instance(organziation) and role type
func RoleRef(instance types.NamespacedName, role RoleType) rbacv1.RoleRef {
	return rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "Role",
		Name:     GetRole(instance, role).Name,
	}
}

// ClusterRole/ClusterRoleBinding on Cluster scope resources

// GetClusterRole returns ClusterRole info by instance(organziation) and role type
func GetClusterRole(instance types.NamespacedName, role RoleType) types.NamespacedName {
	if role == Admin {
		return types.NamespacedName{Name: fmt.Sprintf("%s-%s-clusterrole", instance.Name, AdminSuffix)}
	}
	return types.NamespacedName{Name: fmt.Sprintf("%s-%s-clusterrole", instance.Name, ClientSuffix)}
}

// GetClusterRoleBinding returns ClusterRoleBinding info by instance(organziation) and role type
func GetClusterRoleBinding(instance types.NamespacedName, role RoleType) types.NamespacedName {
	if role == Admin {
		return types.NamespacedName{Name: fmt.Sprintf("%s-%s-clusterrolebinding", instance.Name, AdminSuffix)}
	}
	return types.NamespacedName{Name: fmt.Sprintf("%s-%s-clusterrolebinding", instance.Name, ClientSuffix)}
}

// ClusterRoleRef build a rbacv1.RoleRef by instance(organziation) and role type
func ClusterRoleRef(instance types.NamespacedName, role RoleType) rbacv1.RoleRef {
	return rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     GetClusterRole(instance, role).Name,
	}
}

// OwnerReference in Roles/RoleBindings/ClusterRoles/ClusterRoleBindings
func OwnerReference(kind Resource, owner v1.Object) v1.OwnerReference {
	return v1.OwnerReference{
		Kind:       string(kind),
		APIVersion: GroupVersion.String(),
		Name:       owner.GetName(),
		UID:        owner.GetUID(),
	}
}

type Verb string

const (
	Get    Verb = "get"
	List   Verb = "list"
	Create Verb = "create"
	Update Verb = "update"
	Patch  Verb = "patch"
	Watch  Verb = "watch"
	Delete Verb = "delete"
)

func (verb Verb) String() string {
	return string(verb)
}

// PolicyRule initialize a rbacv1.PolicyRule with specific resource objects which will be configured at `resourceNames`
func PolicyRule(kind Resource, resources []v1.Object, verbs []Verb) rbacv1.PolicyRule {
	resourceNames := make([]string, len(resources))
	for index, resource := range resources {
		resourceNames[index] = resource.GetName()
	}

	verbsStr := make([]string, len(verbs))
	for index, verb := range verbs {
		verbsStr[index] = verb.String()
	}
	return rbacv1.PolicyRule{
		APIGroups:     []string{GroupVersion.Group},
		Resources:     []string{kind.String()},
		ResourceNames: resourceNames,
		Verbs:         verbsStr,
	}
}

// CheckPolicyRule checks/returns the rule's specific index in rules. False will be returns if rule not exist in rules
func CheckPolicyRule(rules []rbacv1.PolicyRule, rule rbacv1.PolicyRule) (int, bool) {
	des := PolicyRuleString(rule)
	for index, pr := range rules {
		if PolicyRuleString(pr) == des {
			return index, true
		}
	}
	return 0, false
}

// PolicyRuleString create a joined string with resources and resourceNames to make sure its uniqueness
func PolicyRuleString(rule rbacv1.PolicyRule) string {
	return strings.Join(append(rule.Resources, rule.ResourceNames...), ".")
}
