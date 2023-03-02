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
	"net/http"
	"os"
	"strings"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	lcpackager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/lifecycle"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"k8s.io/apimachinery/pkg/types"
)

const (
	approveOutputTemplate = `org[%s] peer [%s] approval status %s.`
	networkOrgLabel       = "bestchains.network.initiator"
)

func (c *baseChaincode) ApproveChaincode(instance *current.Chaincode) (string, error) {
	method := fmt.Sprintf("%s [base.chaincode.ApproveChaincode]", stepPrefix)
	signedPolicy, err := ChaincodeEndorsePolicy(c.client, instance)
	if err != nil {
		return err.Error(), err
	}
	packagePath := fmt.Sprintf("%s/%s", ChaincodeStorageDir("", instance), ChaincodePacakgeFile(instance))
	log.Info(fmt.Sprintf("%s full packagePath %s", method, packagePath))

	packageBytes, err := os.ReadFile(packagePath)
	if err != nil {
		log.Error(err, "failed to read package file")
		return err.Error(), err
	}

	packageID := lcpackager.ComputePackageID(instance.Spec.Label, packageBytes)
	log.Info(fmt.Sprintf("%s calculate packageID %s", method, packageID))

	ch := &current.Channel{}
	log.Info(fmt.Sprintf("%s get channel %s info", method, instance.Spec.Channel))
	if err := c.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Channel}, ch); err != nil {
		log.Error(err, "")
		return err.Error(), err
	}

	connectProfile, err := ProfileProvider(c.client, instance.Spec.Channel)
	if err != nil {
		log.Error(err, "")
		return err.Error(), err
	}
	orgPeer, peerAdmin, err := SetChannelPeerProfile(c.client, connectProfile, ch)
	if err != nil {
		log.Error(err, "")
		return err.Error(), err
	}

	log.Info(fmt.Sprintf("%s get network %s", method, ch.Spec.Network))
	network := current.Network{}
	if err = c.client.Get(context.TODO(), types.NamespacedName{Name: ch.Spec.Network}, &network); err != nil {
		log.Error(err, ")")
		return err.Error(), err
	}

	orderOrg := network.Labels[networkOrgLabel]
	log.Info(fmt.Sprintf("%s channel %s' network initiator %s", method, ch.GetName(), orderOrg))
	orderList, err := getOrderNodes(c.client, orderOrg, network.GetName())
	if err != nil {
		log.Error(err, "")
		return err.Error(), err
	}

	selectOne := ""
	for _, o := range orderList.Items {
		cur := current.NamespacedName{Name: o.GetName(), Namespace: o.GetNamespace()}
		if err = connectProfile.SetOrderer(c.client, cur); err != nil {
			log.Error(err, "")
			continue
		}
		selectOne = cur.String()
		break
	}
	if selectOne == "" {
		err = fmt.Errorf("org %s can't find orderer node", orderOrg)
		return err.Error(), err
	}

	peerConnector, err := NewChaincodeConnector(connectProfile)
	if err != nil {
		log.Error(err, fmt.Sprintf("%s chaincode get new connector error", method))
		return "chaincode get new connector error", err
	}

	buf := strings.Builder{}
	var finalErr error
	for orgName, p := range orgPeer {
		peer := current.NamespacedName{Name: p.GetName(), Namespace: p.GetNamespace()}
		orgAdminCtx := peerConnector.SDK().Context(fabsdk.WithUser(peerAdmin[peer.String()]), fabsdk.WithOrg(orgName))
		pc, err := resmgmt.New(orgAdminCtx)
		if err != nil {
			out := fmt.Sprintf(approveOutputTemplate, peer.Namespace, peer.Name, "SDKErr")
			finalErr = err
			log.Error(err, out)
			buf.WriteString(out)
			continue
		}

		log.Info(fmt.Sprintf("%s try to find out if the pacakge %s approved on the peer %s org %s", method, packageID, peer.String(), orgName))
		_, err = pc.LifecycleQueryApprovedCC(instance.Spec.Channel, resmgmt.LifecycleQueryApprovedCCRequest{
			Name:     instance.Spec.ID,
			Sequence: instance.Status.Sequence,
		}, resmgmt.WithTargetEndpoints(peer.String()))
		if err == nil {
			log.Info(fmt.Sprintf("%s chaincode %s has been approved on the peer %s", http.MethodHead, instance.GetName(), peer.String()))
			buf.WriteString(fmt.Sprintf(approveOutputTemplate, peer.Namespace, peer.Name, Approved))
			continue
		}
		log.Error(err, fmt.Sprintf("failed to query apprvoed cc for peer %s, continue to approve logic", peer.String()))

		req := resmgmt.LifecycleApproveCCRequest{
			Name:              instance.Spec.ID,
			Version:           instance.Spec.Version,
			PackageID:         packageID,
			Sequence:          instance.Status.Sequence,
			EndorsementPlugin: DefaultEndorsementPlugin,
			ValidationPlugin:  DefaultValidationPlugin,
			InitRequired:      instance.Spec.InitRequired,
			SignaturePolicy:   signedPolicy,
		}

		if _, err = pc.LifecycleApproveCC(instance.Spec.Channel, req,
			resmgmt.WithTargetEndpoints(peer.String()), resmgmt.WithOrdererEndpoint(selectOne)); err != nil {
			finalErr = err
			log.Error(err, fmt.Sprintf("%s failed to approve chaincode %s with req %+v\n", method, instance.GetName(), req))
			buf.WriteString(fmt.Sprintf(approveOutputTemplate, peer.Namespace, peer.Name, Failed))
			continue
		}
		log.Info(fmt.Sprintf("%s org %s peer %s approved", method, orgName, peer.String()))
		buf.WriteString(fmt.Sprintf(approveOutputTemplate, peer.Namespace, peer.Name, Approved))
	}

	return buf.String(), finalErr
}
