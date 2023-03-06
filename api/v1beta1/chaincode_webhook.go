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

	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// log is for logging in this package.
var ccLogger = logf.Log.WithName("chaincode-resource")

//+kubebuilder:webhook:path=/mutate-ibp-com-v1beta1-chaincode,mutating=true,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=chaincodes,verbs=create;update,versions=v1beta1,name=chaincode.mutate.webhook,admissionReviewVersions=v1

var _ defaulter = &Chaincode{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Chaincode) Default(ctx context.Context, client client.Client, user authenticationv1.UserInfo) {
	ccLogger.Info("default", "name", r.Name, "user", user.String())
}

//+kubebuilder:webhook:path=/validate-ibp-com-v1beta1-chaincode,mutating=false,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=chaincodes,verbs=create;update;delete,versions=v1beta1,name=chaincode.validate.webhook,admissionReviewVersions=v1

var _ validator = &Chaincode{}

// checkChAndEp Check if both ch and ep are present
func checkChAndEp(c client.Client, chName, epName string) error {
	ch := &Channel{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: chName}, ch); err != nil {
		return err
	}

	ep := &EndorsePolicy{}
	return c.Get(context.TODO(), types.NamespacedName{Name: epName}, ep)
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Chaincode) ValidateCreate(ctx context.Context, client client.Client, user authenticationv1.UserInfo) error {
	ccLogger.Info("validate create", "name", r.Name, "user", user.String())
	return checkChAndEp(client, r.Spec.Channel, r.Spec.EndorsePolicyRef.Name)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Chaincode) ValidateUpdate(ctx context.Context, client client.Client, old runtime.Object, user authenticationv1.UserInfo) error {
	ccLogger.Info("validate update", "name", r.Name, "user", user.String())
	oldcc := old.(*Chaincode)
	return checkChAndEp(client, oldcc.Spec.Channel, oldcc.Spec.EndorsePolicyRef.Name)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Chaincode) ValidateDelete(ctx context.Context, c client.Client, user authenticationv1.UserInfo) error {
	ccLogger.Info("validate delete", "name", r.Name, "user", user.String())
	if err := checkChAndEp(c, r.Spec.Channel, r.Spec.EndorsePolicyRef.Name); err != nil {
		return err
	}
	if r.Status.Phase != ChaincodePhaseUnapproved {
		return fmt.Errorf("it can only be deleted if the vote is not approved")
	}
	return nil
}
