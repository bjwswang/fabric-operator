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
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ChaincodeBuildSpec defines the desired state of ChaincodeBuild
type ChaincodeBuildSpec struct {
	// License should be accepted by the user to be able to setup console
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	License License `json:"license"`

	// Network of the chaincode belongs to
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Network string `json:"network"`

	// Name of the chaincode
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	ID string `json:"id"`

	// Version of the chaincode
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Version string `json:"version"`

	// Initiator is the organization who initiates this chaincode build
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Initiator string `json:"initiator"`

	// PipelineRunSpec defines the tekton  pipelinerun which reference pipeline `ChaincodeBuild`
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	PipelineRunSpec PipelineRunSpec `json:"pipelineRunSpec"`
}

type PipelineRunSpec struct {
	*Git         `json:"git,omitempty"`
	*Minio       `json:"minio,omitempty"`
	*Dockerbuild `json:"dockerBuild"`
}

type SourceType string

const (
	SourceGit   SourceType = "git"
	SourceMinio SourceType = "minio"
)

type Git struct {
	URL       string `json:"url,omitempty"`
	Reference string `json:"reference,omitempty"`
}

type Minio struct {
	Host      string `json:"host,omitempty"`
	AccessKey string `json:"accessKey,omitempty"`
	SecretKey string `json:"secretKey,omitempty"`
	Bucket    string `json:"bucket,omitempty"`
	Object    string `json:"object,omitempty"`
}

type Dockerbuild struct {
	PushSecret string `json:"pushSecret,omitempty"`
	AppImage   string `json:"appImage,omitempty"`
	Dockerfile string `json:"dockerfile,omitempty"`
	Context    string `json:"context,omitempty"`
}

// ChaincodeBuildStatus defines the observed state of ChaincodeBuild
type ChaincodeBuildStatus struct {
	CRStatus `json:",inline"`
	// PipelineRunResults after pipeline completed
	PipelineRunResults []pipelinev1beta1.PipelineRunResult `json:"pipelineResults,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=ccb;ccbs
// +genclient
// +genclient:nonNamespaced
// ChaincodeBuild is the Schema for the chaincodebuilds API
type ChaincodeBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChaincodeBuildSpec   `json:"spec,omitempty"`
	Status ChaincodeBuildStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ChaincodeBuildList contains a list of ChaincodeBuild
type ChaincodeBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ChaincodeBuild `json:"items"`
}
