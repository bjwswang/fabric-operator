package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ChaincodePhase string

const (
	// ChaincodePhasePending create chaincode default state
	ChaincodePhasePending ChaincodePhase = "ChaincodePending"
	// ChaincodePhaseApproved the proposal is approved and the status of the chaincode
	ChaincodePhaseApproved ChaincodePhase = "ChaincodeApproved"
	// ChaincodePhaseUnapproved the proposal fails and the status of the chaincode
	ChaincodePhaseUnapproved ChaincodePhase = "ChaincodeUnapproved"

	// ChaincodePhaseRunning amazing
	ChaincodePhaseRunning ChaincodePhase = "ChaincodeRunning"
)

type ChaincodeConditionType string

const (
	// ChaincodeCondPackaged chaincode packaged successfully
	ChaincodeCondPackaged ChaincodeConditionType = "Packaged"
	// ChaincodeCondInstalled chaincode code installed successfully
	ChaincodeCondInstalled ChaincodeConditionType = "Installed"
	// ChaincodeCondApproved chaincode definition approved
	ChaincodeCondApproved ChaincodeConditionType = "Approved"
	// ChaincodeCondCommitted chaincode commits to the chain
	ChaincodeCondCommitted ChaincodeConditionType = "Committed"
	// ChaincodeCondRunning chaincode is running, pod is in running state
	ChaincodeCondRunning ChaincodeConditionType = "Running"
	// ChaincodeCondError problems with these processes: package, install, approve, commit
	ChaincodeCondError ChaincodeConditionType = "Error"
	// ChaincodeCondDone process execution is complete and chaincode is running correctly
	ChaincodeCondDone ChaincodeConditionType = "Done"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:deepcopy-gen=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=cc
type Chaincode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	Spec ChaincodeSpec `json:"spec"`
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	Status ChaincodeStatus `json:"status,omitempty"`
}

// ChaincodeList contains a list of Chaincode.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:deepcopy-gen=true
type ChaincodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Chaincode `json:"items"`
}

type ChaincodeSpec struct {
	License License `json:"license"`

	// Which channel does chaincode belong to.
	Channel string `json:"channel"`
	// chaincode id
	ID string `json:"id,omitempty"`
	// current version
	Version string `json:"version,omitempty"`
	// +kubebuilder:validation:Pattern:=`^[[:alnum:]][[:alnum:]_.+-]*$`
	Label        string `json:"label,omitempty"`
	InitRequired bool   `json:"initRequired"`

	EndorsePolicyRef `json:"endorsePolicyRef"`
	// ExternalBuilder used, default is k8s
	// +optional
	ExternalBuilder string `json:"externalBuilder,omitempty"`
	// the image used by the current version of chaincode
	Images ChaincodeImage `json:"images,omitempty"`
}

type EndorsePolicyRef struct {
	Name string `json:"name,omitempty"`
}

type ChaincodeImage struct {
	Name   string `json:"name"`
	Digest string `json:"digest"`

	// +optional
	// +kubebuilder:default:=Always
	PullSecret string `json:"pullSecret"`
}

type ChaincodeHistory struct {
	Version         string         `json:"version"`
	Image           ChaincodeImage `json:"image"`
	ExternalBuilder string         `json:"externalBuilder"`
	UpgradeTime     metav1.Time    `json:"upgradeTime"`
}

type ChaincodeCondition struct {
	// +optional
	Type ChaincodeConditionType `json:"type"`
	// +optional
	Status metav1.ConditionStatus `json:"status"`
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// +optional
	Reason string `json:"reason"`
	// +optional
	Message string `json:"message"`
	// +optional
	// if an error occurs, which step to start from
	NextStage ChaincodeConditionType `json:"nextStage"`
}

type ChaincodeStatus struct {
	// Chaincode upgrade history
	// +optional
	History []ChaincodeHistory `json:"history,omitempty"`
	// +optional
	Phase ChaincodePhase `json:"phase,omitempty"`
	// +optional
	Conditions []ChaincodeCondition `json:"conditions,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
	// +optional
	Reason string `json:"reason,omitempty"`
	// LastHeartbeatTime is when the controller reconciled this component
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +optional
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime,omitempty"`
	// +kubebuilder:validation:Minimum:=1
	Sequence int64 `json:"sequence"`
}

func init() {
	SchemeBuilder.Register(&Chaincode{}, &ChaincodeList{})
}
