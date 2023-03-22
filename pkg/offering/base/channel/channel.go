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
	"bytes"
	"context"
	"fmt"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	config "github.com/IBM-Blockchain/fabric-operator/operatorconfig"
	"github.com/IBM-Blockchain/fabric-operator/pkg/connector"
	chaninit "github.com/IBM-Blockchain/fabric-operator/pkg/initializer/channel"
	"github.com/IBM-Blockchain/fabric-operator/pkg/initializer/orderer/configtx"
	"github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/common"
	bcrbac "github.com/IBM-Blockchain/fabric-operator/pkg/rbac"
	"github.com/IBM-Blockchain/fabric-operator/pkg/util"
	"github.com/IBM-Blockchain/fabric-operator/version"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-config/protolator"
	proto_common "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric/common/channelconfig"
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

func (baseChan *BaseChannel) CreateManagers() {
	baseChan.RBACManager = bcrbac.NewRBACManager(baseChan.Client, nil)
}

// PreReconcileChecks on Channel upon Update
func (baseChan *BaseChannel) PreReconcileChecks(instance *current.Channel, update Update) error {
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
	err = baseChan.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Network}, network)
	if err != nil {
		return errors.Wrap(err, "get channel's network")
	}
	if network.Status.Type != current.Deployed {
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
	if instance.Status.Type != current.ChannelCreated {
		err := baseChan.Initializer.CreateChannel(instance)
		if err != nil {
			return err
		}
	}

	return nil
}

// ReconcileManagers on Channel upon Update
func (baseChan *BaseChannel) ReconcileManagers(instance *current.Channel, update Update) error {
	var err error

	// set channel's owner reference to its network
	err = baseChan.ReconcileOwnerReference(instance, update)
	if err != nil {
		return err
	}

	// reconcile channel member's rbac
	err = baseChan.ReconcileRBAC(instance, update)
	if err != nil {
		return err
	}

	// member changed
	if update.MemberUpdated() {
		err = baseChan.ReconcileConnectionProfile(instance, update)
		if err != nil {
			return err
		}
		if instance.HasType() {
			// Initializer.CreateChannel done
			err = baseChan.ReconcileChannelMember(instance)
			if err != nil {
				return err
			}
		}
	}

	// Reconcile peer if peer updated
	// - join new peer into channel
	if update.PeerUpdated() {
		for _, p := range instance.Spec.Peers {
			err = baseChan.ReconcilePeer(instance, p)
			if err != nil {
				return errors.Wrap(err, "failed to reconcile channel peer")
			}
		}
		// Update connection profile after peer updated
		err = baseChan.ReconcileConnectionProfile(instance, update)
		if err != nil {
			return err
		}
	}

	return nil
}

// CheckStates on Channel(do nothing)
func (baseChan *BaseChannel) CheckStates(instance *current.Channel, update Update) (common.Result, error) {
	if !instance.HasType() {
		return common.Result{
			Status: &current.CRStatus{
				Type:    current.ChannelCreated,
				Version: version.Operator,
			},
		}, nil
	}

	return common.Result{}, nil
}

