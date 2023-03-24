package v1beta1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:deepcopy-gen=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=epolicy;epolicies
// +genclient
// +genclient:nonNamespaced
type EndorsePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Spec EndorsePolicySpec `json:"spec,omitempty"`
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	Status EndorsePolicyStatus `json:"status,omitempty"`
}

type EndorsePolicySpec struct {
	Channel     string `json:"channel"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
	DisplayName string `json:"displayName"`
}

type EndorsePolicyStatus struct {
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime,omitempty"`
}

// EndorsePolicyList contains a list of EndorsePolicy.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:deepcopy-gen=true
type EndorsePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []EndorsePolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EndorsePolicy{}, &EndorsePolicyList{})
}
