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

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/pkg/connector"
	"github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// https://hyperledger-fabric.readthedocs.io/en/release-2.2/pluggable_endorsement_and_validation.html#configuration
	DefaultEndorsementPlugin = "escc"
	DefaultValidationPlugin  = "vscc"
)

func ProfileProvider(cli controllerclient.Client, channelID string) (*connector.Profile, error) {
	var err error
	operatorNamespace := os.Getenv("WATCH_NAMESPACE")
	if operatorNamespace == "" {
		operatorNamespace = os.Getenv("OPERATOR_NAMESPACE")
	}
	cm := &corev1.ConfigMap{}
	err = cli.Get(context.TODO(), types.NamespacedName{
		Namespace: operatorNamespace,
		Name:      fmt.Sprintf("chan-%s-connection-profile", channelID)},
		cm)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get channel connection profile")
	}
	profile := &connector.Profile{}
	err = profile.Unmarshal(cm.BinaryData["profile.yaml"])
	if err != nil {
		return nil, errors.Wrap(err, "invalid channel connection profile")
	}
	return profile, nil
}

func ProfileFn(p *connector.Profile) func() ([]byte, error) {
	return func() ([]byte, error) {
		return p.Marshal()
	}
}

func NewChaincodeConnector(p *connector.Profile) (*connector.Connector, error) {
	p1 := ProfileFn(p)
	return connector.NewConnector(p1)
}

func getOrderNodes(cli controllerclient.Client, namespace, parentNode string) (*current.IBPOrdererList, error) {
	orderList := &current.IBPOrdererList{}
	labelSelector, _ := labels.Parse(fmt.Sprintf("parent=%s", parentNode))
	listOptions := &client.ListOptions{
		LabelSelector: labelSelector,
		Namespace:     namespace,
	}
	err := cli.List(context.TODO(), orderList, listOptions)
	return orderList, err
}

// SetChannelPeerProfile set the peer's connection information and return the peer's organization admin
func SetChannelPeerProfile(cli controllerclient.Client, p *connector.Profile, ch *current.Channel) (map[string]current.IBPPeer, map[string]string, error) {
	orgPeers := make(map[string]current.IBPPeer)
	peerAdmin := make(map[string]string)
	info := p.GetChannel(ch.GetName())
	for _, memberOrg := range ch.GetMembers() {
		org := &current.Organization{}
		if err := cli.Get(context.TODO(), types.NamespacedName{Name: memberOrg.Name}, org); err != nil {
			return nil, nil, err
		}

		peerList := &current.IBPPeerList{}
		if err := cli.List(context.TODO(), peerList, client.InNamespace(memberOrg.Name)); err != nil {
			return nil, nil, err
		}
		if len(peerList.Items) == 0 {
			return nil, nil, fmt.Errorf("org %s don't have any peer", memberOrg.Name)
		}

		firstPeer := peerList.Items[0]
		orgPeers[memberOrg.Name] = firstPeer
		cur := current.NamespacedName{Name: firstPeer.GetName(), Namespace: firstPeer.GetNamespace()}

		info.Peers[cur.String()] = *connector.DefaultPeerInfo()
		if err := p.SetPeer(cli, cur); err != nil {
			return nil, nil, err
		}
		peerAdmin[cur.String()] = org.Spec.Admin
	}

	p.Channels[ch.GetName()] = info
	return orgPeers, peerAdmin, nil
}
