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

package clusterrolebinding

import (
	"context"
	"fmt"

	k8sclient "github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources"
	"github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources/clusterrole"
	"github.com/IBM-Blockchain/fabric-operator/pkg/operatorerrors"
	"github.com/IBM-Blockchain/fabric-operator/pkg/util"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("clusterrolebinding_manager")

type Manager struct {
	Client                 k8sclient.Client
	Scheme                 *runtime.Scheme
	ClusterRoleBindingFile string
	Name                   string

	LabelsFunc   func(v1.Object) map[string]string
	OverrideFunc func(v1.Object, *rbacv1.ClusterRoleBinding, resources.Action) error
}

type SubjectKind string

const (
	ServiceAccount SubjectKind = "ServiceAccount"
	User           SubjectKind = "User"
)

func (m *Manager) GetName(instance v1.Object) string {
	if m.Name != "" {
		return fmt.Sprintf("%s-%s-clusterrolebinding", instance.GetName(), m.Name)
	}
	return fmt.Sprintf("%s-clusterrolebinding", instance.GetName())
}

// Reconcle handles:
//   - create on clusterRoleBinding
//   - update on currentClusterRoleBinding
func (m *Manager) Reconcile(instance v1.Object, update bool) error {
	clusterRoleBinding, err := m.GetClusterRoleBindingBasedOnCRFromFile(instance)
	if err != nil {
		return err
	}
	currentClusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	err = m.Client.Get(context.TODO(), types.NamespacedName{Name: clusterRoleBinding.Name, Namespace: clusterRoleBinding.Namespace}, currentClusterRoleBinding)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info(fmt.Sprintf("Creating cluster role binding '%s'", clusterRoleBinding.Name))

			err = m.Client.Create(context.TODO(), clusterRoleBinding, k8sclient.CreateOption{Owner: instance, Scheme: m.Scheme})
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}

	// update if exists(same like Create)
	if update {
		if m.OverrideFunc != nil {
			err := m.OverrideFunc(instance, currentClusterRoleBinding, resources.Update)
			if err != nil {
				return operatorerrors.New(operatorerrors.InvalidClusterRoleBindingUpdateRequest, err.Error())
			}
		}
		if err = m.Client.Update(context.TODO(), currentClusterRoleBinding); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) GetClusterRoleBindingBasedOnCRFromFile(instance v1.Object) (*rbacv1.ClusterRoleBinding, error) {
	clusterRoleBinding, err := util.GetClusterRoleBindingFromFile(m.ClusterRoleBindingFile)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error reading cluster role binding configuration file: %s", m.ClusterRoleBindingFile))
		return nil, err
	}

	name := m.GetName(instance)
	clusterRoleBinding.Name = name
	clusterRoleBinding.Namespace = instance.GetNamespace()
	clusterRoleBinding.Labels = m.LabelsFunc(instance)

	clusterRoleBinding.RoleRef = rbacv1.RoleRef{
		Name:     clusterrole.GetName(instance.GetName()),
		Kind:     "ClusterRole",
		APIGroup: "",
	}

	return m.BasedOnCR(instance, clusterRoleBinding)
}

func (m *Manager) BasedOnCR(instance v1.Object, clusterRoleBinding *rbacv1.ClusterRoleBinding) (*rbacv1.ClusterRoleBinding, error) {
	if m.OverrideFunc != nil {
		err := m.OverrideFunc(instance, clusterRoleBinding, resources.Create)
		if err != nil {
			return nil, operatorerrors.New(operatorerrors.InvalidClusterRoleBindingCreateRequest, err.Error())
		}
	}

	return clusterRoleBinding, nil
}

func (m *Manager) Get(instance v1.Object) (client.Object, error) {
	if instance == nil {
		return nil, nil // Instance has not been reconciled yet
	}

	name := m.GetName(instance)
	crb := &rbacv1.ClusterRoleBinding{}
	err := m.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: instance.GetNamespace()}, crb)
	if err != nil {
		return nil, err
	}

	return crb, nil
}

func (m *Manager) Exists(instance v1.Object) bool {
	_, err := m.Get(instance)

	return err == nil
}

func (m *Manager) Delete(instance v1.Object) error {
	crb, err := m.Get(instance)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	if crb == nil {
		return nil
	}

	err = m.Client.Delete(context.TODO(), crb)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func (m *Manager) CheckState(instance v1.Object) error {
	// NO-OP
	return nil
}

func (m *Manager) RestoreState(instance v1.Object) error {
	// NO-OP
	return nil
}

func (m *Manager) SetCustomName(name string) {
	// NO-OP
}
