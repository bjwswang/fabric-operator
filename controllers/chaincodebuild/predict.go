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
	"context"
	"fmt"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	k8sclient "github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/go-test/deep"
	"github.com/pkg/errors"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func (r *ReconcileChaincodeBuild) CreateFunc(e event.CreateEvent) bool {
	chaincodeBuild := e.Object.(*current.ChaincodeBuild)
	log.Info(fmt.Sprintf("Create event detected for ChaincodeBuild '%s'", chaincodeBuild.GetName()))

	update := Update{}

	if chaincodeBuild.HasType() {
		log.Info(fmt.Sprintf("Operator restart detected, running update flow on existing ChaincodeBuild '%s'", chaincodeBuild.GetName()))

		// Get the spec state of the resource before the operator went down, this
		// will be used to compare to see if the spec of resources has changed
		cm, err := r.GetSpecState(chaincodeBuild)
		if err != nil {
			log.Info(fmt.Sprintf("Failed getting saved ChaincodeBuild spec '%s', triggering create: %s", chaincodeBuild.GetName(), err.Error()))
			return true
		}

		specBytes := cm.BinaryData["spec"]
		existingchaincodeBuild := &current.ChaincodeBuild{}
		err = yaml.Unmarshal(specBytes, &existingchaincodeBuild.Spec)
		if err != nil {
			log.Info(fmt.Sprintf("Unmarshal failed for saved ChaincodeBuild spec '%s', triggering create: %s", chaincodeBuild.GetName(), err.Error()))
			return true
		}

		diff := deep.Equal(chaincodeBuild.Spec, existingchaincodeBuild.Spec)
		if diff != nil {
			log.Info(fmt.Sprintf("ChaincodeBuild '%s' spec was updated while operator was down", chaincodeBuild.GetName()))
			log.Info(fmt.Sprintf("Difference detected: %v", diff))
			update.specUpdated = true
			update.pipelineSpecUpdated = true
		}

		log.Info(fmt.Sprintf("Create event triggering reconcile for updating ChaincodeBuild '%s'", chaincodeBuild.GetName()))
		r.PushUpdate(chaincodeBuild.GetName(), update)
		return true
	}

	update.specUpdated = true
	update.pipelineSpecUpdated = true

	r.PushUpdate(chaincodeBuild.GetName(), update)

	return true
}

func (r *ReconcileChaincodeBuild) PipelineRunUpdateFunc(e event.UpdateEvent) bool {
	// oldPipelineRun := e.ObjectOld.(*pipelinev1beta1.PipelineRun)
	newPipelineRun := e.ObjectNew.(*pipelinev1beta1.PipelineRun)
	log.Info(fmt.Sprintf("Update event detected for Pipelinerun '%s'", newPipelineRun.GetName()))

	build, err := r.getChaincodeBuildFromPipelineRun(newPipelineRun)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Error(err, "ChaincodeBuild not found")
			return false
		}
		log.Error(err, "Failed to get ChaincodeBuild")
		return false
	}
	err = r.patchChaincodeBuildStatus(build, newPipelineRun)
	if err != nil {
		log.Error(err, "Patch ChaincodeBuild status based on PipelineRun update", build.Name, newPipelineRun.Name)
	}

	return false
}

func (r *ReconcileChaincodeBuild) getChaincodeBuildFromPipelineRun(pipelineRun *pipelinev1beta1.PipelineRun) (*current.ChaincodeBuild, error) {
	for _, owner := range pipelineRun.OwnerReferences {
		if owner.Kind != KIND {
			continue
		}
		build := &current.ChaincodeBuild{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: owner.Name}, build)
		if err != nil {
			return nil, err
		}
		return build, nil
	}
	return nil, errors.New("Skip since this pipelienrun not initiated by a ChaincodeBuild")
}

func (r *ReconcileChaincodeBuild) patchChaincodeBuildStatus(build *current.ChaincodeBuild, pipelineRun *pipelinev1beta1.PipelineRun) error {
	newStatus := pipelineRun.Status
	for _, condition := range newStatus.Conditions {
		// Pipeline completed
		if condition.Reason == pipelinev1beta1.PipelineRunReasonCompleted.String() && condition.Status == corev1.ConditionTrue {
			// retrieve and patch pipelinerun results
			if newStatus.PipelineResults != nil {
				build.Status.PipelineRunResults = newStatus.PipelineResults
				err := r.client.PatchStatus(context.TODO(), build, nil, k8sclient.PatchOption{
					Resilient: &k8sclient.ResilientPatch{
						Retry:    3,
						Into:     &current.ChaincodeBuild{},
						Strategy: client.MergeFrom,
					},
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
