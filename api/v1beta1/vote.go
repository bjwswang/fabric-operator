package v1beta1

func (v *Vote) GetNamespacedName() NamespacedName {
	return NamespacedName{
		Name:      v.GetName(),
		Namespace: v.GetNamespace(),
	}
}

func (v *Vote) GetOrganization() NamespacedName {
	return NamespacedName{
		Name:      v.Spec.OrganizationName,
		Namespace: v.GetNamespace(),
	}
}
