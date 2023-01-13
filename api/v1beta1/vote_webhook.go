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

	"github.com/IBM-Blockchain/fabric-operator/pkg/util"
	"github.com/pkg/errors"
	authenticationv1 "k8s.io/api/authentication/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	errChangeVoteDecision = errors.New("decision can not change after vote")
)

// log is for logging in this package.
var votelog = logf.Log.WithName("vote-resource")

//+kubebuilder:webhook:path=/mutate-ibp-com-v1beta1-vote,mutating=true,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=votes,verbs=create;update,versions=v1beta1,name=vote.mutate.webhook,admissionReviewVersions=v1

var _ defaulter = &Vote{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (v *Vote) Default(ctx context.Context, client client.Client, user authenticationv1.UserInfo) {
	votelog.Info("default", "name", v.Name, "user", user.String())

}

//+kubebuilder:webhook:path=/validate-ibp-com-v1beta1-vote,mutating=false,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=votes,verbs=create;update;delete,versions=v1beta1,name=vote.validate.webhook,admissionReviewVersions=v1

var _ validator = &Vote{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (v *Vote) ValidateCreate(ctx context.Context, client client.Client, user authenticationv1.UserInfo) error {
	votelog.Info("validate create", "name", v.Name, "user", user.String())

	if err := validateVoteProAndOrg(ctx, client, v.Spec.ProposalName, v.Spec.OrganizationName, v.Namespace); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (v *Vote) ValidateUpdate(ctx context.Context, client client.Client, old runtime.Object, user authenticationv1.UserInfo) error {
	votelog.Info("validate update", "name", v.Name, "user", user.String())
	instance := old.(*Vote)
	if instance.Spec.Decision != nil {
		if v.Spec.Decision == nil || *instance.Spec.Decision != *v.Spec.Decision {
			return errChangeVoteDecision
		}
	}

	if err := validateVoteProAndOrg(ctx, client, v.Spec.ProposalName, v.Spec.OrganizationName, v.Namespace); err != nil {
		return err
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (v *Vote) ValidateDelete(ctx context.Context, client client.Client, user authenticationv1.UserInfo) error {
	votelog.Info("validate delete", "name", v.Name, "user", user.String())

	if isSuperUser(ctx, user) {
		return nil
	}

	fakeMembers := []Member{{
		Name:      v.Spec.OrganizationName,
		Initiator: true,
	}}

	if err := validateInitiator(ctx, client, user, fakeMembers); err != nil {
		return err
	}

	return nil
}

func validateVoteProAndOrg(ctx context.Context, c client.Client, proposalName, orgName, orgNS string) error {
	pro := &Proposal{}
	pro.Name = proposalName
	if err := c.Get(ctx, client.ObjectKeyFromObject(pro), pro); err != nil {
		if apierrors.IsNotFound(err) {
			return errNoFederation
		}
		return errors.Wrap(err, "failed to get federation")
	}

	org := &Organization{}
	org.Name = orgName
	org.Namespace = orgNS
	if err := c.Get(ctx, client.ObjectKeyFromObject(org), org); err != nil {
		if apierrors.IsNotFound(err) {
			return errNoFederation
		}
		return errors.Wrap(err, "failed to get organization")
	}

	needCheckOrgInFed := true
	if pro.IsPurpose(AddMemberProposal) && util.ContainsValue(org.Name, pro.Spec.AddMember.Members) {
		needCheckOrgInFed = false
	}
	if !needCheckOrgInFed {
		return nil
	}
	fakeMembers := []Member{{
		Name:      org.Name,
		Initiator: true,
	}}
	if err := validateMemberInFederation(ctx, c, pro.Spec.Federation, fakeMembers); err != nil {
		return err
	}

	return nil
}
