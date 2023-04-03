package v1beta1

import (
	"context"
	"errors"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	NETWORK_INITIATOR_LABEL  = "bestchains.network.initiator"
	NETWORK_FEDERATION_LABEL = "bestchains.network.federation"
)

var MemberMisMatchError = errors.New("mismatch members")

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

func (network *Network) GetOrdererName() string {
	return network.GetName()
}

func (network *Network) GetOrdererNamespace() string {
	return (&Organization{ObjectMeta: metav1.ObjectMeta{Name: network.GetInitiatorMember()}}).GetUserNamespace()
}

// MergeMembers The function is used to update the network.spec.members.
// members is the list of members of the federation,
// so the last value of network.sepc.members must be members,
// but we need to update the duplicate members of network.spec.members to members.
func (network *Network) MergeMembers(members []Member) {
	// Make a copy of the data to avoid modifying the original data
	target := make([]Member, len(members))
	copy(target, members)
	needMembers := make(map[string]int)

	now := metav1.Now()
	for index, member := range target {
		needMembers[member.Name] = index
		target[index].JoinedAt = &now
		target[index].Initiator = false
	}

	// The members of the network are found in the list of members of the federation,
	// we need to update the member information of the network into members,
	// if not, it may cause the fields such as initiator of the network to mismatch with the previous ones.
	for _, member := range network.Spec.Members {
		if refIndex, ok := needMembers[member.Name]; ok {
			target[refIndex] = member
		}
	}

	network.Spec.Members = target
}

// HaveSameMembers Determine if networks and alliances have the same members
// there are two clustered clients here, as there may be some structures that contain different clients.
// the update parameter is whether to update the members of the network when the network does not match the members of the federation
// If the members do not match the function will return an error, even if update is true, otherwise it will return nil.
func (network *Network) HaveSameMembers(ctx context.Context, r client.Reader, updateMembers bool) error {
	fed := &Federation{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: network.Spec.Federation}, fed); err != nil {
		return err
	}

	// If the member list lengths are different, there must be a mismatch
	if len(fed.Spec.Members) != len(network.Spec.Members) {
		if updateMembers {
			network.MergeMembers(fed.Spec.Members)
		}
		return MemberMisMatchError
	}

	fedMembers := make(map[string]struct{})
	for _, member := range fed.Spec.Members {
		fedMembers[member.Name] = struct{}{}
	}

	// If a member of network is not in the list of federation, it is also a mismatch.
	for _, member := range network.Spec.Members {
		if _, ok := fedMembers[member.Name]; !ok {
			if updateMembers {
				network.MergeMembers(fed.Spec.Members)
			}
			return MemberMisMatchError
		}
	}
	return nil
}
