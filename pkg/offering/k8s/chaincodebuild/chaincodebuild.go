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

package k8schan

import (
	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	config "github.com/IBM-Blockchain/fabric-operator/operatorconfig"
	k8sclient "github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	baseccb "github.com/IBM-Blockchain/fabric-operator/pkg/offering/base/chaincodebuild"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/common"
	"github.com/IBM-Blockchain/fabric-operator/pkg/operatorerrors"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ baseccb.ChaincodeBuild = &ChaincodeBuild{}

type ChaincodeBuild struct {
	BaseChaincodeBuild baseccb.ChaincodeBuild
}

func New(client k8sclient.Client, scheme *runtime.Scheme, config *config.Config) *ChaincodeBuild {
	chaincodeBuild := &ChaincodeBuild{
		BaseChaincodeBuild: baseccb.New(client, scheme, config),
	}
	return chaincodeBuild
}

func (chaincodeBuild *ChaincodeBuild) Reconcile(instance *current.ChaincodeBuild, update baseccb.Update) (common.Result, error) {
	var err error

	if err = chaincodeBuild.PreReconcileChecks(instance, update); err != nil {
		return common.Result{}, errors.Wrap(err, "failed on prereconcile checks")
	}

	if err = chaincodeBuild.Initialize(instance, update); err != nil {
		return common.Result{}, operatorerrors.Wrap(err, operatorerrors.FederationInitilizationFailed, "failed to initialize chaincodeBuild")
	}

	if err = chaincodeBuild.ReconcileManagers(instance, update); err != nil {
		return common.Result{}, errors.Wrap(err, "failed to reconcile managers")
	}

	return chaincodeBuild.CheckStates(instance, update)
}

// PreReconcileChecks on ChaincodeBuild
func (chaincodeBuild *ChaincodeBuild) PreReconcileChecks(instance *current.ChaincodeBuild, update baseccb.Update) error {
	return chaincodeBuild.BaseChaincodeBuild.PreReconcileChecks(instance, update)
}

// Initialize on ChaincodeBuild after PreReconcileChecks
func (chaincodeBuild *ChaincodeBuild) Initialize(instance *current.ChaincodeBuild, update baseccb.Update) error {
	return chaincodeBuild.BaseChaincodeBuild.Initialize(instance, update)
}

// ReconcileManagers on ChaincodeBuild after Initialize
func (chaincodeBuild *ChaincodeBuild) ReconcileManagers(instance *current.ChaincodeBuild, update baseccb.Update) error {
	return chaincodeBuild.BaseChaincodeBuild.ReconcileManagers(instance, update)
}

// CheckStates on ChaincodeBuild after ReconcileManagers
func (chaincodeBuild *ChaincodeBuild) CheckStates(instance *current.ChaincodeBuild, update baseccb.Update) (common.Result, error) {
	return chaincodeBuild.BaseChaincodeBuild.CheckStates(instance, update)
}
