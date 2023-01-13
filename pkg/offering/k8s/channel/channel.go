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
	basechan "github.com/IBM-Blockchain/fabric-operator/pkg/offering/base/channel"
	baseoverride "github.com/IBM-Blockchain/fabric-operator/pkg/offering/base/channel/override"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/common"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/k8s/channel/override"
	"github.com/IBM-Blockchain/fabric-operator/pkg/operatorerrors"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ basechan.Channel = &Channel{}

type Channel struct {
	BaseChannel basechan.Channel
}

func New(client k8sclient.Client, scheme *runtime.Scheme, config *config.Config) *Channel {
	o := &override.Override{
		BaseOverride: &baseoverride.Override{},
	}
	channel := &Channel{
		BaseChannel: basechan.New(client, scheme, config, o),
	}
	return channel
}

func (channel *Channel) Reconcile(instance *current.Channel, update basechan.Update) (common.Result, error) {
	var err error

	if err = channel.PreReconcileChecks(instance, update); err != nil {
		return common.Result{}, errors.Wrap(err, "failed on prereconcile checks")
	}

	if err = channel.Initialize(instance, update); err != nil {
		return common.Result{}, operatorerrors.Wrap(err, operatorerrors.FederationInitilizationFailed, "failed to initialize channel")
	}

	if err = channel.ReconcileManagers(instance, update); err != nil {
		return common.Result{}, errors.Wrap(err, "failed to reconcile managers")
	}

	return channel.CheckStates(instance, update)
}

// TODO: customize for kubernetes

// PreReconcileChecks on Channel
func (channel *Channel) PreReconcileChecks(instance *current.Channel, update basechan.Update) error {
	return channel.BaseChannel.PreReconcileChecks(instance, update)
}

// Initialize on Channel after PreReconcileChecks
func (channel *Channel) Initialize(instance *current.Channel, update basechan.Update) error {
	return channel.BaseChannel.Initialize(instance, update)
}

// ReconcileManagers on Channel after Initialize
func (channel *Channel) ReconcileManagers(instance *current.Channel, update basechan.Update) error {
	return channel.BaseChannel.ReconcileManagers(instance, update)
}

// CheckStates on Channel after ReconcileManagers
func (channel *Channel) CheckStates(instance *current.Channel, update basechan.Update) (common.Result, error) {
	return channel.BaseChannel.CheckStates(instance, update)
}
