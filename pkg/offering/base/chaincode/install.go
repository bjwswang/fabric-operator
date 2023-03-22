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
	"github.com/IBM-Blockchain/fabric-operator/pkg/connector"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	lcpackager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/lifecycle"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"k8s.io/apimachinery/pkg/types"
)

const (
	Installed   = "Installed"
	Failed      = "Failed"
	QueryFailed = "QueryFailed"
	Approved    = "Approved"
	Committed   = "Committed"

	installOutputTemplate = `org[%s] peer [%s] installation status %s.`
)

func (c *baseChaincode) InstallChaincode(instance *current.Chaincode) (string, error) {
	method := fmt.Sprintf(" %s [base.chaincode.InstallChaincode]", stepPrefix)
	packagePath := fmt.Sprintf("%s/%s", ChaincodeStorageDir("", instance), ChaincodePacakgeFile(instance))
	log.Info(fmt.Sprintf("%s full packagePath %s", method, packagePath))

	ch := &current.Channel{}
	if err := c.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Channel}, ch); err != nil {
		log.Error(err, "")
		return err.Error(), err
	}

	connectProfile, err := connector.ChannelProfile(c.client, instance.Spec.Channel)
	if err != nil {
		log.Error(err, "")
		return err.Error(), err
	}

	info := connectProfile.GetChannel(instance.Spec.Channel)
	peerOrgAdmin := make(map[string]string)
	log.Info(fmt.Sprintf("%s configure the endpoint of all peer nodes and the admin of their group", method))

	// TODO: 这里的peer节点，后面需要改进，在chaincode里写明那些Peer安装，然后这里需要做过滤。目前先按照channel里所有 Join的peer进行安装.
	for _, peer := range ch.Status.PeerConditions {
		if peer.Type != current.PeerJoined {
			log.Info(fmt.Sprintf("%s peer node %s does not join channel %s",
				method, peer.Name, instance.Spec.Channel))
			continue
		}

		info.Peers[peer.String()] = *connector.DefaultPeerInfo()
		log.Info(fmt.Sprintf("%s set peer node %s connection profile", method, peer.Name))
		if err = connectProfile.SetPeer(c.client, peer.NamespacedName); err != nil {
			log.Info(fmt.Sprintf("%s set peer node %s info error", method, peer.Name))
			log.Error(err, "")
			return err.Error(), err
		}

		if _, ok := peerOrgAdmin[peer.String()]; ok {
			log.Info(fmt.Sprintf("%s peer node %s has already been processed and skipped", method, peer.Name))
			continue
		}
		org := &current.Organization{}
		if err = c.client.Get(context.TODO(), types.NamespacedName{Name: peer.Namespace}, org); err != nil {
			log.Error(err, fmt.Sprintf("get org %s error", peer.Namespace))
			return err.Error(), err
		}
		log.Info(fmt.Sprintf("%s set peer node %s's org admin %s", method, peer.Name, org.Spec.Admin))
		peerOrgAdmin[peer.String()] = org.Spec.Admin
	}

	connectProfile.Channels[instance.Spec.Channel] = info

	chaincodeBytes, err := os.ReadFile(packagePath)
	if err != nil {
		log.Error(err, "")
		return "can't read package file", err
	}
	packageID := lcpackager.ComputePackageID(instance.Spec.Label, chaincodeBytes)
	log.Info(fmt.Sprintf("%s chaincode package id %s", method, packageID))

	peerConnector, err := NewChaincodeConnector(connectProfile)
	if err != nil {
		log.Error(err, fmt.Sprintf("%s chaincode get new connector error", method))
		return err.Error(), err
	}
	defer peerConnector.Close()

	buf := strings.Builder{}
	var finalErr error
	for _, peer := range ch.Status.PeerConditions {
		if peer.Type != current.PeerJoined {
			log.Info(fmt.Sprintf("%s peer node %s has not yet joined the channel %s", method, peer.Name, instance.Spec.Channel))
			continue
		}
		orgAdminCtx := peerConnector.SDK().Context(fabsdk.WithUser(peerOrgAdmin[peer.String()]), fabsdk.WithOrg(peer.Namespace))
		pc, err := resmgmt.New(orgAdminCtx)
		if err != nil {
			finalErr = err
			out := fmt.Sprintf(installOutputTemplate, peer.Namespace, peer.Name, "SdkErr")
			log.Error(err, out)
			buf.WriteString(out)
			continue
		}

		log.Info(fmt.Sprintf("%s try to find out if the package %s is installed on the peer node", method, packageID))
		_, err = pc.LifecycleGetInstalledCCPackage(packageID, resmgmt.WithTargetEndpoints(peer.String()))
		if err == nil {
			log.Info(fmt.Sprintf("%s chaincode have been installed on peer %s", method, peer.Name))
			out := fmt.Sprintf(installOutputTemplate, peer.Namespace, peer.Name, Installed)
			buf.WriteString(out)
			continue
		}
		if !strings.Contains(err.Error(), fmt.Sprintf("chaincode install package '%s' not found", packageID)) {
			finalErr = err
			log.Error(err, fmt.Sprintf("%s failed to query installed pakcage %s", method, packageID))
			out := fmt.Sprintf(installOutputTemplate, peer.Namespace, peer.Name, QueryFailed)
			buf.WriteString(out)
			continue
		}

		log.Info(fmt.Sprintf("%s the peer node %s has no chaincode installed, start to install it",
			method, peer.Name))
		req := resmgmt.LifecycleInstallCCRequest{
			Label:   instance.Spec.Label,
			Package: chaincodeBytes,
		}
		if _, err = pc.LifecycleInstallCC(req, resmgmt.WithTargetEndpoints(peer.String()), resmgmt.WithRetry(retry.DefaultResMgmtOpts)); err != nil {
			finalErr = err
			out := fmt.Sprintf(installOutputTemplate, peer.Namespace, peer.Name, Failed)
			log.Error(err, fmt.Sprintf("%s failed to intall cc with req %+v", method, req))
			buf.WriteString(out)
			continue
		}
		buf.WriteString(fmt.Sprintf(installOutputTemplate, peer.Namespace, peer.Name, Installed))
	}

	return buf.String(), finalErr
}
