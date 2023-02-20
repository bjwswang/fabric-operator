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

// ChannelSpec defines the desired state of Channel
type ChannelSpec struct {
	// License should be accepted by the user to be able to setup console
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	License License `json:"license"`

	// Network which this channel belongs to
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Network string `json:"network"`

	// Members list all organization in this Channel
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Members []Member `json:"members"`

	// Peers list all fabric peers joined at this channel
	Peers []NamespacedName `json:"peers,omitempty"`

	// Description for this Channel
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Description string `json:"description,omitempty"`
}

type PeerConditionType string

const (
	PeerJoined PeerConditionType = "PeerJoined"
	PeerError  PeerConditionType = "PeerError"
)

// ChannelPeer is the IBPPeer which joins this channel
type PeerCondition struct {
	NamespacedName `json:",inline"`
	// Type is the type of the condition.
	Type PeerConditionType `json:"type"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status metav1.ConditionStatus `json:"status"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// ChannelStatus defines the observed state of Channel
type ChannelStatus struct {
	CRStatus       `json:",inline"`
	ArchivedStatus CRStatus        `json:"archivedStatus,omitempty"`
	PeerConditions []PeerCondition `json:"peerConditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=chan;chans

// Channel is the Schema for the channels API
type Channel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChannelSpec   `json:"spec,omitempty"`
	Status ChannelStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ChannelList contains a list of Channel
type ChannelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Channel `json:"items"`
}
