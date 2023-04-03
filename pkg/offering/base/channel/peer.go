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

package channel

import (
	"context"
	"fmt"
	"strings"
	"time"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/pkg/connector"
	"github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	errPeerAlreadyJoined = errors.New("already exists")
)

const (
	// pollDuration in each poll
	pollDuration = 5 * time.Second
	// pollTimeout in WaitPeer
	pollTimeout = 10 * pollDuration
)

// ReconcilePeer called when peer joines/leaves
func (baseChan *BaseChannel) ReconcilePeer(instance *current.Channel, peer current.NamespacedName) error {
	var err error

	err = baseChan.CheckPeer(peer)
	if err != nil {
		return errors.Wrap(err, "check peer")
	}
	index, condition := instance.GetPeerCondition(peer)
	if condition.Type == current.PeerJoined {
		return nil
	}

	err = baseChan.JoinChannel(instance.GetName(), instance.GetChannelID(), peer)
	if err != nil && !strings.Contains(err.Error(), errPeerAlreadyJoined.Error()) {
		log.Error(err, "failed to reconcile peer", "peer", peer.String())
		condition.Type = current.PeerError
		condition.Status = v1.ConditionTrue
		condition.Reason = err.Error()
		condition.LastTransitionTime = v1.Now()
	} else {
		condition.Type = current.PeerJoined
		condition.Status = v1.ConditionTrue
		condition.Reason = string(current.PeerJoined)
		condition.LastTransitionTime = v1.Now()
	}

	if index != -1 {
		instance.Status.PeerConditions[index] = condition
	} else {
		instance.Status.PeerConditions = append(instance.Status.PeerConditions, condition)
	}
	err = baseChan.Client.PatchStatus(context.TODO(), instance, nil, controllerclient.PatchOption{
		Resilient: &controllerclient.ResilientPatch{
			Retry:    2,
			Into:     &current.Channel{},
			Strategy: client.MergeFrom,
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to patch channel status")
	}

	return nil
}

// CheckPeer make sure peer is at good status
func (baseChan *BaseChannel) CheckPeer(peer current.NamespacedName) error {
	var err error
	peerDeploy := appsv1.Deployment{}
	err = wait.Poll(pollDuration, pollTimeout, func() (bool, error) {
		log.Info(fmt.Sprintf("CheckPeer: poll deployment %s status", peer.String()))
		err := baseChan.Client.Get(context.TODO(), types.NamespacedName{Namespace: peer.Namespace, Name: peer.Name}, &peerDeploy)
		if err != nil {
			if !k8serrors.IsNotFound(err) {
				return false, err
			}
			return false, nil
		}

		if peerDeploy.Status.AvailableReplicas != *peerDeploy.Spec.Replicas {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return errors.Errorf("exceed the poll timeout of %f seconds", pollTimeout.Seconds())
	}
	return nil
}

// JoinChannel calls peer api to join it into a existing channel
func (baseChan *BaseChannel) JoinChannel(channelName, channelID string, peer current.NamespacedName) error {
	c, err := connector.NewConnector(baseChan.ConnectorProfile(channelName, channelID, peer))
	if err != nil {
		return err
	}

	peerOrg := peer.Namespace
	organization := &current.Organization{}
	err = baseChan.Client.Get(context.TODO(), types.NamespacedName{Name: peerOrg}, organization)
	if err != nil {
		return err
	}
	adminContext := c.SDK().Context(fabsdk.WithUser(organization.Spec.Admin), fabsdk.WithOrg(peerOrg))
	client, err := resmgmt.New(adminContext)
	if err != nil {
		return err
	}
	err = client.JoinChannel(channelID, resmgmt.WithTargetEndpoints(peer.String()))
	if err != nil {
		return errors.Wrap(err, "failed to join peer into channel")
	}
	return nil
}

// ConnectorProfile customizes channel connection profile with peer info
func (baseChan *BaseChannel) ConnectorProfile(channelName, channelID string, peer current.NamespacedName) connector.ProfileFunc {
	return func() ([]byte, error) {
		var err error

		cm := &corev1.ConfigMap{}
		err = baseChan.Client.Get(context.TODO(), types.NamespacedName{Namespace: peer.Namespace, Name: fmt.Sprintf("chan-%s-connection-profile", channelName)}, cm)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get channel connection profile")
		}
		profile := &connector.Profile{}
		err = profile.Unmarshal(cm.BinaryData["profile.yaml"], connector.YAML)
		if err != nil {
			return nil, errors.Wrap(err, "invalid channel connection profile")
		}

		profile.Client.Organization = peer.Namespace
		// add peer under channel section
		info := profile.GetChannel(channelID)
		info.Peers[peer.String()] = *connector.DefaultPeerInfo()
		profile.Channels[channelID] = info
		// add peer under peers&organization
		err = profile.SetPeer(baseChan.Client, current.NamespacedName{Name: peer.Name, Namespace: peer.Namespace})
		if err != nil {
			return nil, errors.Wrap(err, "failed to add current peer into connection profile")
		}

		return profile.Marshal(connector.YAML)
	}
}
