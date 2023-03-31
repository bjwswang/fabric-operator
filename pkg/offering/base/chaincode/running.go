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
	"regexp"
	"strings"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/pkg/connector"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	lcpackager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/lifecycle"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var mangledRegExp = regexp.MustCompile("[^a-zA-Z0-9-_.]")

const (
	peerPodSelector = `app.kubernetes.io/name=fabric,app.kubernetes.io/component=chaincode,app.kubernetes.io/created-by=fabric-builder-k8s,app.kubernetes.io/managed-by=fabric-builder-k8s,fabric-builder-k8s-mspid=%s,fabric-builder-k8s-peerid=%s`
	// lable.sha256
	chaincodeIDAnnotations = "fabric-builder-k8s-ccid"

	runningOutputTemplate = "peer [%s] chaincode is Running: %c"
)

func PodName(mspID, peerID, chaincodeID string) string {
	baseName := fmt.Sprintf("cc-%s-%s%s", mspID, peerID, chaincodeID)
	return strings.ToLower(mangledRegExp.ReplaceAllString(baseName, "-")[:63])
}

func (c *baseChaincode) RunningChecker(instance *current.Chaincode) (string, error) {
	method := fmt.Sprintf("%s [base.chaincode.RunningChecker]", stepPrefix)

	packagePath := fmt.Sprintf("%s/%s", ChaincodeStorageDir("", instance), ChaincodePacakgeFile(instance))
	chaincodeBytes, err := os.ReadFile(packagePath)
	if err != nil {
		log.Error(err, "")
		return "can't read package file", err
	}
	packageID := lcpackager.ComputePackageID(instance.Spec.Label, chaincodeBytes)
	log.Info(fmt.Sprintf("%s chaincode package id %s", method, packageID))

	ch, err := instance.GetChannel(c.client)
	log.Info(fmt.Sprintf("%s get channel %s info", method, instance.Spec.Channel))
	if err != nil {
		log.Error(err, "")
		return err.Error(), err
	}

	connectProfile, err := connector.ChannelProfile(c.client, ch.GetName())
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

	lcd, err := pc.LifecycleQueryCommittedCC(ch.GetChannelID(),
		resmgmt.LifecycleQueryCommittedCCRequest{Name: instance.Spec.ID}, resmgmt.WithTargetEndpoints(peer.String()))
	if err != nil {
		log.Error(err, "")
		return err.Error(), err
	}

	found := false
	var resp resmgmt.LifecycleChaincodeDefinition
	for _, item := range lcd {
		if item.Name == instance.Spec.ID {
			found = true
			resp = item
			log.Info(fmt.Sprintf("%s chaincode %s has been committed", http.MethodHead, instance.GetName()))
		}
	}
	if !found {
		log.Info(fmt.Sprintf("%s chaincode %s don't approve", method, instance.GetName()))
		return fmt.Sprintf("chaincode %s have not been approved", instance.GetName()), fmt.Errorf("chaincode %s have not been approved", instance.GetName())
	}

	for _, p := range ch.Status.PeerConditions {
		if p.Type == current.PeerJoined && resp.Approvals[p.Namespace] {
			name := PodName(peer.Namespace, peer.Name, packageID)
			pod := &v1.Pod{}
			if err = c.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: peer.Namespace}, pod); err != nil {
				log.Error(err, "")
				return err.Error(), err
			}
			if pod.Status.Phase != v1.PodRunning {
				err = fmt.Errorf("chaincode is not running on this peer node %s", peer.Name)
				return err.Error(), err
			}
		}
	}
	return "", nil
}
