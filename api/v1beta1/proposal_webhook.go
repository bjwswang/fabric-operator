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
	"context"
	"errors"
	"fmt"
	"time"

	authenticationv1 "k8s.io/api/authentication/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	errChangeProposalPurpose  = errors.New("the purpose of the proposal cannot be changed")
	errNullProposalPurpose    = errors.New("the proposal should have a purpose")
	errMoreProposalPurpose    = errors.New("the proposal should have only one purpose")
	errChangeInitiator        = errors.New("the initiator of the proposal cannot be changed")
	errChangeFederation       = errors.New("the federation of the proposal cannot be changed")
	errChannelNotFound        = errors.New("the relevant channel in the proposal cannot be found")
	errChannelInError         = errors.New("the relevant channel in the proposal is in Error status")
	errChannelAlreadyArchived = errors.New("the relevant channel in the proposal is already archived")
	errChannelNotArchivedYet  = errors.New("the relevant channel in the proposal not archived yet")
)

// log is for logging in this package.
var proposallog = logf.Log.WithName("proposal-resource")

//+kubebuilder:webhook:path=/mutate-ibp-com-v1beta1-proposal,mutating=true,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=proposals,verbs=create;update,versions=v1beta1,name=proposal.mutate.webhook,admissionReviewVersions=v1

var _ defaulter = &Proposal{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Proposal) Default(ctx context.Context, client client.Client, user authenticationv1.UserInfo) {
	proposallog.Info("default", "name", r.Name, "user", user.String())
	if r.Spec.StartAt.IsZero() {
		r.Spec.StartAt = metav1.Now()
	}
	if r.Spec.EndAt.IsZero() {
		r.Spec.EndAt = metav1.NewTime(time.Now().Add(time.Hour * 24))
	}
}

//+kubebuilder:webhook:path=/validate-ibp-com-v1beta1-proposal,mutating=false,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=proposals,verbs=create;update;delete,versions=v1beta1,name=proposal.validate.webhook,admissionReviewVersions=v1

var _ validator = &Proposal{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Proposal) ValidateCreate(ctx context.Context, client client.Client, user authenticationv1.UserInfo) error {
	proposallog.Info("validate create", "name", r.Name, "user", user.String())
	purpose := r.GetPurpose()
	if purpose == 0 {
		return errNullProposalPurpose
	}
	if purpose&(purpose-1) != 0 {
		return errMoreProposalPurpose
	}
	if ok := PolicyMap[r.Spec.Policy.String()]; !ok {
		return errInvalidPolicy
	}

	fakeMembers := []Member{{Name: r.Spec.InitiatorOrganization, Initiator: true}}
	if err := validateMemberInFederation(ctx, client, r.Spec.Federation, fakeMembers); err != nil {
		return err
	}

	if err := validateInitiator(ctx, client, user, fakeMembers); err != nil {
		return err
	}

	if err := validateProposalSource(ctx, client, r.Spec.ProposalSource); err != nil {
		return err
	}

	return validateChaincodeBuild(ctx, client, r.Spec.ProposalSource)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Proposal) ValidateUpdate(ctx context.Context, client client.Client, old runtime.Object, user authenticationv1.UserInfo) error {
	proposallog.Info("validate update", "name", r.Name, "user", user.String())
	oldProposal := old.(*Proposal)
	if oldProposal.GetPurpose() != r.GetPurpose() {
		return errChangeProposalPurpose
	}
	if oldProposal.Spec.InitiatorOrganization != r.Spec.InitiatorOrganization {
		return errChangeInitiator
	}
	if oldProposal.Spec.Federation != r.Spec.Federation {
		return errChangeFederation
	}
	if r.Spec.Policy.String() != oldProposal.Spec.Policy.String() {
		return errUpdatePolicy
	}

	fakeMembers := []Member{{Name: r.Spec.InitiatorOrganization, Initiator: true}}
	if err := validateMemberInFederation(ctx, client, r.Spec.Federation, fakeMembers); err != nil {
		return err
	}

	if err := validateInitiator(ctx, client, user, fakeMembers); err != nil {
		return err
	}

	if err := validateProposalSource(ctx, client, r.Spec.ProposalSource); err != nil {
		return err
	}

	return validateChaincodeBuild(ctx, client, r.Spec.ProposalSource)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Proposal) ValidateDelete(ctx context.Context, client client.Client, user authenticationv1.UserInfo) error {
	proposallog.Info("validate delete", "name", r.Name, "user", user.String())

	if isSuperUser(ctx, user) {
		return nil
	}

	fakeMembers := []Member{{Name: r.Spec.InitiatorOrganization, Initiator: true}}
	if err := validateMemberInFederation(ctx, client, r.Spec.Federation, fakeMembers); err != nil {
		return err
	}

	if err := validateInitiator(ctx, client, user, fakeMembers); err != nil {
		return err
	}

	return nil
}

func validateProposalSource(ctx context.Context, c client.Client, proposalSource ProposalSource) error {
	if proposalSource.ArchiveChannel != nil {
		if err := validateChannel(ctx, c, proposalSource.ArchiveChannel.Channel, ArchiveChannelProposal); err != nil {
			return err
		}
	}

	if proposalSource.UnarchiveChannel != nil {
		if err := validateChannel(ctx, c, proposalSource.UnarchiveChannel.Channel, UnarchiveChannelProposal); err != nil {
			return err
		}
	}

	return nil
}

func validateChannel(ctx context.Context, c client.Client, channel string, purpose uint) error {
	ch := &Channel{}
	ch.Name = channel
	err := c.Get(ctx, client.ObjectKeyFromObject(ch), ch)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return errChannelNotFound
		}
		return err
	}

	if ch.Status.Type == Error {
		return errChannelInError
	}

	switch purpose {
	case ArchiveChannelProposal:
		if ch.Status.Type == ChannelArchived {
			return errChannelAlreadyArchived
		}
	case UnarchiveChannelProposal:
		if ch.Status.Type != ChannelArchived {
			return errChannelNotArchivedYet
		}
	}

	return nil
}

func validateChaincodeBuild(ctx context.Context, c client.Client, proposalSource ProposalSource) error {
	if proposalSource.DeployChaincode != nil && proposalSource.UpgradeChaincode != nil {
		return fmt.Errorf("the deployChaincode and upgradeChaincode cannot co-exist")
	}
	if proposalSource.UpgradeChaincode == nil {
		return nil
	}
	ccbName := proposalSource.UpgradeChaincode.ExternalBuilder
	if ccbName == "" {
		return fmt.Errorf("empty chaincodebuild name")
	}
	chaincodebuild := &ChaincodeBuild{}
	if err := c.Get(ctx, types.NamespacedName{Name: ccbName}, chaincodebuild); err != nil {
		return err
	}
	return chaincodebuild.HasImage()
}
