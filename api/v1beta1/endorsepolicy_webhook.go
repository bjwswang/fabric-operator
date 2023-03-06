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

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/policydsl"
	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// log is for logging in this package.
var epLogger = logf.Log.WithName("endorsement-policy-resource")

//+kubebuilder:webhook:path=/mutate-ibp-com-v1beta1-endorsepolicy,mutating=true,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=endorsepolicies,verbs=create;update,versions=v1beta1,name=endorsepolicy.mutate.webhook,admissionReviewVersions=v1

var _ defaulter = &EndorsePolicy{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *EndorsePolicy) Default(ctx context.Context, client client.Client, user authenticationv1.UserInfo) {
	epLogger.Info("default", "name", r.Name, "user", user.String())
}

//+kubebuilder:webhook:path=/validate-ibp-com-v1beta1-endorsepolicy,mutating=false,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=endorsepolicies,verbs=create;update;delete,versions=v1beta1,name=endorsepolicy.validate.webhook,admissionReviewVersions=v1

var _ validator = &EndorsePolicy{}

// checkCh Check if both ch and ep are present
func checkCh(c client.Client, chName string) error {
	ch := &Channel{}
	return c.Get(context.TODO(), types.NamespacedName{Name: chName}, ch)
}

func isPolicyValidate(p string) bool {
	_, err := policydsl.FromString(p)
	if err != nil {
		epLogger.Error(err, "")
	}
	return err == nil
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *EndorsePolicy) ValidateCreate(ctx context.Context, client client.Client, user authenticationv1.UserInfo) error {
	epLogger.Info("validate create", "name", r.Name, "user", user.String())
	if !isPolicyValidate(r.Spec.Value) {
		return fmt.Errorf("invalid policy %s", r.Spec.Value)
	}
	return checkCh(client, r.Spec.Channel)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *EndorsePolicy) ValidateUpdate(ctx context.Context, client client.Client, old runtime.Object, user authenticationv1.UserInfo) error {
	epLogger.Info("validate update", "name", r.Name, "user", user.String())
	if !isPolicyValidate(r.Spec.Value) {
		return fmt.Errorf("invalid policy %s", r.Spec.Value)
	}
	return checkCh(client, r.Spec.Channel)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *EndorsePolicy) ValidateDelete(ctx context.Context, c client.Client, user authenticationv1.UserInfo) error {
	epLogger.Info("validate delete", "name", r.Name, "user", user.String())

	ccList := &ChaincodeList{}
	if err := c.List(context.TODO(), ccList, client.MatchingLabels{ChaincodeUsedEndorsementPolicy: r.GetName()}); err != nil {
		return err
	}
	if len(ccList.Items) > 0 {
		return fmt.Errorf("there are other chaincode that are using this policy and cannot be deleted")
	}
	return nil
}
