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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkSpec defines the desired state of Network
type NetworkSpec struct {
	// License should be accepted by the user to be able to setup console
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	License License `json:"license"`

	// Federation which this network belongs to
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Federation string `json:"federation"`

	// Initiator is the organization who initiates this network and host consensus cluster
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Initiator string `json:"initiator"`
	// InitialToken is the default value of the OrderSpec.ClusterSecret.[].Enrollment.TLS/Component.EnrollToken
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	InitialToken string `json:"initialToken"`

	// OrderSpec is the configurations of network's related Order
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	OrderSpec IBPOrdererSpec `json:"orderSpec,omitempty"`
}

// NetworkStatus defines the observed state of Network
type NetworkStatus struct {
	CRStatus `json:",inline"`
	// Channels in this network
	Channels []string `json:"channels,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// Network is the Schema for the networks API
type Network struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkSpec   `json:"spec,omitempty"`
	Status NetworkStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NetworkList contains a list of Network
type NetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Network `json:"items"`
}
