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

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/pkg/connector"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"k8s.io/apimachinery/pkg/types"
)

const (
	commitOutputTemplate = `org[%s] peer [%s] commit status %s.`
)

func (c *baseChaincode) CommitChaincode(instance *current.Chaincode) (string, error) {
	method := fmt.Sprintf("%s [base.chaincode.CommitChaincode]", stepPrefix)

	signedPoilciy, err := ChaincodeEndorsePolicy(c.client, instance)
	if err != nil {
		return err.Error(), err
	}

	ch := &current.Channel{}
	log.Info(fmt.Sprintf("%s get channel %s info", method, instance.Spec.Channel))
	if err := c.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Channel}, ch); err != nil {
		log.Error(err, "")
		return err.Error(), err
	}

	connectProfile, err := connector.ChannelProfile(c.client, instance.Spec.Channel)
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
		log.Error(err, "")
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
		log.Info(fmt.Sprintf("%s chaincode get new connector error %s\n", method, err))
		log.Error(err, "")
		return "chaincode get new connector error", err
	}
	defer peerConnector.Close()

	var (
		pc   *resmgmt.Client
		peer current.NamespacedName
	)

	targetPoints := make([]string, 0)
	initClient := true
	for org, p := range orgPeer {
		peer = current.NamespacedName{Name: p.GetName(), Namespace: p.GetNamespace()}
		if initClient {
			log.Info(fmt.Sprintf("%s pick org %s, peer %v", method, org, peer))
			orgAdminCtx := peerConnector.SDK().Context(fabsdk.WithUser(peerAdmin[peer.String()]), fabsdk.WithOrg(org))
			pc, err = resmgmt.New(orgAdminCtx)
			if err != nil {
				out := fmt.Sprintf(commitOutputTemplate, peer.Namespace, peer.Name, "SdkErr")
				log.Error(err, out)
				return err.Error(), err
			}
			initClient = false
		}
		targetPoints = append(targetPoints, peer.String())
	}

	ccReadinessReq := resmgmt.LifecycleCheckCCCommitReadinessRequest{
		Name:              instance.Spec.ID,
		Version:           instance.Spec.Version,
		Sequence:          instance.Status.Sequence,
		EndorsementPlugin: DefaultEndorsementPlugin,
		ValidationPlugin:  DefaultValidationPlugin,
		SignaturePolicy:   signedPoilciy,
		InitRequired:      instance.Spec.InitRequired,
	}
	resp, _ := pc.LifecycleCheckCCCommitReadiness(instance.Spec.Channel, ccReadinessReq, resmgmt.WithTargetEndpoints(peer.String()))
	if len(resp.Approvals) == 0 {
		return "there are no organizations with approved chaincode", fmt.Errorf("there are no organizations with approved chaincode")
	}

	// TODO: 目前默认的策略都是 Majority。
	mid := len(ch.Spec.Members)/2 + 1
	approve := 0
	for _, v := range resp.Approvals {
		if v {
			approve++
		}
	}
	if approve < mid {
		out := fmt.Sprintf("don't match to majority endorsement policy, need %d, approved %d", mid, approve)
		log.Info(fmt.Sprintf("%s %s, mismatch approvals: %+v", method, out, resp.Approvals))
		return out, fmt.Errorf(out)
	}
	log.Info(fmt.Sprintf("%s commitReadiness check successful, start to commit", method))

	log.Info(fmt.Sprintf("%s try to find out if the package committed", method))
	lcd, err := pc.LifecycleQueryCommittedCC(instance.Spec.Channel,
		resmgmt.LifecycleQueryCommittedCCRequest{Name: instance.Spec.ID}, resmgmt.WithTargetEndpoints(peer.String()))
	if err != nil {
		log.Error(err, "")
	}
	for _, item := range lcd {
		if item.Name == instance.Spec.ID {
			log.Info(fmt.Sprintf("%s chaincode %s has been committed", http.MethodHead, instance.GetName()))
			return "", nil
		}
	}

	log.Error(err, "failed to query committed chaincode, continue to commit logic")

	req := resmgmt.LifecycleCommitCCRequest{
		Name:              instance.Spec.ID,
		Version:           instance.Spec.Version,
		Sequence:          instance.Status.Sequence,
		EndorsementPlugin: DefaultEndorsementPlugin,
		ValidationPlugin:  DefaultValidationPlugin,
		SignaturePolicy:   signedPoilciy,
		InitRequired:      instance.Spec.InitRequired,
	}
	if _, err = pc.LifecycleCommitCC(instance.Spec.Channel, req,
		resmgmt.WithOrdererEndpoint(selectOne),
		resmgmt.WithTargetEndpoints(targetPoints...)); err != nil {
		log.Error(err, fmt.Sprintf("%s failed to committed chaincode %s with req %+v\n", method, instance.GetName(), req))
		return err.Error(), err
	}

	return "", nil
}
