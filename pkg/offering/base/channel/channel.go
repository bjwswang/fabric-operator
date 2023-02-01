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

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	config "github.com/IBM-Blockchain/fabric-operator/operatorconfig"
	"github.com/IBM-Blockchain/fabric-operator/pkg/connector"
	chaninit "github.com/IBM-Blockchain/fabric-operator/pkg/initializer/channel"
	"github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/common"
	bcrbac "github.com/IBM-Blockchain/fabric-operator/pkg/rbac"
	"github.com/IBM-Blockchain/fabric-operator/version"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("base_channel")

//go:generate counterfeiter -o mocks/update.go -fake-name Update . Update

type Update interface {
	SpecUpdated() bool
	MemberUpdated() bool
	NetworkUpdated() bool
	PeerUpdated() bool
}

//go:generate counterfeiter -o mocks/override.go -fake-name Override . Override

type Override interface{}

//go:generate counterfeiter -o mocks/basechannel.go -fake-name Channel . Channel

type Channel interface {
	PreReconcileChecks(instance *current.Channel, update Update) error
	Initialize(instance *current.Channel, update Update) error
	ReconcileManagers(instance *current.Channel, update Update) error
	CheckStates(instance *current.Channel, update Update) (common.Result, error)
}

var _ Channel = (*BaseChannel)(nil)

const (
	KIND = "CHANNEL"
)

type BaseChannel struct {
	Client controllerclient.Client
	Scheme *runtime.Scheme

	Config *config.Config

	Override Override

	Initializer *chaninit.Initializer

	RBACManager *bcrbac.Manager
}

func New(client controllerclient.Client, scheme *runtime.Scheme, config *config.Config, o Override) *BaseChannel {
	base := &BaseChannel{
		Client:   client,
		Scheme:   scheme,
		Config:   config,
		Override: o,
	}

	base.Initializer = chaninit.New(client, scheme, config.ChannelInitConfig)

	base.CreateManagers()

	return base
}

func (channel *BaseChannel) CreateManagers() {
	channel.RBACManager = bcrbac.NewRBACManager(channel.Client, nil)
}

// PreReconcileChecks on Channel upon Update
func (channel *BaseChannel) PreReconcileChecks(instance *current.Channel, update Update) error {
	var err error
	log.Info(fmt.Sprintf("PreReconcileChecks on Channel %s", instance.GetName()))

	if !instance.HasNetwork() {
		return errors.New("channel's network is empty")
	}

	if !instance.HashMembers() {
		return errors.New("channel has no members")
	}

	// make sure channel members is the subset of network's members
	network := &current.Network{}
	err = channel.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Network}, network)
	if err != nil {
		return errors.Wrap(err, "get channel's network")
	}
	if network.Status.Type != current.Created {
		return errors.Errorf("network %s not created yet", network.GetName())
	}
	members := make(map[string]struct{})
	for _, m := range network.GetMembers() {
		members[m.Name] = struct{}{}
	}
	for _, m := range instance.GetMembers() {
		_, ok := members[m.Name]
		if !ok {
			return errors.Errorf("channel member %s not a network member", m.Name)
		}
	}

	return nil
}

// Initialize on Channel upon Update
func (baseChan *BaseChannel) Initialize(instance *current.Channel, update Update) error {
	err := baseChan.Initializer.CreateChannel(instance)
	if err != nil {
		return err
	}

	// Patch status

	return nil
}

// ReconcileManagers on Channel upon Update
func (baseChan *BaseChannel) ReconcileManagers(instance *current.Channel, update Update) error {
	var err error

	// set channel's owner reference to its network
	err = baseChan.SetOwnerReference(instance, update)
	if err != nil {
		return err
	}

	// reconcile channel member's rbac
	err = baseChan.ReconcileRBAC(instance, update)
	if err != nil {
		return err
	}

	// Channel changed or network changed or peer updated
	if update.SpecUpdated() || update.NetworkUpdated() || update.PeerUpdated() {
		err = baseChan.ReconcileConnectionProfile(instance, update)
		if err != nil {
			return err
		}
	}

	if update.PeerUpdated() {
		for _, p := range instance.Spec.Peers {
			err = baseChan.ReconcilePeer(instance, p)
			if err != nil {
				return errors.Wrap(err, "failed to patch channel status")
			}
		}
	}

	return nil
}

// CheckStates on Channel(do nothing)
func (baseChan *BaseChannel) CheckStates(instance *current.Channel, update Update) (common.Result, error) {
	return common.Result{
		Status: &current.CRStatus{
			Type:    current.ChannelCreated,
			Version: version.Operator,
		},
	}, nil
}

