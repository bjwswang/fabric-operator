package v1beta1

import (
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

func DifferClients(old []string, new []string) (added []string, removed []string) {
	// cache in map
	oldMapper := make(map[string]struct{}, len(old))
	for _, c := range old {
		oldMapper[c] = struct{}{}
	}

	// calculate differences
	for _, c := range new {

		// added: in new ,but not in old
		if _, ok := oldMapper[c]; !ok {
			added = append(added, c)
			continue
		}

		// delete the intersection
		delete(oldMapper, c)
	}

	for c := range oldMapper {
		removed = append(removed, c)
	}

	return
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
