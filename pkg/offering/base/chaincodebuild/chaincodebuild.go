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

package chaincodebuild

import (
	"fmt"

	"github.com/pkg/errors"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	config "github.com/IBM-Blockchain/fabric-operator/operatorconfig"
	"github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources"
	resourcemanager "github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources/manager"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/base/chaincodebuild/override"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/common"
	"github.com/IBM-Blockchain/fabric-operator/version"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("base_chaincodeBuild")

//go:generate counterfeiter -o mocks/update.go -fake-name Update . Update

type Update interface {
	SpecUpdated() bool
	PipelineSpecUpdated() bool
}

//go:generate counterfeiter -o mocks/override.go -fake-name Override . Override

type Override interface {
	ChaincodeBuildPipelineRun(object v1.Object, pipelineRun *pipelinev1beta1.PipelineRun, action resources.Action) error
	ChaincodeBuildPVC(object v1.Object, pvc *corev1.PersistentVolumeClaim, action resources.Action) error
}

//go:generate counterfeiter -o mocks/basechaincodebuild.go -fake-name ChaincodeBuild . ChaincodeBuild

type ChaincodeBuild interface {
	PreReconcileChecks(instance *current.ChaincodeBuild, update Update) error
	Initialize(instance *current.ChaincodeBuild, update Update) error
	ReconcileManagers(instance *current.ChaincodeBuild, update Update) error
	CheckStates(instance *current.ChaincodeBuild, update Update) (common.Result, error)
}

var _ ChaincodeBuild = (*BaseChaincodeBuild)(nil)

const (
	KIND = "ChaincodeBuild"
)

type BaseChaincodeBuild struct {
	Client controllerclient.Client
	Scheme *runtime.Scheme

	PVCManager      resources.Manager
	PipelineManager resources.Manager

	Config *config.Config

	Override Override
}

func New(client controllerclient.Client, scheme *runtime.Scheme, config *config.Config) *BaseChaincodeBuild {
	o := &override.Override{
		Client: client,
		Config: config,
	}
	base := &BaseChaincodeBuild{
		Client:   client,
		Scheme:   scheme,
		Config:   config,
		Override: o,
	}

	base.CreateManagers()

	return base
}

func (chaincodeBuild *BaseChaincodeBuild) CreateManagers() {
	resourceManager := resourcemanager.New(chaincodeBuild.Client, chaincodeBuild.Scheme)
	chaincodeBuild.PVCManager = resourceManager.CreatePVCManager("", chaincodeBuild.Override.ChaincodeBuildPVC, chaincodeBuild.GetLabels, chaincodeBuild.Config.ChaincodeBuildInitConfig.PipelineRunPVCFile)
	chaincodeBuild.PipelineManager = resourceManager.CreatePipelinerunManager("", chaincodeBuild.Override.ChaincodeBuildPipelineRun, chaincodeBuild.GetLabels, chaincodeBuild.Config.ChaincodeBuildInitConfig.PipelineRunFile)
}

// PreReconcileChecks on ChaincodeBuild upon Update
func (chaincodeBuild *BaseChaincodeBuild) PreReconcileChecks(instance *current.ChaincodeBuild, update Update) error {
	log.Info(fmt.Sprintf("PreReconcileChecks on ChaincodeBuild %s", instance.GetName()))

	if !instance.Spec.HasPipelineSource() {
		return errors.New("invalid chaincode build.must provide a pipeline source(Git/Minio)")
	}

	return nil
}

// Initialize on ChaincodeBuild upon Update
func (baseChaincodeBuild *BaseChaincodeBuild) Initialize(instance *current.ChaincodeBuild, update Update) error {
	return nil
}

// ReconcileManagers on ChaincodeBuild upon Update
func (baseChaincodeBuild *BaseChaincodeBuild) ReconcileManagers(instance *current.ChaincodeBuild, update Update) error {
	if update.PipelineSpecUpdated() {
		if err := baseChaincodeBuild.PVCManager.Reconcile(instance, false); err != nil {
			return err
		}

		if err := baseChaincodeBuild.PipelineManager.Reconcile(instance, false); err != nil {
			return err
		}
	}
	return nil
}

// CheckStates on ChaincodeBuild
func (baseChaincodeBuild *BaseChaincodeBuild) CheckStates(instance *current.ChaincodeBuild, update Update) (common.Result, error) {
	if !instance.HasType() {
		return common.Result{
			Status: &current.CRStatus{
				Type:    current.Created,
				Version: version.Operator,
			},
		}, nil
	}

	return common.Result{}, nil
}

// GetLabels from instance.GetLabels
func (baseChaincodeBuild *BaseChaincodeBuild) GetLabels(instance v1.Object) map[string]string {
	return instance.GetLabels()
}
