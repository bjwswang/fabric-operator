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

package connector

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"path/filepath"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/util"
	"github.com/IBM-Blockchain/fabric-operator/pkg/util/pointer"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Profile contasins all we need to connect with a blockchain network. Currently we use embeded pem by default
// +k8s:deepcopy-gen=true
type Profile struct {
	Version       string `yaml:"version,omitempty" json:"version,omitempty"`
	Client        `yaml:"client,omitempty" json:"client,omitempty"`
	Channels      map[string]ChannelInfo      `yaml:"channels" json:"channels"`
	Organizations map[string]OrganizationInfo `yaml:"organizations,omitempty" json:"organizations,omitempty"`
	// Orderers defines all orderer endpoints which can be used
	Orderers map[string]NodeEndpoint `yaml:"orderers,omitempty" json:"orderers,omitempty"`
	// Peers defines all peer endpoints which can be used
	Peers map[string]NodeEndpoint `yaml:"peers,omitempty" json:"peers,omitempty"`
}

// Client defines who is trying to connect with networks
type Client struct {
	Organization string `yaml:"organization,omitempty" json:"organization,omitempty"`
	Logging      `yaml:"logging,omitempty" json:"logging,omitempty"`
	// For blockchain explorer
	AdminCredential `yaml:"adminCredential,omitempty" json:"adminCredential,omitempty"`
	CredentialStore `yaml:"credentialStore,omitempty" json:"credentialStore,omitempty"`
	TLSEnable       bool `yaml:"tlsEnable,omitempty" json:"tlsEnable,omitempty"`
}

// +k8s:deepcopy-gen=true
type Logging struct {
	Level string `yaml:"level,omitempty" json:"level,omitempty"`
}

