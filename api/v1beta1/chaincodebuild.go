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

package v1beta1

import (
	"errors"
	"fmt"

	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

var (
	errInvalidSourceMinio = errors.New("pipeline source minio is invalid.missing bucket or object")
	errInvalidSourceGit   = errors.New("pipeline source git is invalid.missing url")
)

const (
	IMAGE_URL    = "IMAGE_URL"
	IMAGE_DIGEST = "IMAGE_DIGEST"
)

func init() {
	SchemeBuilder.Register(&ChaincodeBuild{}, &ChaincodeBuildList{})
}

func (build *ChaincodeBuild) HasType() bool {
	return build.Status.CRStatus.Type != ""
}

func (buildSpec *ChaincodeBuildSpec) HasPipelineSource() bool {
	return buildSpec.PipelineRunSpec.Git != nil || buildSpec.PipelineRunSpec.Minio != nil
}

func (buildSpec *ChaincodeBuildSpec) ValidatePipelineSource() error {
	if buildSpec.PipelineRunSpec.Minio != nil {
		if buildSpec.PipelineRunSpec.Minio.Bucket == "" || buildSpec.PipelineRunSpec.Minio.Object == "" {
			return errInvalidSourceMinio
		}
	}
	if buildSpec.PipelineRunSpec.Git != nil {
		if buildSpec.PipelineRunSpec.Git.URL == "" {
			return errInvalidSourceGit
		}
	}
	return nil
}

func (build *ChaincodeBuild) GetPipelineRunID() string {
	return build.GetName() + "-pipelinerun"
}

func (build *ChaincodeBuild) GetDockerPushSecret() string {
	return build.Spec.PipelineRunSpec.PushSecret
}

func (p PipelineRunSpec) GetSourceType() SourceType {
	if p.Minio != nil {
		return SourceMinio
	}
	if p.Git != nil {
		return SourceGit
	}
	return ""
}

func (p PipelineRunSpec) ToPipelineParams() []pipelinev1beta1.Param {
	params := make([]pipelinev1beta1.Param, 0)
	params = append(params, pipelinev1beta1.Param{
		Name:  "SOURCE",
		Value: *pipelinev1beta1.NewArrayOrString(string(p.GetSourceType())),
	})
	if p.Git != nil {
		gitParams := p.Git.ToPipelineParams()
		params = append(params, gitParams...)
	}
	if p.Minio != nil {
		minioParams := p.Minio.ToPipelineParams()
		params = append(params, minioParams...)
	}
	if p.Dockerbuild != nil {
		dockerParams := p.Dockerbuild.ToPipelineParams()
		params = append(params, dockerParams...)
	}

	return params
}

func (p Git) ToPipelineParams() []pipelinev1beta1.Param {
	return []pipelinev1beta1.Param{
		{
			Name:  "SOURCE_GIT_URL",
			Value: *pipelinev1beta1.NewArrayOrString(p.URL),
		},
		{
			Name:  "SOURCE_GIT_REFERENCE",
			Value: *pipelinev1beta1.NewArrayOrString(p.Reference),
		},
	}
}

func (p Minio) ToPipelineParams() []pipelinev1beta1.Param {
	return []pipelinev1beta1.Param{
		{
			Name:  "SOURCE_MINIO_HOST",
			Value: *pipelinev1beta1.NewArrayOrString(p.Host),
		},
		{
			Name:  "SOURCE_MINIO_ACCESS_KEY",
			Value: *pipelinev1beta1.NewArrayOrString(p.AccessKey),
		},
		{
			Name:  "SOURCE_MINIO_SECRET_KEY",
			Value: *pipelinev1beta1.NewArrayOrString(p.SecretKey),
		},
		{
			Name:  "SOURCE_MINIO_BUCKET",
			Value: *pipelinev1beta1.NewArrayOrString(p.Bucket),
		},
		{
			Name:  "SOURCE_MINIO_OBJECT",
			Value: *pipelinev1beta1.NewArrayOrString(p.Object),
		},
	}
}

func (p Dockerbuild) ToPipelineParams() []pipelinev1beta1.Param {
	return []pipelinev1beta1.Param{
		{
			Name:  "APP_IMAGE",
			Value: *pipelinev1beta1.NewArrayOrString(p.AppImage),
		},
		{
			Name:  "DOCKERFILE",
			Value: *pipelinev1beta1.NewArrayOrString(p.Dockerfile),
		},
		{
			Name:  "CONTEXT",
			Value: *pipelinev1beta1.NewArrayOrString(p.Context),
		},
	}
}

func (ccb *ChaincodeBuild) HasImage() error {
	url, digest := "", ""
	for _, item := range ccb.Status.PipelineRunResults {
		if item.Name == IMAGE_URL {
			url = item.Value
		}
		if item.Name == IMAGE_DIGEST {
			digest = item.Value
		}
	}
	if url == "" || digest == "" {
		return fmt.Errorf("chaincodebuild %s's image don't exist image: '%s', digest: '%s'", ccb.GetName(), url, digest)
	}
	return nil
}
