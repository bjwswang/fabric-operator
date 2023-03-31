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

	"github.com/pkg/errors"
	authenticationv1 "k8s.io/api/authentication/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	errNoNetwork            = errors.New("cant find network")
	errMemberNotInNetwork   = errors.New("some channel member not in network")
	errNoPermOperatePeer    = errors.New("no permission to operate peer")
	errUpdateChannelNetwork = errors.New("cant update channel's network")
	errUpdateChannelID      = errors.New("can not update channels' id")
	errUpdateChannelMember  = errors.New("cant update channel's members directly(must use proposal-vote)")
	errChannelHasPeers      = errors.New("channel still have peers joined")
)

// log is for logging in this package.
var channellog = logf.Log.WithName("channel-resource")

//+kubebuilder:webhook:path=/mutate-ibp-com-v1beta1-channel,mutating=true,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=channels,verbs=create;update,versions=v1beta1,name=channel.mutate.webhook,admissionReviewVersions=v1

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Channel) Default(ctx context.Context, client client.Client, user authenticationv1.UserInfo) {
	channellog.Info("default", "name", r.Name, "user", user.String())

	for index := range r.Spec.Members {
		if r.Spec.Members[index].JoinedAt != nil {
			continue
		}
		now := metav1.Now()
		r.Spec.Members[index].JoinedAt = &now
	}
}

//+kubebuilder:webhook:path=/validate-ibp-com-v1beta1-channel,mutating=false,failurePolicy=fail,sideEffects=None,groups=ibp.com,resources=channels,verbs=create;update;delete,versions=v1beta1,name=channel.validate.webhook,admissionReviewVersions=v1

var _ validator = &Channel{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Channel) ValidateCreate(ctx context.Context, c client.Client, user authenticationv1.UserInfo) error {
	channellog.Info("validate create", "name", r.Name, "user", user.String())

	err := validateMemberInNetwork(ctx, c, r.Spec.Network, r.Spec.Members)
	if err != nil {
		return err
	}

	// managedOrgs which this user can manage
	managedOrgs, err := filterManagedOrgs(ctx, c, user, r.Spec.Members)
	if err != nil {
		return err
	}
	// initialized peers should under user's management
	err = validatePeersOwnership(ctx, c, managedOrgs, r.Spec.Peers)
	if err != nil {
		return err
	}
	return validateChannelID(ctx, c, r.Spec.Network, r.Spec.ID)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Channel) ValidateUpdate(ctx context.Context, c client.Client, old runtime.Object, user authenticationv1.UserInfo) error {
	channellog.Info("validate update", "name", r.Name, "user", user.String())

	oldChannel := old.(*Channel)

	// forbid to udpate channel network
	if oldChannel.Spec.Network != r.Spec.Network {
		return errUpdateChannelNetwork
	}
	if oldChannel.Spec.ID != r.Spec.ID {
		return errUpdateChannelID
	}

	// forbid to update channel members directly
	added, removed := DifferMembers(oldChannel.Spec.Members, r.Spec.Members)
	if len(added) != 0 || len(removed) != 0 {
		if !isSuperUser(ctx, user) {
			return errUpdateChannelMember
		}
	}

	// forbid to update peers which not belongs to user's organizations
	addedPeers, removedPeers := DifferChannelPeers(oldChannel.Spec.Peers, r.Spec.Peers)
	if len(addedPeers) != 0 || len(removedPeers) != 0 {
		managedOrgs, err := filterManagedOrgs(ctx, c, user, r.Spec.Members)
		if err != nil {
			return err
		}
		// updated peers should under user's management
		err = validatePeersOwnership(ctx, c, managedOrgs, append(addedPeers, removedPeers...))
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Channel) ValidateDelete(ctx context.Context, client client.Client, user authenticationv1.UserInfo) error {
	channellog.Info("validate delete", "name", r.Name, "user", user.String())

	// forbid to delete channel if still have peers joined
	if len(r.Spec.Peers) != 0 {
		return errors.Wrapf(errChannelHasPeers, "count %d", len(r.Spec.Peers))
	}

	return nil
}

func validateMemberInNetwork(ctx context.Context, c client.Client, netName string, members []Member) error {
	net := &Network{}
	net.Name = netName
	if err := c.Get(ctx, client.ObjectKeyFromObject(net), net); err != nil {
		if apierrors.IsNotFound(err) {
			return errNoNetwork
		}
		return errors.Wrap(err, "failed to get network")
	}
	if net.Status.Type == Error {
		return errors.Errorf("network %s has error %s:%s", netName, net.Status.Reason, net.Status.Message)
	}

	allMembers := make(map[string]bool, len(net.Spec.Members))
	for _, m := range net.Spec.Members {
		allMembers[m.Name] = true
	}
	for _, m := range members {
		if ok := allMembers[m.Name]; !ok {
			return errors.Wrapf(errMemberNotInNetwork, "allMembers:%#v, members:%#v", allMembers, members)
		}
	}
	return nil
}

// filterManagedOrgs will get the organizations which under user's management
func filterManagedOrgs(ctx context.Context, c client.Client, user authenticationv1.UserInfo, members []Member) ([]string, error) {
	var err error

	// managedOrgs which this user can manage
	managedOrgs := make([]string, 0)
	// validate ownership
	for _, member := range members {
		org := &Organization{}
		org.Name = member.Name
		err = c.Get(ctx, client.ObjectKeyFromObject(org), org)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get organization")
		}
		if isSuperUser(ctx, user) || org.Spec.Admin == user.Username {
			managedOrgs = append(managedOrgs, member.Name)
		}
	}

	return managedOrgs, nil
}

// validatePeersOwnership validate whether peers belongs to ownerOrgs
func validatePeersOwnership(ctx context.Context, c client.Client, ownerOrgs []string, peers []NamespacedName) error {
	// cache owners
	owners := make(map[string]bool, len(ownerOrgs))
	for _, ownerOrg := range ownerOrgs {
		owners[ownerOrg] = true
	}
	// make sure peers run in ownerOrgs
	for _, peer := range peers {
		// peer must belongs to owners
		if !owners[peer.Namespace] {
			return errors.Wrapf(errNoPermOperatePeer, "peer belongs to %s not in %v", peer.Namespace, ownerOrgs)
		}
		p := &IBPPeer{}
		err := c.Get(ctx, types.NamespacedName{Namespace: peer.Namespace, Name: peer.Name}, p)
		if err != nil {
			return errors.Wrapf(err, "failed to get peer %s", peer.String())
		}

		switch p.Status.Type {
		case Error:
			return errors.Errorf("peer %s has error %s:%s", peer.String(), p.Status.Reason, p.Status.Message)
		case Deploying:
			return errors.Errorf("peer %s still deploying. cannot be used to join channel", peer.String())
		}
	}
	return nil
}

func validateChannelID(ctx context.Context, c client.Client, network, channelID string) error {
	selector := labels.NewSelector()
	req, _ := labels.NewRequirement(CHANNEL_NETWORK_LABEL, selection.In, []string{network})
	selector = selector.Add(*req)
	listOption := client.ListOptions{
		LabelSelector: selector,
	}
	channelList := &ChannelList{}
	if err := c.List(ctx, channelList, &listOption); err != nil {
		return err
	}
	for _, ch := range channelList.Items {
		channellog.Info(fmt.Sprintf("by selector %s, get ch %s", selector.String(), ch.GetName()))
		if ch.Spec.ID == channelID {
			return fmt.Errorf("channel id %s already exists in channel %s", channelID, ch.GetName())
		}
	}
	return nil
}
