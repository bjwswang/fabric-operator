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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	errNoFederation          = errors.New("cant find federation")
	errMemberNotInFederation = errors.New("some member not belongs to this federation")
	errOnlyDissolvedNetwork  = errors.New("only dissolved network can be deleted")
)

// log is for logging in this package.
var networklog = logf.Log.WithName("network-resource")

//+kubebuilder:webhook:path=/mutate-ibp-com-v1beta1-network,mutating=true,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=networks,verbs=create;update,versions=v1beta1,name=network.mutate.webhook,admissionReviewVersions=v1

var _ defaulter = &Network{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Network) Default(ctx context.Context, client client.Client, user authenticationv1.UserInfo) {
	networklog.Info("default", "name", r.Name, "user", user.String())
	r.Spec.OrderSpec.License.Accept = true
	r.Spec.OrderSpec.OrdererType = "etcdraft"
	if r.Spec.OrderSpec.ClusterSize == 0 {
		r.Spec.OrderSpec.ClusterSize = 1
	}

	for index := range r.Spec.Members {
		now := metav1.Now()
		r.Spec.Members[index].JoinedAt = &now
	}
}

//+kubebuilder:webhook:path=/validate-ibp-com-v1beta1-network,mutating=false,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=networks,verbs=create;update;delete,versions=v1beta1,name=network.validate.webhook,admissionReviewVersions=v1

var _ validator = &Network{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Network) ValidateCreate(ctx context.Context, client client.Client, user authenticationv1.UserInfo) error {
	networklog.Info("validate create", "name", r.Name, "user", user.String())

	if err := r.HaveSameMembers(ctx, client, nil, false); err != nil {
		return err
	}
	if err := validateInitiator(ctx, client, user, r.Spec.Members); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Network) ValidateUpdate(ctx context.Context, client client.Client, old runtime.Object, user authenticationv1.UserInfo) error {
	networklog.Info("validate update", "name", r.Name, "user", user.String())
	if err := r.HaveSameMembers(ctx, client, nil, false); err != nil {
		return err
	}

	if isSuperUser(ctx, user) {
		return nil
	}

	if err := validateInitiator(ctx, client, user, r.Spec.Members); err != nil {
		return err
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Network) ValidateDelete(ctx context.Context, client client.Client, user authenticationv1.UserInfo) error {
	networklog.Info("validate delete", "name", r.Name, "user", user.String())
	if err := validateInitiator(ctx, client, user, r.Spec.Members); err != nil {
		return err
	}

	if r.Status.CRStatus.Type != NetworkDissoleved {
		return errOnlyDissolvedNetwork
	}

	return nil
}

func validateMemberInFederation(ctx context.Context, c client.Client, fedName string, members []Member) error {
	fed := &Federation{}
	fed.Name = fedName
	if err := c.Get(ctx, client.ObjectKeyFromObject(fed), fed); err != nil {
		if apierrors.IsNotFound(err) {
			return errNoFederation
		}
		return errors.Wrap(err, "failed to get federation")
	}
	allMembers := make(map[string]bool, len(fed.Spec.Members))
	for _, m := range fed.Spec.Members {
		allMembers[m.Name] = true
	}
	for _, m := range members {
		if ok := allMembers[m.Name]; !ok {
			return errors.Wrapf(errMemberNotInFederation, "allMembers:%#v, members:%#v", allMembers, members)
		}
	}
	return nil
}
