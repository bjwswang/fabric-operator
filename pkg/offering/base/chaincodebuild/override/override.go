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
	config "github.com/IBM-Blockchain/fabric-operator/operatorconfig"
	"github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Override struct {
	Client controllerclient.Client

	Config *config.Config
}

func (o *Override) ChaincodeBuildPipelineRun(object v1.Object, pipelineRun *pipelinev1beta1.PipelineRun, action resources.Action) error {
	instance := object.(*current.ChaincodeBuild)
	switch action {
	case resources.Create, resources.Update:
		return o.CreateChaincodeBuildPipelineRun(instance, pipelineRun)
	}

	return nil
}

func (o *Override) CreateChaincodeBuildPipelineRun(instance *current.ChaincodeBuild, pipelineRun *pipelinev1beta1.PipelineRun) error {
	pipelineRunSpec := instance.Spec.PipelineRunSpec

	pipelineRun.Name = instance.GetPipelineRunID()
	pipelineRun.Namespace = o.Config.ChaincodeBuildInitConfig.PipelinRunNamespace

	if pipelineRunSpec.GetSourceType() == current.SourceMinio {
		if pipelineRunSpec.Minio == nil {
			pipelineRunSpec.Minio = &current.Minio{}
		}
		if pipelineRunSpec.Minio.Host == "" {
			pipelineRunSpec.Minio.Host = o.Config.ChaincodeBuildInitConfig.MinioHost
		}
		if pipelineRunSpec.Minio.AccessKey == "" {
			pipelineRunSpec.Minio.AccessKey = o.Config.ChaincodeBuildInitConfig.MinioAccessKey
		}
		if pipelineRunSpec.Minio.SecretKey == "" {
			pipelineRunSpec.Minio.SecretKey = o.Config.ChaincodeBuildInitConfig.MinioSecretKey
		}
	}

	pipelineRun.Spec.Params = pipelineRunSpec.ToPipelineParams()

	pipelineRun.Spec.Workspaces = []pipelinev1beta1.WorkspaceBinding{
		{
			Name:    "source-ws",
			SubPath: "source",
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: instance.GetName() + "-source-ws",
			},
		},
		{
			Name: "dockerconfig-ws",
			Secret: &corev1.SecretVolumeSource{
				SecretName: instance.GetDockerPushSecret(),
			},
		},
	}

	return nil
}

func (o *Override) ChaincodeBuildPVC(object v1.Object, pvc *corev1.PersistentVolumeClaim, action resources.Action) error {
	instance := object.(*current.ChaincodeBuild)
	switch action {
	case resources.Create, resources.Update:
		return o.CreateChaincodeBuildPVC(instance, pvc)
	}

	return nil
}

func (o *Override) CreateChaincodeBuildPVC(instance *current.ChaincodeBuild, pvc *corev1.PersistentVolumeClaim) error {
	pvc.Name = instance.GetName() + "-source-ws"
	pvc.Namespace = o.Config.ChaincodeBuildInitConfig.PipelinRunNamespace
	return nil
}
