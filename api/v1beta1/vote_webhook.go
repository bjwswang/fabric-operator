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
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	ErrChangeVoteDecision = errors.New("decision can not change after vote")
)

// log is for logging in this package.
var votelog = logf.Log.WithName("vote-resource")

func (v *Vote) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(v).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-ibp-com-v1beta1-vote,mutating=true,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=votes,verbs=create;update,versions=v1beta1,name=vote.mutate.webhook,admissionReviewVersions=v1

var _ webhook.Defaulter = &Vote{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (v *Vote) Default() {
	votelog.Info("default", "name", v.Name)

}

//+kubebuilder:webhook:path=/validate-ibp-com-v1beta1-vote,mutating=false,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=votes,verbs=create;update;delete,versions=v1beta1,name=vote.validate.webhook,admissionReviewVersions=v1

var _ webhook.Validator = &Vote{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (v *Vote) ValidateCreate() error {
	votelog.Info("validate create", "name", v.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (v *Vote) ValidateUpdate(old runtime.Object) error {
	votelog.Info("validate update", "name", v.Name)
	instance := old.(*Vote)
	if instance.Spec.Decision != nil {
		if v.Spec.Decision == nil || *instance.Spec.Decision != *v.Spec.Decision {
			return ErrChangeVoteDecision
		}
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (v *Vote) ValidateDelete() error {
	votelog.Info("validate delete", "name", v.Name)
	return nil
}
