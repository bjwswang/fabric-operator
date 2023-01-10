package v1beta1

import (
	"context"
	"fmt"
	"os"

	k8sclient "github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/util"
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

func (p *Proposal) GetCandidateOrganizations(ctx context.Context, client k8sclient.Client) ([]NamespacedName, error) {
	federation := &Federation{}
	if err := client.Get(ctx, types.NamespacedName{Name: p.Spec.Federation}, federation); err != nil {
		return nil, err
	}
	orgs := make([]NamespacedName, 0)
	switch p.GetPurpose() {
	case CreateFederationProposal, DissolveFederationProposal:
		for _, o := range federation.Spec.Members {
			orgs = append(orgs, NamespacedName{
				Name:      o.Name,
				Namespace: o.Namespace,
			})
		}
	case AddMemberProposal:
		for _, o := range federation.Spec.Members {
			orgs = append(orgs, NamespacedName{
				Name:      o.Name,
				Namespace: o.Namespace,
			})
		}
		for _, o := range p.Spec.AddMember.Members {
			orgs = append(orgs, NamespacedName{
				Name:      o.Name,
				Namespace: o.Namespace,
			})
		}
	case DeleteMemberProposal:
		for _, o := range federation.Spec.Members {
			if o.Name == p.Spec.DeleteMember.Member.Name && o.Namespace == p.Spec.DeleteMember.Member.Namespace {
				continue
			}
			orgs = append(orgs, NamespacedName{
				Name:      o.Name,
				Namespace: o.Namespace,
			})
		}
	case DissolveNetworkProposal:
		if exist := util.ContainsValue(p.Spec.DissolveNetwork.Name, federation.Status.Networks); !exist {
			return orgs, nil
		}
		network := &Network{}
		if err := client.Get(ctx, types.NamespacedName{Name: p.Spec.DissolveNetwork.Name}, network); err != nil {
			return nil, err
		}
		for _, o := range network.Spec.Members {
			orgs = append(orgs, NamespacedName{
				Name:      o.Name,
				Namespace: o.Namespace,
			})
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
	default:
		return ""
	}
}
