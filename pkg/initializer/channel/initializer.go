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
	"github.com/IBM-Blockchain/fabric-operator/pkg/initializer/common/secretmanager"
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

var log = logf.Log.WithName("base_channel_initializer")

type Config struct {
	ConfigtxFile string
	StoragePath  string
}

type Initializer struct {
	Config *Config
	Scheme *runtime.Scheme
	Client k8sclient.Client

	SecretManager *secretmanager.SecretManager
}

func New(client controllerclient.Client, scheme *runtime.Scheme, cfg *Config) *Initializer {
	initializer := &Initializer{
		Client: client,
		Scheme: scheme,
		Config: cfg,
	}

	initializer.SecretManager = secretmanager.New(client, scheme, nil)

	return initializer
}

func (i *Initializer) GetStoragePath(instance *current.Channel) string {
	return filepath.Join("/", i.Config.StoragePath, instance.GetName())
}

func (i *Initializer) GetOrgMSPDir(instance *current.Channel, orgMSPID string) string {
	return filepath.Join(i.GetStoragePath(instance), orgMSPID, "msp")
}

func (i *Initializer) CreateOrUpdateChannel(instance *current.Channel) error {
	var err error

	network := &current.Network{}
	err = i.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Network}, network)
	if err != nil {
		return err
	}

	ordererorg := network.Labels["bestchains.network.initiator"]
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

	osn, err := NewOSNAdmin(i.Client, ordererorg, clusterNodes.Items...)
	if err != nil {
		return err
	}

	var exist = true
	resp, err := osn.Query(clusterNodes.Items[0].GetName(), instance.GetName())
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		exist = false
	}
	// DO NOT SUPPORT UPDATE FOR NOW
	if exist {
		return nil
	}

	block, err := i.CreateGenesisBlock(instance, ordererorg, parentOrderer, clusterNodes)
	if err != nil {
		return err
	}

	// Join all cluster nodes into this channel
	for _, target := range clusterNodes.Items {
		err = osn.Join(target.GetName(), block)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Initializer) CreateGenesisBlock(instance *current.Channel, ordererorg string, parentOrderer *current.IBPOrderer, clusterNodes *current.IBPOrdererList) ([]byte, error) {
	configTx := configtx.New()
	profile, err := configTx.GetProfile("Initial")
	if err != nil {
		return nil, err
	}

	mspConfigs, err := i.ConfigureOrderer(instance, profile, ordererorg, parentOrderer, clusterNodes)
	if err != nil {
		return nil, err
	}

	isUsingChannelLess := true
	if !isUsingChannelLess {
		return nil, errors.New("system channel not supported yet")
	} else {
		err = i.ConfigureApplication(instance, profile)
		if err != nil {
			return nil, err
		}
	}
	channelID := instance.GetName()
	block, err := profile.GenerateBlock(channelID, mspConfigs)
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (i *Initializer) GetParentNode(namespace string, parentNode string) (*current.IBPOrderer, error) {
	orderer := &current.IBPOrderer{}
	err := i.Client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: parentNode}, orderer)
	if err != nil {
		return nil, err
	}
	return orderer, nil
}

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
