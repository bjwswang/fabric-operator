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
	"net/http"
	"path/filepath"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/pkg/initializer/orderer/configtx"
	"github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	k8sclient "github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	NODE = "node"
)

var (
	ErrChannelAlreadyExist = errors.New("channel already exist in target node")
)

var log = logf.Log.WithName("base_channel_initializer")

type Config struct {
	StoragePath string
}

// Initializer is for channel initialization
type Initializer struct {
	Config *Config

	Scheme *runtime.Scheme
	Client k8sclient.Client
}

func New(client controllerclient.Client, scheme *runtime.Scheme, cfg *Config) *Initializer {
	initializer := &Initializer{
		Client: client,
		Scheme: scheme,
		Config: cfg,
	}

	return initializer
}

// CreateChannel used to help create a channel within network,including:
//   - create a genesis block for channel
//   - join all orderer nodes into channel
func (i *Initializer) CreateChannel(instance *current.Channel) error {
	var err error

	network := &current.Network{}
	err = i.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Network}, network)
	if err != nil {
		return err
	}

	ordererorg := network.Labels["bestchains.network.initiator"]

	// get network's orderer nodes
	parentOrderer, err := i.GetParentNode(ordererorg, network.GetName())
	if err != nil {
		return err
	}
	if parentOrderer.Status.Type != current.Deployed {
		return errors.Errorf("consensus parent node {name:%s,namespace:%s} not deployed yet", parentOrderer.GetName(), parentOrderer.GetNamespace())
	}
	clusterNodes, err := i.GetClusterNodes(ordererorg, network.GetName())
	if err != nil {
		return err
	}

	// create genesis block for channel
	block, err := i.CreateGenesisBlock(instance, ordererorg, parentOrderer, clusterNodes)
	if err != nil {
		return err
	}

	// Join all cluster nodes into this channel
	osn, err := NewOSNAdmin(i.Client, ordererorg, clusterNodes.Items...)
	if err != nil {
		return err
	}
	for _, target := range clusterNodes.Items {
		// make sure orderer not joined yet
		resp, err := osn.Query(target.GetName(), instance.GetChannelID())
		if err != nil {
			return err
		}
		// continue if current orderer node already joins
		if resp.StatusCode != http.StatusNotFound {
			continue
		}
		err = osn.Join(target.GetName(), block)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateGenesisBlock configures and generate a genesis block for channel startup.Here we have these limitations:
// - system channel not supported
// - Capability use `V2_0`
func (i *Initializer) CreateGenesisBlock(instance *current.Channel, ordererorg string, parentOrderer *current.IBPOrderer, clusterNodes *current.IBPOrdererList) ([]byte, error) {
	configTx := configtx.New()

	// `Application` defines a application channel
	profile, err := configTx.GetProfile("Application")
	if err != nil {
		return nil, err
	}

	// add orderer settings into profile
	mspConfigs, err := i.ConfigureOrderer(instance, profile, ordererorg, parentOrderer, clusterNodes)
	if err != nil {
		return nil, err
	}

	// add application settings into profile
	isUsingChannelLess := true
	if !isUsingChannelLess {
		return nil, errors.New("system channel not supported yet")
	} else {
		err = i.ConfigureApplication(instance, profile)
		if err != nil {
			return nil, err
		}
	}

	channelID := instance.GetChannelID()
	block, err := profile.GenerateBlock(channelID, mspConfigs)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// GetParentNode returns the IBPOrderer which acts as the parent in consenus cluster
func (i *Initializer) GetParentNode(namespace string, parentNode string) (*current.IBPOrderer, error) {
	orderer := &current.IBPOrderer{}
	err := i.Client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: parentNode}, orderer)
	if err != nil {
		return nil, err
	}
	return orderer, nil
}

// GetClusterNodes returns the IBPOrderers which acts as the real consensus node
func (i *Initializer) GetClusterNodes(namespace string, parentNode string) (*current.IBPOrdererList, error) {
	ordererList := &current.IBPOrdererList{}

	labelSelector, err := labels.Parse(fmt.Sprintf("parent=%s", parentNode))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse selector for parent name")
	}

	listOptions := &client.ListOptions{
		LabelSelector: labelSelector,
		Namespace:     namespace,
	}

	err = i.Client.List(context.TODO(), ordererList, listOptions)
	if err != nil {
		return nil, err
	}

	return ordererList, nil
}

// GetStoragePath in formart `/chaninit/{channel_name}`
func (i *Initializer) GetStoragePath(instance *current.Channel) string {
	return filepath.Join("/", i.Config.StoragePath, instance.GetName())
}

// GetOrgMSPDir returns channel organization's msp directory which will be used for genesis block generation
func (i *Initializer) GetOrgMSPDir(instance *current.Channel, orgMSPID string) string {
	return filepath.Join(i.GetStoragePath(instance), orgMSPID, "msp")
}
