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

package chaincode

import (
	"context"
	"fmt"
	"os"
	"strings"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	common1 "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/policydsl"
	"k8s.io/apimachinery/pkg/types"
)

// ChaincodeStorageDir /demo/<channel-name>/<chaincodeName>/<crname-chaincodeid-version>.tgz
func ChaincodeStorageDir(baseDir string, instance *current.Chaincode) string {
	if baseDir == "" {
		if baseDir = os.Getenv("STORE"); baseDir == "" {
			baseDir = "/bestchains/chaincodes"
		}
	}
	if _, err := os.Stat(baseDir); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		os.MkdirAll(baseDir, 0755)
	}
	baseDir = strings.TrimSuffix(baseDir, "/")
	return fmt.Sprintf("%s/%s/%s", baseDir, instance.Spec.Channel, instance.GetName())
}

func ChaincodePacakgeFile(instance *current.Chaincode) string {
	return fmt.Sprintf("%s-%s-%s.tgz", instance.GetName(), instance.Spec.ID, instance.Spec.Version)
}

func ChaincodeEndorsePolicy(cli controllerclient.Client, instance *current.Chaincode) (*common1.SignaturePolicyEnvelope, error) {
	policyName := instance.Spec.EndorsePolicyRef.Name
	policy := &current.EndorsePolicy{}
	if err := cli.Get(context.TODO(), types.NamespacedName{Name: policyName}, policy); err != nil {
		return nil, err
	}

	return policydsl.FromString(policy.Spec.Value)
}
