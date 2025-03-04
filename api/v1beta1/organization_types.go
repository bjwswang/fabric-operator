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

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OrganizationSpec defines the desired state of Organization
// +k8s:deepcopy-gen=true
// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
type OrganizationSpec struct {
	// License should be accepted by the user to be able to setup console
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	License License `json:"license"`

	// DisplayName for this organization
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	DisplayName string `json:"displayName,omitempty"`

	// Description
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Description string `json:"description,omitempty"`

	// Admin is the User/ServiceAccount with `Admin` role both in kubernetes and in CA
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Admin      string `json:"admin"`
	AdminToken string `json:"admintoken,omitempty"`

	// Clients are the Users/ServiceAccounts with `Client` role both in kubernetes and in CA
	Clients []string `json:"clients,omitempty"`

	// CASpec is the configurations of organization's related Certificate Authority
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	CASpec IBPCASpec `json:"caSpec,omitempty"`
}

// OrganizationStatus defines the observed state of Organization
type OrganizationStatus struct {
	// CRStatus is the custome resource status
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	CRStatus `json:",inline"`

	// Federations which this organization has been added
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Federations []string `json:"federations,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=org;orgs
// +genclient
// +genclient:nonNamespaced
// Organization is the Schema for the organizations API
type Organization struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrganizationSpec   `json:"spec,omitempty"`
	Status OrganizationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// OrganizationList contains a list of Organization
type OrganizationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Organization `json:"items"`
}
