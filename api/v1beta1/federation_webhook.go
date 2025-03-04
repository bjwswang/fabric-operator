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
	"fmt"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// log is for logging in this package.
var federationlog = logf.Log.WithName("federation-resource")

var (
	errNoInitiator   = errors.New("should have one initiator in members")
	errMoreInitiator = errors.New("should have only one initiator in members")
	errNoPermission  = errors.New("the operator is not the admin user of the initiator organization")
	errInvalidPolicy = errors.New("the policy is invalid")
	errUpdatePolicy  = errors.New("do not support update policy now")
)

//+kubebuilder:webhook:path=/mutate-ibp-com-v1beta1-federation,mutating=true,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=federations,verbs=create;update,versions=v1beta1,name=federation.mutate.webhook,admissionReviewVersions=v1

var _ defaulter = &Federation{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Federation) Default(ctx context.Context, client client.Client, user authenticationv1.UserInfo) {
	federationlog.Info("default", "name", r.Name, "user", user.String())
}

//+kubebuilder:webhook:path=/validate-ibp-com-v1beta1-federation,mutating=false,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=federations,verbs=create;update;delete,versions=v1beta1,name=federation.validate.webhook,admissionReviewVersions=v1

var _ validator = &Federation{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Federation) ValidateCreate(ctx context.Context, client client.Client, user authenticationv1.UserInfo) error {
	federationlog.Info("validate create", "name", r.Name, "user", user.String())
	if ok := PolicyMap[r.Spec.Policy.String()]; !ok {
		return errInvalidPolicy
	}

	if err := validateInitiator(ctx, client, user, r.Spec.Members); err != nil {
		return err
	}

	for _, member := range r.Spec.Members {
		err := validateOrganization(ctx, client, member.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Federation) ValidateUpdate(ctx context.Context, client client.Client, old runtime.Object, user authenticationv1.UserInfo) error {
	federationlog.Info("validate update", "name", r.Name, "user", user.String())
	oldFederation := old.(*Federation)
	if r.Spec.Policy.String() != oldFederation.Spec.Policy.String() {
		return errUpdatePolicy
	}

	if err := validateInitiator(ctx, client, user, r.Spec.Members); err != nil {
		return err
	}

	return validateFedMemberUpdate(ctx, client, oldFederation.Spec.Members, r.Spec.Members, r.GetName())
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Federation) ValidateDelete(ctx context.Context, client client.Client, user authenticationv1.UserInfo) error {
	federationlog.Info("validate delete", "name", r.Name, "user", user.String())

	if err := validateInitiator(ctx, client, user, r.Spec.Members); err != nil {
		return err
	}

	if r.Status.Type != FederationFailed && r.Status.Type != FederationDissolved {
		return errors.Errorf("forbid to delete federation when it is at status %s", r.Status.Type)
	}

	if len(r.Status.Networks) != 0 {
		return errors.Errorf("forbid to delete federation when it still has %d networks", len(r.Status.Networks))
	}

	return nil
}

// validateFedMemberUpdate ensure the initiator of a network or members of a channel are not allowed to be deleted.
// Duplicate members are also not allowed among newMembers.
func validateFedMemberUpdate(ctx context.Context, c client.Client, oldMembers, newMembers []Member, federationName string) error {
	deleted := make(map[string]struct{})
	for _, member := range newMembers {
		if _, exist := deleted[member.Name]; exist {
			return fmt.Errorf("has multiple members with the same name: %s", member.Name)
		}
		deleted[member.Name] = struct{}{}
	}
	chList := &ChannelList{}
	if err := c.List(ctx, chList); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	networkList := &NetworkList{}
	if err := c.List(ctx, networkList); err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	networkNeedCheck := make(map[string]bool) // FIXME: After we figure out why the list uses labelselector fail, we can replace all lists with only filter the objects we need.
	for _, member := range oldMembers {
		if _, ok := deleted[member.Name]; !ok {
			for _, i := range networkList.Items {
				if i.Spec.Federation != federationName {
					continue
				}
				networkNeedCheck[i.GetName()] = true
				if i.Labels != nil && i.Labels[NETWORK_INITIATOR_LABEL] == member.Name {
					return fmt.Errorf("can't remove federation member %s, it's the initiator of netowrk %s", member.Name, i.GetName())
				}
			}

			for _, ch := range chList.Items {
				if shouldCheck := networkNeedCheck[ch.Spec.Network]; !shouldCheck {
					continue
				}
				for _, chMember := range ch.Spec.Members {
					if member.Name == chMember.Name {
						return fmt.Errorf("can't remove federation member %s, it's the member of channel %s", member.Name, ch.GetName())
					}
				}
			}
		}
	}
	return nil
}

func validateInitiator(ctx context.Context, c client.Client, user authenticationv1.UserInfo, members []Member) error {
	org := &Organization{}
	for _, m := range members {
		if !m.Initiator {
			continue
		}
		if org.Name != "" {
			return errMoreInitiator
		}
		org.Name = m.Name
	}
	if org.Name == "" {
		return errNoInitiator
	}
	if !isSuperUser(ctx, user) {
		err := c.Get(ctx, client.ObjectKeyFromObject(org), org)
		if err != nil {
			return errors.Wrap(err, "failed to get initiator organization")
		}
		if org.Spec.Admin != user.Username {
			return errNoPermission
		}
	}
	return nil
}

func validateOrganization(ctx context.Context, c client.Client, organization string) error {
	federationlog.Info("validate organization: %s", organization)
	org := &Organization{}
	org.Name = organization
	err := c.Get(ctx, client.ObjectKeyFromObject(org), org)
	if err != nil {
		return errors.Wrapf(err, "failed to get organization %s", organization)
	}

	if org.Status.Type == Error {
		return errors.Errorf("organization %s has error %s:%s", org.Name, org.Status.Reason, org.Status.Message)
	}

	return nil
}
