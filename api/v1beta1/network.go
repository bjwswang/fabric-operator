package v1beta1

import "os"

func init() {
	SchemeBuilder.Register(&Network{}, &NetworkList{})
}

func (network *Network) GetLabels() map[string]string {
	label := os.Getenv("OPERATOR_LABEL_PREFIX")
	if label == "" {
		label = "fabric"
	}

	return map[string]string{
		"app":                          network.GetName(),
		"creator":                      label,
		"release":                      "operator",
		"helm.sh/chart":                "ibm-" + label,
		"app.kubernetes.io/name":       label,
		"app.kubernetes.io/instance":   label + "network",
		"app.kubernetes.io/managed-by": label + "-operator",
	}
}

func (network *Network) GetMembers() []Member {
	return network.Spec.Members
}

func (network *Network) GetInitiatorMember() string {
	for _, m := range network.GetMembers() {
		if m.Initiator {
			return m.Name
		}
	}
	return ""
}

func (network *Network) HasFederation() bool {
	return network.Spec.Federation != ""
}

func (network *Network) HasOrder() bool {
	return network.Spec.OrderSpec.License.Accept
}

func (network *Network) HasType() bool {
	return network.Status.CRStatus.Type != ""
}

func (network *Network) HasMembers() bool {
	return len(network.Spec.Members) != 0
}

func (networkStatus *NetworkStatus) AddChannel(channel string) bool {
	var conflict bool

	for _, f := range networkStatus.Channels {
		if f == channel {
			conflict = true
			break
		}
	}

	if !conflict {
		networkStatus.Channels = append(networkStatus.Channels, channel)
	}

	return conflict
}

func (networkStatus *NetworkStatus) DeleteChannel(channel string) bool {
	var exist bool
	var index int

	channels := networkStatus.Channels

	for curr, f := range channels {
		if f == channel {
			exist = true
			index = curr
			break
		}
	}

	if exist {
		networkStatus.Channels = append(channels[:index], channels[index+1:]...)
	}

	return exist
}
