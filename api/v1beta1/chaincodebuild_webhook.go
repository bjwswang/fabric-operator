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
	"reflect"

	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// log is for logging in this package.
var ccbLogger = logf.Log.WithName("chaincodebuild-resource")

//+kubebuilder:webhook:path=/mutate-ibp-com-v1beta1-chaincodebuild,mutating=true,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=chaincodebuilds,verbs=create;update,versions=v1beta1,name=chaincodebuild.mutate.webhook,admissionReviewVersions=v1

var _ defaulter = &ChaincodeBuild{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ChaincodeBuild) Default(ctx context.Context, client client.Client, user authenticationv1.UserInfo) {
	ccbLogger.Info("default", "name", r.Name, "user", user.String())
}

//+kubebuilder:webhook:path=/validate-ibp-com-v1beta1-chaincodebuild,mutating=false,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=chaincodebuilds,verbs=create;update;delete,versions=v1beta1,name=chaincodebuild.validate.webhook,admissionReviewVersions=v1

var _ validator = &ChaincodeBuild{}

func versionConflict(c client.Client, networkName, id, version string) error {
	network := &Network{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: networkName}, network); err != nil {
		ccbLogger.Error(err, "")
		return err
	}
	ccbList := &ChaincodeBuildList{}
	if err := c.List(context.TODO(), ccbList); err != nil {
		ccbLogger.Error(err, "")
		return err
	}

	checker := make(map[string]map[string]struct{})

	for _, item := range ccbList.Items {
		if _, ok := checker[item.GetName()]; !ok {
			checker[item.GetName()] = make(map[string]struct{})
		}
		checker[item.GetName()][fmt.Sprintf("%s-%s", item.Spec.ID, item.Spec.Version)] = struct{}{}
	}
	if v, ok := checker[networkName]; ok {
		if _, ok1 := v[fmt.Sprintf("%s-%s", id, version)]; ok1 {
			return fmt.Errorf("the same id and version exist under network")
		}
	}

	return nil
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ChaincodeBuild) ValidateCreate(ctx context.Context, client client.Client, user authenticationv1.UserInfo) error {
	ccbLogger.Info("validate create", "name", r.Name, "user", user.String())
	return versionConflict(client, r.Spec.Network, r.Spec.ID, r.Spec.Version)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ChaincodeBuild) ValidateUpdate(ctx context.Context, client client.Client, old runtime.Object, user authenticationv1.UserInfo) error {
	ccbLogger.Info("validate update", "name", r.Name, "user", user.String())
	oldCCB := old.(*ChaincodeBuild)
	if !reflect.DeepEqual(r.Spec, oldCCB.Spec) {
		return fmt.Errorf("chaincodebuild does not allow updates")
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ChaincodeBuild) ValidateDelete(ctx context.Context, c client.Client, user authenticationv1.UserInfo) error {
	ccbLogger.Info("validate delete", "name", r.Name, "user", user.String())

	chaincodeList := &ChaincodeList{}
	if err := c.List(context.TODO(), chaincodeList); err != nil {
		ccbLogger.Error(err, "")
		return err
	}

	for _, cc := range chaincodeList.Items {
		if cc.Spec.ExternalBuilder == r.GetName() {
			return fmt.Errorf("chaincodebuild is used by chaincode %s", cc.GetName())
		}
		for _, history := range cc.Status.History {
			if history.ExternalBuilder == r.GetName() {
				return fmt.Errorf("chaincodebuild is used by chaincode %s", cc.GetName())
			}
		}
	}

	return nil
}
