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
	"reflect"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	k8sclient "github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/go-test/deep"
	"gopkg.in/yaml.v2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func (r *ReconcileChannel) CreateFunc(e event.CreateEvent) bool {
	channel := e.Object.(*current.Channel)
	log.Info(fmt.Sprintf("Create event detected for channel '%s'", channel.GetName()))

	update := Update{}

	if channel.HasType() {
		log.Info(fmt.Sprintf("Operator restart detected, running update flow on existing channel '%s'", channel.GetName()))

		// Get the spec state of the resource before the operator went down, this
		// will be used to compare to see if the spec of resources has changed
		cm, err := r.GetSpecState(channel)
		if err != nil {
			log.Info(fmt.Sprintf("Failed getting saved channel spec '%s', triggering create: %s", channel.GetName(), err.Error()))
			return true
		}

		specBytes := cm.BinaryData["spec"]
		existingChannel := &current.Channel{}
		err = yaml.Unmarshal(specBytes, &existingChannel.Spec)
		if err != nil {
			log.Info(fmt.Sprintf("Unmarshal failed for saved channel spec '%s', triggering create: %s", channel.GetName(), err.Error()))
			return true
		}

		diff := deep.Equal(channel.Spec, existingChannel.Spec)
		if diff != nil {
			log.Info(fmt.Sprintf("Channel '%s' spec was updated while operator was down", channel.GetName()))
			log.Info(fmt.Sprintf("Difference detected: %v", diff))
			update.specUpdated = true
		}

		added, removed := current.DifferMembers(existingChannel.Spec.Members, channel.Spec.Members)
		if len(added) != 0 || len(removed) != 0 {
			if len(removed) != 0 {
				// TODO: support deleting members from channel later
				log.Error(fmt.Errorf("deleting members from a channel is not yet supported"), "Deleting members from a channel is not yet supported.", "channel", channel.GetName())
				return false
			}
			log.Info(fmt.Sprintf("Channel '%s' members was updated while operator was down", channel.GetName()))
			log.Info(fmt.Sprintf("Difference detected: added members %v", added))
			log.Info(fmt.Sprintf("Difference detected: removed members %v", removed))
			update.memberUpdated = true
		}

		addedPeers, removedPeers := current.DifferChannelPeers(existingChannel.Spec.Peers, channel.Spec.Peers)
		if len(addedPeers) != 0 || len(removedPeers) != 0 {
			log.Info(fmt.Sprintf("Channel '%s' peers was updated while operator was down", channel.GetName()))
			log.Info(fmt.Sprintf("Difference detected: added peers %v", addedPeers))
			log.Info(fmt.Sprintf("Difference detected: removed peers %v", removedPeers))
			update.peerUpdated = true
		}

		log.Info(fmt.Sprintf("Create event triggering reconcile for updating Channel '%s'", channel.GetName()))
		r.PushUpdate(channel.GetName(), update)
		return true
	}

	if len(channel.Spec.Peers) != 0 {
		update.peerUpdated = true
	}

	update.specUpdated = true
	update.memberUpdated = true
	r.PushUpdate(channel.GetName(), update)

	return true
}

func (r *ReconcileChannel) UpdateFunc(e event.UpdateEvent) bool {
	oldChan := e.ObjectOld.(*current.Channel)
	newChan := e.ObjectNew.(*current.Channel)
	log.Info(fmt.Sprintf("Update event detected for channel '%s'", oldChan.GetName()))

	update := Update{}

	if reflect.DeepEqual(oldChan.Spec, newChan.Spec) {
		return false
	}

	update.specUpdated = true

	added, removed := current.DifferMembers(oldChan.GetMembers(), newChan.GetMembers())
	if len(added) != 0 || len(removed) != 0 {
		if len(removed) != 0 {
			log.Error(fmt.Errorf("deleting members from a channel is not yet supported"), "Deleting members from a channel is not yet supported.", "channel", newChan.GetName())
			return false
		}
		log.Info(fmt.Sprintf("Difference detected: added members %v", added))
		log.Info(fmt.Sprintf("Difference detected: removed members %v", removed))
		update.memberUpdated = true
	}

	addedPeers, removedPeers := current.DifferChannelPeers(oldChan.Spec.Peers, newChan.Spec.Peers)
	if len(addedPeers) != 0 || len(removedPeers) != 0 {
		log.Info(fmt.Sprintf("Difference detected: added peers %v", addedPeers))
		log.Info(fmt.Sprintf("Difference detected: removed peers %v", removedPeers))
		update.peerUpdated = true
	}

	r.PushUpdate(oldChan.GetName(), update)

	log.Info(fmt.Sprintf("Spec update triggering reconcile on Channel custom resource %s: update [ %+v ]", oldChan.Name, update.GetUpdateStackWithTrues()))

	return true
}

