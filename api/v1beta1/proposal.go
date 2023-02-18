package v1beta1

import (
	"context"
	"fmt"
	"os"

	k8sclient "github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (p *Proposal) GetVoteName(orgName string) string {
	return fmt.Sprintf("vote-%s-%s", orgName, p.GetName())
}

func (p *Proposal) GetVoteLabel() map[string]string {
	label := os.Getenv("OPERATOR_LABEL_PREFIX")
	if label == "" {
		label = "fabric"
	}

	return map[string]string{
		"app":                          p.GetName(),
		"creator":                      label,
		"release":                      "operator",
		"helm.sh/chart":                "ibm-" + label,
		"app.kubernetes.io/name":       label,
		"app.kubernetes.io/instance":   label + "proposal",
		"app.kubernetes.io/managed-by": label + "-operator",
	}
}

func (p *Proposal) GetCandidateOrganizations(ctx context.Context, client k8sclient.Client) ([]string, error) {
	federation := &Federation{}
	if err := client.Get(ctx, types.NamespacedName{Name: p.Spec.Federation}, federation); err != nil {
		return nil, err
	}
	orgs := make([]string, 0)
	switch p.GetPurpose() {
	case CreateFederationProposal, DissolveFederationProposal:
		for _, o := range federation.Spec.Members {
			orgs = append(orgs, o.Name)
		}
	case AddMemberProposal:
		for _, o := range federation.Spec.Members {
			orgs = append(orgs, o.Name)
		}
		orgs = append(orgs, p.Spec.AddMember.Members...)
	case DeleteMemberProposal:
		for _, o := range federation.Spec.Members {
			if o.Name == p.Spec.DeleteMember.Member {
				continue
			}
			orgs = append(orgs, o.Name)
		}
	case DissolveNetworkProposal:
		if exist := util.ContainsValue(p.Spec.DissolveNetwork.Name, federation.Status.Networks); !exist {
			return orgs, nil
		}
		network := &Network{}
		if err := client.Get(ctx, types.NamespacedName{Name: p.Spec.DissolveNetwork.Name}, network); err != nil {
			if apierrors.IsNotFound(err) {
				// After Dissolve the network will delete the network CR later,
				// and GetCandidateOrganizations should return empty at that time.
				return orgs, nil
			}
			return nil, err
		}
		for _, o := range network.Spec.Members {
			orgs = append(orgs, o.Name)
		}
	case ArchiveChannelProposal:
		channel := p.Spec.ProposalSource.ArchiveChannel.Channel
		ch := &Channel{}
		if err := client.Get(ctx, types.NamespacedName{Name: channel}, ch); err != nil {
			if apierrors.IsNotFound(err) {
				return orgs, nil
			}
			return nil, err
		}
		for _, o := range ch.Spec.Members {
			orgs = append(orgs, o.Name)
		}
	case UnarchiveChannelProposal:
		channel := p.Spec.ProposalSource.UnarchiveChannel.Channel
		ch := &Channel{}
		if err := client.Get(ctx, types.NamespacedName{Name: channel}, ch); err != nil {
			if apierrors.IsNotFound(err) {
				return orgs, nil
			}
			return nil, err
		}
		for _, o := range ch.Spec.Members {
			orgs = append(orgs, o.Name)
		}
	case DeployChaincodeProposal1:
		for _, member := range p.Spec.DeployChaincode.Members {
			orgs = append(orgs, member.Name)
		}
	case UpgradeChaincodeProposal1:
		for _, member := range p.Spec.UpgradeChaincode.Members {
			orgs = append(orgs, member.Name)
		}
	}
	return orgs, nil
}

const (
	CreateFederationProposal = 1 << iota
	AddMemberProposal
	DeleteMemberProposal
	DissolveFederationProposal
	DissolveNetworkProposal
	ArchiveChannelProposal
	UnarchiveChannelProposal
	DeployChaincodeProposal1
	UpgradeChaincodeProposal1
)

func (p *Proposal) GetPurpose() uint {
	var t uint = 0
	if p.Spec.CreateFederation != nil {
		t = t | CreateFederationProposal
	}
	if p.Spec.AddMember != nil {
		t = t | AddMemberProposal
	}
	if p.Spec.DeleteMember != nil {
		t = t | DeleteMemberProposal
	}
	if p.Spec.DissolveFederation != nil {
		t = t | DissolveFederationProposal
	}
	if p.Spec.DissolveNetwork != nil {
		t = t | DissolveNetworkProposal
	}
	if p.Spec.ArchiveChannel != nil {
		t = t | ArchiveChannelProposal
	}
	if p.Spec.UnarchiveChannel != nil {
		t = t | UnarchiveChannelProposal
	}
	if p.Spec.DeployChaincode != nil {
		t = t | DeployChaincodeProposal1
	}
	if p.Spec.UpgradeChaincode != nil {
		t = t | UpgradeChaincodeProposal1
	}
	return t
}

func (p *Proposal) IsPurpose(purpose uint) bool {
	return (p.GetPurpose() & purpose) != 0
}

func (p *Proposal) SelfType() string {
	switch p.GetPurpose() {
	case AddMemberProposal:
		return "AddMemberProposal"
	case CreateFederationProposal:
		return "CreateFederationProposal"
	case DeleteMemberProposal:
		return "DeleteMemberProposal"
	case DissolveFederationProposal:
		return "DissolveFederationProposal"
	case DissolveNetworkProposal:
		return "DissolveNetworkProposal"
	case ArchiveChannelProposal:
		return "ArchiveChannelProposal"
	case UnarchiveChannelProposal:
		return "UnarchiveChannelProposal"
	case DeployChaincodeProposal1:
		return "DeployChaincode"
	case UpgradeChaincodeProposal1:
		return "UpgradeChaincode"
	default:
		return ""
	}
}
