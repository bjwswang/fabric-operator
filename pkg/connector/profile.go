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
	"net/url"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/util/pointer"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Profile contasins all we need to connect with a blockchain network. Currently we use embeded pem by default
type Profile struct {
	Version       string `yaml:"version,omitempty"`
	Client        `yaml:"client,omitempty"`
	Channels      map[string]ChannelInfo      `yaml:"channels,omitempty"`
	Organizations map[string]OrganizationInfo `yaml:"organizations,omitempty"`
	// Orderers defines all orderer endpoints which can be used
	Orderers map[string]NodeEndpoint `yaml:"orderers,omitempty"`
	// Peers defines all peer endpoints which can be used
	Peers map[string]NodeEndpoint `yaml:"peers,omitempty"`
}

// Client defines who is trying to connect with network
type Client struct {
	Organization string `yaml:"organization,omitempty"`
	Logging      `yaml:"logging,omitempty"`
	// CryptoConfig `yaml:"cryptoconfig,omitempty"`
}

type Logging struct {
	Level string `yaml:"level,omitempty"`
}

type CryptoConfig struct {
	Path string `yaml:"path,omitempty"`
}

// ChannelInfo defines configurations when connect to this channel
type ChannelInfo struct {
	// Peers which can be used to connect to this channel
	Peers map[string]PeerInfo `yaml:"peers,omitempty"`
}

type PeerInfo struct {
	EndorsingPeer  *bool `yaml:"endorsingPeer,omitempty"`
	ChaincodeQuery *bool `yaml:"chaincodeQuery,omitempty"`
	LedgerQuery    *bool `yaml:"ledgerQuery,omitempty"`
	EventSource    *bool `yaml:"eventSource,omitempty"`
}

// OrganizationInfo defines a organization along with its users and peers
type OrganizationInfo struct {
	MSPID string          `yaml:"mspid,omitempty"`
	Users map[string]User `yaml:"users,omitempty"`
	// CryptoPath string          `yaml:"cryptoPath,omitempty"`
	Peers []string `yaml:"peers,omitempty"`
}

// User is the ca identity which has a private key(embeded pem) and signed certificate(embeded pem)
type User struct {
	Name string `yaml:"name,omitempty"`
	Key  Pem    `yaml:"key,omitempty"`
	Cert Pem    `yaml:"cert,omitempty"`
}

type Pem struct {
	Pem string `yaml:"pem,omitempty"`
}

type NodeEndpoint struct {
	URL        string `yaml:"url,omitempty"`
	TLSCACerts `yaml:"tlsCACerts,omitempty"`
}

type TLSCACerts struct {
	Path string `yaml:"path,omitempty"`
	Pem  string `yaml:"pem,omitempty"`
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
		// CryptoConfig: CryptoConfig{
		// 	Path: baseDir,
		// },
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
	// profile.Client.CryptoConfig.Path = baseDir
	profile.Client.Organization = clientorg
}

func (profile *Profile) SetChannel(channelID string, peers ...current.NamespacedName) {
	info, ok := profile.Channels[channelID]
	if !ok {
		info.Peers = make(map[string]PeerInfo)
	}
	for _, p := range peers {
		info.Peers[p.String()] = *DefaultPeerInfo()
	}
}

/* Channel settings in Profile */

func (profile *Profile) GetChannel(channelID string) ChannelInfo {
	if profile.Channels == nil {
		profile.Channels = make(map[string]ChannelInfo)
		return ChannelInfo{
			Peers: map[string]PeerInfo{},
		}
	}
	return profile.Channels[channelID]
}

/* Organization settings in Profile */

func (profile *Profile) GetOrganization(organization string) OrganizationInfo {
	if profile.Organizations == nil {
		profile.Organizations = make(map[string]OrganizationInfo)
		return OrganizationInfo{
			MSPID: organization,
			Users: make(map[string]User),
			// CryptoPath: filepath.Join(organization, "users", "{username}", "msp"), // {org_name}/users/{username}/msp
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
		MSPID: organization,
		Users: make(map[string]User),
		// CryptoPath: filepath.Join(organization, "users", "{username}", "msp"), // {org_name}/users/{username}/msp
		Peers: peers,
	}
	for _, user := range users {
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

func (profile *Profile) RemoveOrganization(organization string) {
	if profile.Organizations == nil {
		profile.Organizations = make(map[string]OrganizationInfo)
		return
	}
	delete(profile.Organizations, organization)
}

/* Peer settings in Profile */

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

	apiURL, err := url.Parse(conn.Endpoints.API)
	if err != nil {
		return NodeEndpoint{}, errors.Wrap(err, "invalid node api")
	}

	tlsPem, err := base64.StdEncoding.DecodeString(conn.TLS.SignCerts)
	if err != nil {
		return NodeEndpoint{}, errors.Wrap(err, "not a valid pem format cert")
	}
	return NodeEndpoint{
		URL: apiURL.Host,
		TLSCACerts: TLSCACerts{
			Pem: string(tlsPem),
		},
	}, nil
}

/* Marshal/Unmarshal in Profile*/
func (profile *Profile) Marshal() ([]byte, error) {
	return yaml.Marshal(profile)
}

func (profile *Profile) Unmarshal(in []byte) error {
	return yaml.Unmarshal(in, profile)
}