func (baseChan *BaseChannel) SetOwnerReference(instance *current.Channel, update Update) error {
	var err error

	network := &current.Network{}
	err = baseChan.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Network}, network)
	if err != nil {
		return errors.Wrap(err, "get channel's network")
	}
	ownerReference := bcrbac.OwnerReference(bcrbac.Network, network)

	var exist bool
	for _, reference := range instance.OwnerReferences {
		if reference.UID == ownerReference.UID {
			exist = true
			break
		}
	}
	if !exist {
		instance.OwnerReferences = []v1.OwnerReference{bcrbac.OwnerReference(bcrbac.Network, network)}

		err = baseChan.Client.Update(context.TODO(), instance)
		if err != nil {
			return err
		}
	}

	return nil
}

// ReconcileRBAC will sync current channel to every member's AdminClusterRole
func (baseChan *BaseChannel) ReconcileRBAC(instance *current.Channel, update Update) error {
	if update.MemberUpdated() && baseChan.Config.OrganizationInitConfig.IAMEnabled {
		err := baseChan.RBACManager.Reconcile(bcrbac.Channel, instance, bcrbac.ResourceUpdate)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReconcileConnectionProfile generates connection profile for this channel
func (baseChan *BaseChannel) ReconcileConnectionProfile(instance *current.Channel, update Update) error {
	var err error

	network := &current.Network{}
	err = baseChan.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Network}, network)
	if err != nil {
		return err
	}
	ordererorg := network.Labels["bestchains.network.initiator"]
	clusterNodes, err := baseChan.Initializer.GetClusterNodes(ordererorg, network.GetName())
	if err != nil {
		return err
	}
	profile, err := baseChan.GenerateChannelConnProfile("", instance, clusterNodes)
	if err != nil {
		return err
	}
	binaryData, err := profile.Marshal()
	if err != nil {
		return err
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      instance.GetConnectionPorfile(),
			Namespace: baseChan.Config.Operator.Namespace,
		},
		BinaryData: map[string][]byte{
			"profile.yaml": binaryData,
		},
	}
	err = baseChan.Client.CreateOrUpdate(context.TODO(), cm, controllerclient.CreateOrUpdateOption{
		Owner:  instance,
		Scheme: baseChan.Scheme,
	})
	if err != nil {
		return err
	}
	return nil
}

func (baseChan *BaseChannel) GenerateChannelConnProfile(clientOrg string, channel *current.Channel, clusterNodes *current.IBPOrdererList) (*connector.Profile, error) {
	var err error

	basedir := baseChan.Initializer.GetStoragePath(channel)

	var orgs = make([]string, len(channel.Spec.Members))
	for index, m := range channel.Spec.Members {
		orgs[index] = m.GetName()
	}
	if clientOrg == "" && len(orgs) > 0 {
		clientOrg = orgs[0]
	}

	// default connprofile with default client
	profile := connector.DefaultProfile(basedir, clientOrg)

	// Channel
	profile.SetChannel(channel.GetChannelID(), channel.Spec.Peers...)

	// Peers
	peers := make(map[string][]string)
	for _, p := range channel.Status.PeerConditions {
		// only joined peer can be appended into connection profile
		if p.Type != current.PeerJoined {
			continue
		}
		err = profile.SetPeer(baseChan.Client, p.NamespacedName)
		if err != nil {
			return nil, err
		}
		// cache to peers
		_, ok := peers[p.Namespace]
		if !ok {
			peers[p.Namespace] = make([]string, 0)
		}
		peers[p.Namespace] = append(peers[p.Namespace], p.String())
	}

	// Orderers
	for _, o := range clusterNodes.Items {
		err = profile.SetOrderer(baseChan.Client, current.NamespacedName{Namespace: o.GetNamespace(), Name: o.GetName()})
		if err != nil {
			return nil, err
		}
	}

	// Organizations
	for _, org := range orgs {
		organization := &current.Organization{}
		err = baseChan.Client.Get(context.TODO(), types.NamespacedName{Name: org}, organization)
		if err != nil {
			return nil, errors.Wrap(err, "failed to find organization")
		}
		// read organization admin's secret
		orgMSPSecret := &corev1.Secret{}
		err = baseChan.Client.Get(context.TODO(), types.NamespacedName{Namespace: org, Name: fmt.Sprintf("%s-msp-crypto", org)}, orgMSPSecret)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get channel connection profile")
		}
		adminUser := connector.User{
			Name: organization.Spec.Admin,
			Key: connector.Pem{
				Pem: string(orgMSPSecret.Data["admin-keystore"]),
			},
			Cert: connector.Pem{
				Pem: string(orgMSPSecret.Data["admin-signcert"]),
			},
		}
		profile.SetOrganization(org, peers[org], adminUser)
	}

	return profile, nil
}

// GetLabels from instance.GetLabels
func (baseChan *BaseChannel) GetLabels(instance v1.Object) map[string]string {
	return instance.GetLabels()
}
