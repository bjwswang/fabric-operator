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

package pipelinerun

import (
	"context"
	"fmt"

	k8sclient "github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources"
	"github.com/IBM-Blockchain/fabric-operator/pkg/operatorerrors"
	"github.com/IBM-Blockchain/fabric-operator/pkg/util"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("pipelinerun_manager")

type Manager struct {
	Client         k8sclient.Client
	Scheme         *runtime.Scheme
	Name           string
	PipelinRunFile string
	LabelsFunc     func(v1.Object) map[string]string
	OverrideFunc   func(v1.Object, *pipelinev1beta1.PipelineRun, resources.Action) error
}

func (m *Manager) GetName(instance v1.Object) string {
	return GetName(instance.GetName(), m.Name)
}

func (m *Manager) Reconcile(instance v1.Object, update bool) error {
	var err error

	pipelinRun, err := m.GetPipelineRunBasedOnCRFromFile(instance)
	if err != nil {
		return err
	}

	currrentPipelineRun := &pipelinev1beta1.PipelineRun{}
	err = m.Client.Get(context.TODO(), types.NamespacedName{Name: pipelinRun.Name, Namespace: pipelinRun.Namespace}, currrentPipelineRun)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info(fmt.Sprintf("Creating pipelinerun '%s'", pipelinRun.Name))
			err = m.Client.Create(context.TODO(), pipelinRun, k8sclient.CreateOption{Owner: instance, Scheme: m.Scheme})
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}

	if update {
		if m.OverrideFunc != nil {
			err := m.OverrideFunc(instance, currrentPipelineRun, resources.Update)
			if err != nil {
				return operatorerrors.New(operatorerrors.InvalidClusterRoleUpdateRequest, err.Error())
			}
		}
		if err = m.Client.Update(context.TODO(), currrentPipelineRun); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) GetPipelineRunBasedOnCRFromFile(instance v1.Object) (*pipelinev1beta1.PipelineRun, error) {
	pipelineRun, err := util.GetPipelineRunFromFile(m.PipelinRunFile)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error reading pipelinerun configuration file: %s", m.PipelinRunFile))
		return nil, err
	}

	pipelineRun.Name = m.GetName(instance)
	pipelineRun.Namespace = instance.GetNamespace()
	pipelineRun.Labels = m.LabelsFunc(instance)

	return m.BasedOnCR(instance, pipelineRun)
}

func (m *Manager) BasedOnCR(instance v1.Object, pipelineRun *pipelinev1beta1.PipelineRun) (*pipelinev1beta1.PipelineRun, error) {
	if m.OverrideFunc != nil {
		err := m.OverrideFunc(instance, pipelineRun, resources.Create)
		if err != nil {
			return nil, operatorerrors.New(operatorerrors.InvalidClusterRoleCreateRequest, err.Error())
		}
	}

	return pipelineRun, nil
}

func GetName(instanceName string, suffix ...string) string {
	if len(suffix) != 0 {
		if suffix[0] != "" {
			return fmt.Sprintf("%s-%s-pipelinerun", instanceName, suffix[0])
		}
	}
	return fmt.Sprintf("%s-pipelinerun", instanceName)
}

func (m *Manager) Get(instance v1.Object) (client.Object, error) {
	// NO-OP
	return nil, nil
}

func (m *Manager) Exists(instance v1.Object) bool {
	// NO-OP
	return true
}

func (m *Manager) Delete(instance v1.Object) error {
	// NO-OP
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