func (r *ReconcileChannel) ProposalUpdateFunc(e event.UpdateEvent) bool {
	var err error

	oldProposal := e.ObjectOld.(*current.Proposal)
	newProposal := e.ObjectNew.(*current.Proposal)
	log.Info(fmt.Sprintf("Update event detected for proposal '%s'", oldProposal.GetName()))

	if reflect.DeepEqual(oldProposal.Spec, newProposal.Spec) && reflect.DeepEqual(oldProposal.Status, newProposal.Status) {
		return false
	}

	targetChannel := ""
	if newProposal.Status.Phase == current.ProposalFinished {
		for _, c := range newProposal.Status.Conditions {
			switch c.Type {
			case current.ProposalSucceeded:
				switch newProposal.GetPurpose() {
				case current.ArchiveChannelProposal:
					targetChannel = newProposal.Spec.ArchiveChannel.Channel
					err = r.PatchProposalStatus(targetChannel, newProposal.GetName(), current.ArchiveChannelProposal)
					if err != nil {
						log.Error(err, "patch channel status with proposal %s succ")
					}
				case current.UnarchiveChannelProposal:
					targetChannel = newProposal.Spec.UnarchiveChannel.Channel
					err = r.PatchProposalStatus(targetChannel, newProposal.GetName(), current.UnarchiveChannelProposal)
					if err != nil {
						log.Error(err, "patch channel status by proposal succ", "proposal", newProposal.GetName())
					}
				case current.UpdateChannelMemberProposal:
					targetChannel = newProposal.Spec.UpdateChannelMember.Channel
					ch := &current.Channel{}
					if err := r.client.Get(context.TODO(), types.NamespacedName{Name: targetChannel}, ch); err != nil {
						log.Error(err, "get channel error", "proposal", newProposal.GetName())
						return false
					}
					for i, m := range newProposal.Spec.UpdateChannelMember.Members {
						now := metav1.Now()
						m.JoinedAt = &now
						m.JoinedBy = newProposal.GetName()
						newProposal.Spec.UpdateChannelMember.Members[i] = m
					}
					ch.Spec.Members = append(ch.Spec.Members, newProposal.Spec.UpdateChannelMember.Members...)
					if err = r.client.Update(context.TODO(), ch); err != nil {
						log.Error(err, "update channel memeber error", "proposal", newProposal.GetName())
					}
				default:
					return false
				}
			}
		}
	}

	return false
}

func (r *ReconcileChannel) PatchProposalStatus(targetChannel string, proposal string, purpose uint) error {
	ch := &current.Channel{}
	ch.Name = targetChannel
	err := r.client.Get(context.TODO(), client.ObjectKeyFromObject(ch), ch)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	switch purpose {
	case current.ArchiveChannelProposal:
		ch.Status.ArchivedStatus = ch.Status.CRStatus
		ch.Status.CRStatus = current.CRStatus{
			Type:              current.ChannelArchived,
			Status:            current.True,
			Reason:            "archived",
			Message:           fmt.Sprintf("channel archived by proposal %s", proposal),
			LastHeartbeatTime: metav1.Now(),
		}
	case current.UnarchiveChannelProposal:
		ch.Status.CRStatus = ch.Status.ArchivedStatus
		ch.Status.ArchivedStatus = current.CRStatus{}
	}

	err = r.client.PatchStatus(context.TODO(), ch, nil, controllerclient.PatchOption{
		Resilient: &k8sclient.ResilientPatch{
			Retry:    2,
			Into:     &current.Channel{},
			Strategy: client.MergeFrom,
		},
	})

	if err != nil {
		return err
	}

	return nil
}
