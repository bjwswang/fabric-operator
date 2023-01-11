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

	"github.com/pkg/errors"
	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	errHasNetwork = errors.New("the organization is initiator of one network")
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
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Organization) ValidateUpdate(ctx context.Context, client client.Client, old runtime.Object, user authenticationv1.UserInfo) error {
	organizationlog.Info("validate update", "name", r.Name, "user", user.String())
	oldOrg := old.(*Organization)
	if !isSuperUser(ctx, user) && oldOrg.Spec.Admin != user.Username {
		return errNoPermission
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Organization) ValidateDelete(ctx context.Context, c client.Client, user authenticationv1.UserInfo) error {
	organizationlog.Info("validate delete", "name", r.Name, "user", user.String())
	if !isSuperUser(ctx, user) && r.Spec.Admin != user.Username {
		return errNoPermission
	}
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