func (baseChan *BaseChannel) ReconcileOwnerReference(instance *current.Channel, update Update) error {
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
		instance.OwnerReferences = append(instance.OwnerReferences, ownerReference)

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

	// Full conneciton profile
	profile, err := baseChan.GenerateChannelConnProfile("", instance)
	if err != nil {
		return err
	}
	// Connection profile which only have relevant org's admin credentials
	for _, org := range instance.Spec.Members {
		p := profile.DeepCopy()
		err = baseChan.GenerateConnProfileForOrg(instance, p, org.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (baseChan *BaseChannel) GenerateChannelConnProfile(clientOrg string, channel *current.Channel) (*connector.Profile, error) {
	var err error

	// get network cluster nodes
	network := &current.Network{}
	err = baseChan.Client.Get(context.TODO(), types.NamespacedName{Name: channel.Spec.Network}, network)
	if err != nil {
		return nil, err
	}
	ordererorg := network.Labels["bestchains.network.initiator"]
	clusterNodes, err := baseChan.Initializer.GetClusterNodes(ordererorg, network.GetName())
	if err != nil {
		return nil, err
	}

	// default connprofile with default client
	basedir := baseChan.Initializer.GetStoragePath(channel)
	profile := connector.DefaultProfile(basedir, "")

	// Channel
	profile.SetChannel(channel.GetChannelID(), channel.Spec.Peers...)

	peersInSpec := make(map[string]bool)
	for _, p := range channel.Spec.Peers {
		peersInSpec[p.String()] = true
	}
	// Peers
	peers := make(map[string][]string)
	for _, p := range channel.Status.PeerConditions {
		// only joined peer can be appended into connection profile
		// only peer in spec can be appended into connection profile
		if p.Type != current.PeerJoined || !peersInSpec[p.String()] {
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
	var orgs = make([]string, len(channel.Spec.Members))
	for index, m := range channel.Spec.Members {
		orgs[index] = m.GetName()
	}
	for _, org := range orgs {
		adminUser, err := baseChan.GetOrgAdminCredentials(org)
		if err != nil {
			return nil, err
		}
		profile.SetOrganization(org, peers[org], adminUser)
	}

	yamlBinaryData, err := profile.Marshal(connector.YAML)
	if err != nil {
		return nil, err
	}
	jsonBinaryData, err := profile.Marshal(connector.JSON)
	if err != nil {
		return nil, err
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      channel.GetConnectionPorfile(),
			Namespace: baseChan.Config.Operator.Namespace,
		},
		BinaryData: map[string][]byte{
			"profile.yaml": yamlBinaryData,
			"profile.json": jsonBinaryData,
		},
	}
	err = baseChan.Client.CreateOrUpdate(context.TODO(), cm, controllerclient.CreateOrUpdateOption{
		Owner:  channel,
		Scheme: baseChan.Scheme,
	})
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (baseChan *BaseChannel) GenerateConnProfileForOrg(instance *current.Channel, profile *connector.Profile, org string) error {
	profile.SetClient(org)

	// set this org's admin credentails
	for _, member := range instance.Spec.Members {
		if member.Name != org {
			orgInfo := profile.GetOrganization(member.Name)
			profile.RemoveChannelPeers(instance.GetChannelID(), orgInfo.Peers...)
			profile.RemoveOrganizationPeers(member.Name)
			profile.RemoveOrganization(member.Name)
		}
	}

	// set 1st user as client's admin credential(for blockchain explorer)
	for _, u := range profile.GetOrganizationUsers(org) {
		profile.SetClientAdminCredential(u.Name, "")
		break
	}

	// Create conn profile cm for this organization
	yamlBinaryData, err := profile.Marshal(connector.YAML)
	if err != nil {
		return err
	}
	jsonBinaryData, err := profile.Marshal(connector.JSON)
	if err != nil {
		return err
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      instance.GetConnectionPorfile(),
			Namespace: org,
		},
		BinaryData: map[string][]byte{
			"profile.yaml": yamlBinaryData,
			"profile.json": jsonBinaryData,
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

func (baseChan *BaseChannel) GetOrgAdminCredentials(org string) (connector.User, error) {
	var err error
	organization := &current.Organization{}
	err = baseChan.Client.Get(context.TODO(), types.NamespacedName{Name: org}, organization)
	if err != nil {
		return connector.User{}, errors.Wrap(err, "failed to find organization")
	}
	orgMSPSecret := &corev1.Secret{}
	err = baseChan.Client.Get(context.TODO(), types.NamespacedName{Namespace: org, Name: fmt.Sprintf("%s-msp-crypto", org)}, orgMSPSecret)
	if err != nil {
		return connector.User{}, errors.Wrap(err, "failed to get org's msp secret")
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
	return adminUser, nil
}

// GetLabels from instance.GetLabels
func (baseChan *BaseChannel) GetLabels(instance v1.Object) map[string]string {
	return instance.GetLabels()
}

func (baseChan *BaseChannel) ReconcileChannelMember(instance *current.Channel) (err error) {
	org, err := baseChan.GetNetworkInitiatorOrg(instance)
	if err != nil {
		return errors.Wrap(err, "cant get network initiator org")
	}
	con, err := baseChan.GetChannelConnector(baseChan.Client, instance, org.GetName())
	if err != nil {
		return errors.Wrap(err, "cant get channel connector")
	}
	defer con.Close()
	client, channelConfig, err := baseChan.GetChannelConfig(con, instance, org)
	if err != nil {
		return errors.Wrap(err, "cant get channel config")
	}
	log.Info(fmt.Sprintf("channelConfig:%+v", channelConfig), "channel", instance.GetName())
	newMembers := make([]string, 0)
	for _, m := range instance.Spec.Members {
		if exist := baseChan.IsMemberInChanConfig(channelConfig, m); exist {
			continue
		}
		newMembers = append(newMembers, m.GetName())
	}
	if len(newMembers) == 0 {
		return
	}
	log.Info(fmt.Sprintf("newMembers:%s", newMembers), "channel", instance.GetName())
	if err = baseChan.AddMemberToChan(client, instance, channelConfig, newMembers); err != nil {
		return errors.Wrap(err, "cant add member to channel config")
	}
	return
}

func (baseChan *BaseChannel) GetNetworkInitiatorOrg(instance *current.Channel) (org *current.Organization, err error) {
	network := &current.Network{}
	if err = baseChan.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Network}, network); err != nil {
		return
	}
	orgName := ""
	for _, m := range network.Spec.Members {
		if m.Initiator {
			orgName = m.Name
			break
		}
	}
	org = &current.Organization{}
	if err = baseChan.Client.Get(context.TODO(), types.NamespacedName{Name: orgName}, org); err != nil {
		return
	}
	return
}

func (baseChan *BaseChannel) GetChannelConnector(c controllerclient.Client, instance *current.Channel, org string) (*connector.Connector, error) {
	profile, err := connector.ChannelProfile(c, instance.GetChannelID())
	if err != nil {
		return nil, err
	}
	profileFunc := func() (b []byte, err error) {
		profile.Client.Organization = org
		return profile.Marshal(connector.YAML)
	}
	con, err := connector.NewConnector(profileFunc)
	if err != nil {
		return nil, err
	}
	return con, err
}

func (baseChan *BaseChannel) GetChannelConfig(con *connector.Connector, instance *current.Channel, org *current.Organization) (client *resmgmt.Client, config *proto_common.Config, err error) {
	adminContext := con.SDK().Context(fabsdk.WithUser(org.Spec.Admin), fabsdk.WithOrg(org.GetName()))
	client, err = resmgmt.New(adminContext)
	if err != nil {
		return
	}
	var block *proto_common.Block
	block, err = client.QueryConfigBlockFromOrderer(instance.GetChannelID())
	if err != nil {
		return
	}
	config, err = resource.ExtractConfigFromBlock(block)
	return
}

func (baseChan *BaseChannel) IsMemberInChanConfig(config *proto_common.Config, m current.Member) (exist bool) {
	group := config.ChannelGroup.Groups[channelconfig.ApplicationGroupKey].Groups
	_, exist = group[m.GetName()]
	return
}

func (baseChan *BaseChannel) AddMemberToChan(client *resmgmt.Client, instance *current.Channel, currentConfig *proto_common.Config, orgNames []string) error {
	// Make a deep copy of the raw config as the basis for modified config
	modifiedConfig := &proto_common.Config{}
	modifiedConfigBytes, err := proto.Marshal(currentConfig)
	if err != nil {
		return errors.Wrap(err, "marshal currentConfig error")
	}
	err = proto.Unmarshal(modifiedConfigBytes, modifiedConfig)
	if err != nil {
		return errors.Wrap(err, "unmarshal currentConfig error")
	}

	// add new org to modified config
	applicationGroup := modifiedConfig.ChannelGroup.Groups[channelconfig.ApplicationGroupKey]
	for _, orgName := range orgNames {
		msg := fmt.Sprintf("org: %s ", orgName)
		org, err := baseChan.Initializer.GetApplicationOrganization(instance, orgName)
		if err != nil {
			return errors.Wrap(err, msg+"get Application organization config error")
		}
		applicationGroup.Groups[orgName], err = configtx.NewApplicationOrgGroup(org)
		if err != nil {
			return errors.Wrap(err, msg+"create application org error")
		}
	}

	// calculate  configUpdate and get its envlope bytes
	configUpdate, err := resmgmt.CalculateConfigUpdate(instance.GetChannelID(), currentConfig, modifiedConfig)
	if err != nil {
		return errors.Wrap(err, "calculate config update error")
	}
	configEnvelopeBytes, err := GetConfigEnvelopeBytes(configUpdate)
	if err != nil {
		return errors.Wrap(err, "get config envelope bytes error")
	}
	configReader := bytes.NewReader(configEnvelopeBytes)

	// get all signingIdentities needed
	signIdentities := make([]msp.SigningIdentity, 0)
	orgCon, err := baseChan.GetChannelConnector(baseChan.Client, instance, "")
	if err != nil {
		return errors.Wrap(err, "get channel connector error")
	}
	defer orgCon.Close()
	for _, member := range instance.Spec.Members {
		if util.ContainsValue(member.Name, orgNames) {
			log.Info("skip org sign config update, because of new org to channel", "org", member.GetName())
			continue
		}
		msg := fmt.Sprintf("org: %s ", member.GetName())
		organization := &current.Organization{}
		organization.Name = member.GetName()
		if err = baseChan.Client.Get(context.TODO(), types.NamespacedName{Name: organization.GetName()}, organization); err != nil {
			return errors.Wrap(err, msg+"get org error")
		}
		clientContext := orgCon.SDK().Context(fabsdk.WithUser(organization.Spec.Admin), fabsdk.WithOrg(organization.GetName()))
		cctx, err := clientContext()
		if err != nil {
			return err
		}
		signIdentities = append(signIdentities, cctx)
	}

	// update channel config
	txID, err := client.SaveChannel(resmgmt.SaveChannelRequest{
		ChannelID:         instance.GetChannelID(),
		ChannelConfig:     configReader,
		SigningIdentities: signIdentities,
	})
	if err != nil {
		return errors.Wrap(err, "save channel error")
	}
	log.Info(fmt.Sprintf("update channel config to update member in txID:%s", txID.TransactionID), "channel", instance.GetName(), "newMember", orgNames)
	return nil
}

func GetConfigEnvelopeBytes(configUpdate *proto_common.ConfigUpdate) ([]byte, error) {
	var buf bytes.Buffer
	if err := protolator.DeepMarshalJSON(&buf, configUpdate); err != nil {
		return nil, err
	}
	channelConfigBytes, err := proto.Marshal(configUpdate)
	if err != nil {
		return nil, err
	}
	configUpdateEnvelope := &proto_common.ConfigUpdateEnvelope{
		ConfigUpdate: channelConfigBytes,
		Signatures:   nil,
	}
	configUpdateEnvelopeBytes, err := proto.Marshal(configUpdateEnvelope)
	if err != nil {
		return nil, err
	}
	payload := &proto_common.Payload{
		Data: configUpdateEnvelopeBytes,
	}
	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		return nil, err
	}
	configEnvelope := &proto_common.Envelope{
		Payload: payloadBytes,
	}
	return proto.Marshal(configEnvelope)
}
