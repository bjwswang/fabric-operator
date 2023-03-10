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

	iam "github.com/IBM-Blockchain/fabric-operator/api/iam/v1alpha1"
	"github.com/pkg/errors"
	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	errAdminIsEmpty      = errors.New("the organization's admin is empty")
	errAdminCantBeClient = errors.New("user can't be admin and client at the same time")
	errUserNotFound      = errors.New("user not found")
	errUserDuplicated    = errors.New("found more than one user with same username")
	errHasNetwork        = errors.New("the organization is initiator of one network")
	errHasFederation     = errors.New("the organization is initiator of one federation")
)

// log is for logging in this package.
var organizationlog = logf.Log.WithName("organization-resource")

//+kubebuilder:webhook:path=/mutate-ibp-com-v1beta1-organization,mutating=true,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=organizations,verbs=create;update,versions=v1beta1,name=organization.mutate.webhook,admissionReviewVersions=v1

var _ defaulter = &Organization{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Organization) Default(ctx context.Context, client client.Client, user authenticationv1.UserInfo) {
	organizationlog.Info("default", "name", r.Name, "user", user.String())
	if r.Spec.DisplayName == "" {
		r.Spec.DisplayName = r.Name
	}
}

//+kubebuilder:webhook:path=/validate-ibp-com-v1beta1-organization,mutating=false,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=organizations,verbs=create;update;delete,versions=v1beta1,name=organization.validate.webhook,admissionReviewVersions=v1

var _ validator = &Organization{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Organization) ValidateCreate(ctx context.Context, client client.Client, user authenticationv1.UserInfo) error {
	organizationlog.Info("validate create", "name", r.Name, "user", user.String())
	if !isSuperUser(ctx, user) && r.Spec.Admin != user.Username {
		return errNoPermission
	}

	if len(r.Spec.Clients) != 0 {
		for _, orgClient := range r.Spec.Clients {
			if orgClient == r.Spec.Admin {
				return errAdminCantBeClient
			}
			if err := r.validateUser(ctx, client, orgClient); err != nil {
				return errors.Wrap(err, "client add")
			}
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Organization) ValidateUpdate(ctx context.Context, client client.Client, old runtime.Object, user authenticationv1.UserInfo) error {
	organizationlog.Info("validate update", "name", r.Name, "user", user.String())
	oldOrg := old.(*Organization)
	if !isSuperUser(ctx, user) && oldOrg.Spec.Admin != user.Username {
		return errNoPermission
	}
	if r.Spec.Admin == "" {
		return errAdminIsEmpty
	}
	if oldOrg.Spec.Admin != r.Spec.Admin {
		if err := r.validateUser(ctx, client, r.Spec.Admin); err != nil {
			return errors.Wrap(err, "admin update")
		}
	}

	if len(r.Spec.Clients) != 0 {
		for _, orgClient := range r.Spec.Clients {
			if orgClient == r.Spec.Admin {
				return errAdminCantBeClient
			}
			if err := r.validateUser(ctx, client, orgClient); err != nil {
				return errors.Wrap(err, "client update")
			}
		}
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Organization) ValidateDelete(ctx context.Context, c client.Client, user authenticationv1.UserInfo) error {
	organizationlog.Info("validate delete", "name", r.Name, "user", user.String())
	if !isSuperUser(ctx, user) && r.Spec.Admin != user.Username {
		return errNoPermission
	}
	// validate whether the organization is the initiator of a federation(initiator responsibility)
	federationList := &FederationList{}
	if err := c.List(ctx, federationList); err != nil {
		return errors.Wrap(err, "cant get federation list")
	}
	for _, fed := range federationList.Items {
		for _, m := range fed.Spec.Members {
			if m.Initiator && m.Name == r.Name {
				return errHasFederation
			}
		}
	}
	// validate whether the organization is the initiator of a network(initiator responsibility)
	networkList := &NetworkList{}
	if err := c.List(ctx, networkList); err != nil {
		return errors.Wrap(err, "cant get network list")
	}
	for _, net := range networkList.Items {
		for _, m := range net.Spec.Members {
			if m.Initiator && m.Name == r.Name {
				return errHasNetwork
			}
		}
	}
	return nil
}

func (r *Organization) validateUser(ctx context.Context, c client.Client, username string) error {
	_, err := r.getUser(ctx, c, username)
	if err != nil {
		return err
	}
	return nil
}

func (r *Organization) getUser(ctx context.Context, c client.Client, username string) (*iam.User, error) {
	selector, err := labels.Parse(fmt.Sprintf("t7d.io.username=%s", username))
	if err != nil {
		return nil, err
	}
	userList := &iam.UserList{}
	err = c.List(context.TODO(), userList, &client.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}
	if len(userList.Items) == 0 {
		return nil, errors.Wrapf(errUserNotFound, "user: %s", username)
	}
	if len(userList.Items) > 1 {
		return nil, errors.Wrapf(errUserDuplicated, "user: %s", username)
	}
	return &userList.Items[0], nil
}
