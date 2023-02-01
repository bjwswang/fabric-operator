package v1beta1

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/types"
)

func init() {
	SchemeBuilder.Register(&Organization{}, &OrganizationList{})
}

func (organization *Organization) GetLabels() map[string]string {
	label := os.Getenv("OPERATOR_LABEL_PREFIX")
	if label == "" {
		label = "fabric"
	}

	return map[string]string{
		"app":                          organization.GetName(),
		"creator":                      label,
		"release":                      "operator",
		"helm.sh/chart":                "ibm-" + label,
		"app.kubernetes.io/name":       label,
		"app.kubernetes.io/instance":   label + "organization",
		"app.kubernetes.io/managed-by": label + "-operator",
	}
}
func (organization *Organization) GetNamespaced() types.NamespacedName {
	return types.NamespacedName{Name: organization.Name, Namespace: organization.GetUserNamespace()}
}

func (organization *Organization) GetUserNamespace() string {
	return organization.GetName()
}

func (organization *Organization) GetCAConnectinProfile() types.NamespacedName {
	return types.NamespacedName{Namespace: organization.GetUserNamespace(), Name: organization.GetName() + "-connection-profile"}
}
func (organization *Organization) GetMSPCrypto() types.NamespacedName {
	return types.NamespacedName{Namespace: organization.GetUserNamespace(), Name: fmt.Sprintf("%s-msp-crypto", organization.GetName())}
}

func (organization *Organization) GetCA() NamespacedName {
	return NamespacedName{Namespace: organization.GetUserNamespace(), Name: organization.GetName()}
}

func (organization *Organization) HasDisplayName() bool {
	return organization.Spec.DisplayName != ""
}

func (organization *Organization) HasAdmin() bool {
	return organization.Spec.Admin != ""
}

func (organization *Organization) HasType() bool {
	return organization.Status.CRStatus.Type != ""
}

func (organizationStatus *OrganizationStatus) AddFederation(federation NamespacedName) bool {
	var conflict bool

	for _, f := range organizationStatus.Federations {
		if f.String() == federation.String() {
			conflict = true
			break
		}
	}

	if !conflict {
		organizationStatus.Federations = append(organizationStatus.Federations, federation)
	}

	return conflict
}

func (organizationStatus *OrganizationStatus) DeleteFederation(federation NamespacedName) bool {
	var exist bool
	var index int

	federations := organizationStatus.Federations

	for curr, f := range federations {
		if f.String() == federation.String() {
			exist = true
			index = curr
			break
		}
	}

	if exist {
		organizationStatus.Federations = append(federations[:index], federations[index+1:]...)

	}

	return exist
}