// +k8s:deepcopy-gen=true
type CryptoConfig struct {
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

// +k8s:deepcopy-gen=true
type AdminCredential struct {
	ID       string `yaml:"id,omitempty" json:"id,omitempty"`
	Password string `yaml:"password,omitempty" json:"password" default:"passw0rd"`
}

// +k8s:deepcopy-gen=true
type CredentialStore struct {
	Path        string `yaml:"path,omitempty" json:"path,omitempty"`
	CryptoStore `yaml:"cryptoStore,omitempty" json:"cryptoStore,omitempty"`
}

// +k8s:deepcopy-gen=true
type CryptoStore struct {
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

// ChannelInfo defines configurations when connect to this channel
// +k8s:deepcopy-gen=true
type ChannelInfo struct {
	// Peers which can be used to connect to this channel
	Peers map[string]PeerInfo `yaml:"peers" json:"peers"`
}

// +k8s:deepcopy-gen=true
type PeerInfo struct {
	EndorsingPeer  *bool `yaml:"endorsingPeer,omitempty" json:"endorsingPeer,omitempty"`
	ChaincodeQuery *bool `yaml:"chaincodeQuery,omitempty" json:"chaincodeQuery,omitempty"`
	LedgerQuery    *bool `yaml:"ledgerQuery,omitempty" json:"ledgerQuery,omitempty"`
	EventSource    *bool `yaml:"eventSource,omitempty" json:"eventSource,omitempty"`
}

// OrganizationInfo defines a organization along with its users and peers
// +k8s:deepcopy-gen=true
type OrganizationInfo struct {
	MSPID string          `yaml:"mspid,omitempty" json:"mspid,omitempty"`
	Users map[string]User `yaml:"users,omitempty" json:"users,omitempty"`
	Peers []string        `yaml:"peers,omitempty" json:"peers,omitempty"`

	// For blockchain explorer
	AdminPrivateKey Pem `yaml:"adminPrivateKey,omitempty" json:"adminPrivateKey,omitempty"`
	SignedCert      Pem `yaml:"signedCert,omitempty" json:"signedCert,omitempty"`
}

// User is the ca identity which has a private key(embeded pem) and signed certificate(embeded pem)
// +k8s:deepcopy-gen=true
type User struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	Key  Pem    `yaml:"key,omitempty" json:"key,omitempty"`
	Cert Pem    `yaml:"cert,omitempty" json:"cert,omitempty"`
}

// +k8s:deepcopy-gen=true
type Pem struct {
	Pem string `yaml:"pem,omitempty" json:"pem,omitempty"`
}

// +k8s:deepcopy-gen=true
type NodeEndpoint struct {
	URL        string `yaml:"url,omitempty" json:"url,omitempty"`
	TLSCACerts `yaml:"tlsCACerts,omitempty" json:"tlsCACerts,omitempty"`
}

// +k8s:deepcopy-gen=true
type TLSCACerts struct {
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
	Pem  string `yaml:"pem,omitempty" json:"pem,omitempty"`
}

/* Default in Profile */

func DefaultProfile(baseDir string, org string) *Profile {
	return &Profile{
		Version:       "1.0.0",
		Client:        *DefaultClient(baseDir, org),
		Channels:      make(map[string]ChannelInfo),
		Organizations: make(map[string]OrganizationInfo),
		Peers:         make(map[string]NodeEndpoint),
		Orderers:      make(map[string]NodeEndpoint),
	}
}

func DefaultClient(baseDir string, org string) *Client {
	return &Client{
		Organization: org,
		Logging: Logging{
			Level: "info",
		},
		// Must specify CredentialStore to avoid `mkdir keystore permission error`
		CredentialStore: CredentialStore{
			Path: filepath.Join(baseDir, org, "hfc-kvs"),
			CryptoStore: CryptoStore{
				Path: filepath.Join(baseDir, org, "hfc-cvs"),
			},
		},
		TLSEnable: true,
	}
}

func DefaultChannelInfo() *ChannelInfo {
	return &ChannelInfo{
		Peers: make(map[string]PeerInfo),
	}
}

func DefaultPeerInfo() *PeerInfo {
	return &PeerInfo{
		EndorsingPeer:  pointer.True(),
		ChaincodeQuery: pointer.True(),
		LedgerQuery:    pointer.True(),
		EventSource:    pointer.True(),
	}
}

/* Client settings in Profile */

func (profile *Profile) SetClient(clientorg string) {
	profile.Client.Organization = clientorg
}

func (profile *Profile) SetClientAdminCredential(id string, password string) {
	if password == "" {
		password = "passw0rd"
	}
	profile.Client.AdminCredential = AdminCredential{
		ID:       id,
		Password: password,
	}
}

/* Channel settings in Profile */

func (profile *Profile) SetChannel(channelID string, peers ...current.NamespacedName) {
	info, ok := profile.Channels[channelID]
	if !ok {
		info.Peers = make(map[string]PeerInfo)
	}
	for _, p := range peers {
		info.Peers[p.String()] = *DefaultPeerInfo()
	}
	profile.Channels[channelID] = info
}

func (profile *Profile) RemoveChannelPeers(channelID string, peers ...string) {
	info, ok := profile.Channels[channelID]
	if !ok {
		info.Peers = make(map[string]PeerInfo)
	}
	for _, p := range peers {
		delete(info.Peers, p)
	}
	profile.Channels[channelID] = info
}

func (profile *Profile) GetChannel(channelID string) ChannelInfo {
	if profile.Channels == nil {
		profile.Channels = make(map[string]ChannelInfo)
	}
	v, ok := profile.Channels[channelID]
	if !ok {
		v = ChannelInfo{
			Peers: make(map[string]PeerInfo),
		}
	}
	return v
}

/* Organization settings in Profile */

func (profile *Profile) GetOrganization(organization string) OrganizationInfo {
	if profile.Organizations == nil {
		profile.Organizations = make(map[string]OrganizationInfo)
		return OrganizationInfo{
			MSPID: organization,
			Users: make(map[string]User),
			Peers: make([]string, 0),
		}
	}
	return profile.Organizations[organization]
}

func (profile *Profile) SetOrganization(organization string, peers []string, users ...User) {
	if profile.Organizations == nil {
		profile.Organizations = make(map[string]OrganizationInfo)
	}
	info := OrganizationInfo{
		MSPID:           organization,
		Users:           make(map[string]User),
		Peers:           peers,
		AdminPrivateKey: Pem{},
		SignedCert:      Pem{},
	}
	for _, user := range users {
		if info.AdminPrivateKey.Pem == "" {
			info.AdminPrivateKey.Pem = user.Key.Pem
			info.SignedCert.Pem = user.Cert.Pem
		}
		info.Users[user.Name] = user
	}
	profile.Organizations[organization] = info
}

func (profile *Profile) SetOrganizationUsers(organization string, users ...User) {
	orgInfo := profile.GetOrganization(organization)
	if orgInfo.Users == nil {
		orgInfo.Users = make(map[string]User)
	}
	for _, user := range users {
		orgInfo.Users[user.Name] = user
	}
	profile.Organizations[organization] = orgInfo
}

func (profile *Profile) GetOrganizationUsers(organization string) map[string]User {
	orgInfo := profile.GetOrganization(organization)
	if orgInfo.Users == nil {
		orgInfo.Users = make(map[string]User)
	}
	return orgInfo.Users
}

func (profile *Profile) RemoveOrganizationUsers(organization string) {
	orgInfo := profile.GetOrganization(organization)
	orgInfo.Users = make(map[string]User)
	profile.Organizations[organization] = orgInfo
}

func (profile *Profile) RemoveOrganization(organization string) {
	if profile.Organizations == nil {
		profile.Organizations = make(map[string]OrganizationInfo)
		return
	}
	delete(profile.Organizations, organization)
}

/* Peer settings in Profile */

func (profile *Profile) RemoveOrganizationPeers(org string) {
	orgInfo := profile.GetOrganization(org)
	for _, p := range orgInfo.Peers {
		delete(profile.Peers, p)
	}
}

func (profile *Profile) SetPeer(client controllerclient.Client, peer current.NamespacedName) error {
	if profile.Peers == nil {
		profile.Peers = make(map[string]NodeEndpoint)
	}
	endpoint, err := GetNodeEndpoint(client, peer)
	if err != nil {
		return err
	}
	profile.Peers[peer.String()] = endpoint

	// add peer to its organization
	orgInfo := profile.GetOrganization(peer.Namespace)
	orgInfo.Peers = append(orgInfo.Peers, peer.String())
	if profile.Organizations == nil {
		profile.Organizations = make(map[string]OrganizationInfo)
	}
	profile.Organizations[peer.Namespace] = orgInfo

	return nil
}

func (profile *Profile) RemovePeer(peer current.NamespacedName) {
	if profile.Peers == nil {
		profile.Peers = make(map[string]NodeEndpoint)
	}
	delete(profile.Peers, peer.String())
}

/* Orderer settings in Profile */

func (profile *Profile) SetOrderer(client controllerclient.Client, orderer current.NamespacedName) error {
	if profile.Orderers == nil {
		profile.Orderers = make(map[string]NodeEndpoint)
	}
	endpoint, err := GetNodeEndpoint(client, orderer)
	if err != nil {
		return err
	}
	profile.Orderers[orderer.String()] = endpoint
	return nil
}

func (profile *Profile) RemoveOrderer(orderer current.NamespacedName) {
	if profile.Orderers == nil {
		profile.Orderers = make(map[string]NodeEndpoint)
	}
	delete(profile.Orderers, orderer.String())
}

// GetNodeEndpoint with node(peer/orderer)'s connection profile
func GetNodeEndpoint(client controllerclient.Client, node current.NamespacedName) (NodeEndpoint, error) {
	cm := &corev1.ConfigMap{}
	err := client.Get(context.TODO(), types.NamespacedName{Namespace: node.Namespace, Name: node.Name + "-connection-profile"}, cm)
	if err != nil {
		return NodeEndpoint{}, err
	}

	conn := &current.PeerConnectionProfile{}
	if err := json.Unmarshal(cm.BinaryData["profile.json"], conn); err != nil {
		return NodeEndpoint{}, err
	}

	tlsPem, err := base64.StdEncoding.DecodeString(conn.TLS.CACerts[0])
	if err != nil {
		return NodeEndpoint{}, errors.Wrap(err, "not a valid pem format cert")
	}
	return NodeEndpoint{
		URL: conn.Endpoints.API,
		TLSCACerts: TLSCACerts{
			Pem: string(tlsPem),
		},
	}, nil
}

type Format string

const (
	JSON Format = "json"
	YAML Format = "yaml"
)

/* Marshal/Unmarshal in Profile*/
func (profile *Profile) Marshal(format Format) ([]byte, error) {
	switch format {
	case JSON:
		return json.Marshal(profile)
	}
	return yaml.Marshal(profile)
}

func (profile *Profile) Unmarshal(in []byte, format Format) error {
	switch format {
	case JSON:
		return json.Unmarshal(in, profile)
	}
	return yaml.Unmarshal(in, profile)
}

func ChannelProfile(cli controllerclient.Client, channelID string) (p *Profile, err error) {
	operatorNamespace, err := util.GetNamespace()
	if err != nil {
		return nil, err
	}
	channel := current.Channel{}
	channel.Name = channelID
	cm := &corev1.ConfigMap{}
	cm.Name = channel.GetConnectionPorfile()
	cm.Namespace = operatorNamespace
	if err = cli.Get(context.TODO(), client.ObjectKeyFromObject(cm), cm); err != nil {
		return nil, errors.Wrap(err, "failed to get channel connection profile")
	}
	profile := &Profile{}
	if err = profile.Unmarshal(cm.BinaryData["profile.yaml"], YAML); err != nil {
		return nil, errors.Wrap(err, "invalid channel connection profile")
	}
	return profile, nil
}
