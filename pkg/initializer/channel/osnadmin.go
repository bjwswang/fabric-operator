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
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/pkg/errors"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	k8sclient "github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// JoinChannelAPI for join orderer into channel
	JoinChannelAPI = "%s/participation/v1/channels"
	// ListChannelsAPI list all channels
	ListChannelsAPI = "%s/participation/v1/channels"
	// QueryChannelAPI query a specific channel
	QueryChannelAPI = "%s/participation/v1/channels/%s"
)

var (
	ErrTargetNotFound = errors.New("target not found")
)

// OSNClient orderering service node client
type OSNAdmin struct {
	client  k8sclient.Client
	targets map[string]*Target
}

type Target struct {
	URL    string
	Client *http.Client
}

func NewOSNAdmin(client k8sclient.Client, ordererorg string, targetOrderers ...current.IBPOrderer) (*OSNAdmin, error) {
	var err error
	osn := &OSNAdmin{
		client:  client,
		targets: make(map[string]*Target),
	}

	organization := &current.Organization{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: ordererorg}, organization)
	if err != nil {
		return nil, errors.Wrapf(err, "get ordererorg %s", ordererorg)
	}
	orgmsp := &corev1.Secret{}
	err = client.Get(context.TODO(), organization.GetMSPCrypto(), orgmsp)
	if err != nil {
		return nil, errors.Wrapf(err, "get ordererorg %s msp crypto", ordererorg)
	}

	for index := range targetOrderers {
		err = osn.AddTarget(orgmsp, &targetOrderers[index])
		if err != nil {
			return nil, err
		}
	}
	return osn, nil
}

func (osn *OSNAdmin) AddTarget(orderermsp *corev1.Secret, targetOrderer *current.IBPOrderer) error {
	// get connection profile
	connProfile, err := GetTargetConnectionProfile(osn.client, targetOrderer)
	if err != nil {
		return err
	}
	url := connProfile.Endpoints.Admin
	tlsServerCert := connProfile.TLS.CACerts[0]

	// load orderering service's server tls cert
	tlsServerCertPem, err := base64.StdEncoding.DecodeString(tlsServerCert)
	if err != nil {
		return err
	}
	certPool, err := util.LoadToCertPool(tlsServerCertPem)
	if err != nil {
		return err
	}

	// Load ordererorg's admin user tls key & cert
	tlsClientKeyPem := orderermsp.Data["admin-tls-keystore"]
	tlsClientCertPem := orderermsp.Data["admin-tls-signcert"]
	tlsCert, err := tls.X509KeyPair(tlsClientCertPem, tlsClientKeyPem)
	if err != nil {
		return err
	}

	// target with http client (mutual tls)
	osn.targets[targetOrderer.Name] = &Target{
		URL: url,
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:      certPool,
					Certificates: []tls.Certificate{tlsCert},
				},
			},
		},
	}

	return nil
}

func (osn *OSNAdmin) GetTarget(target string) (*Target, error) {
	if osn.targets == nil {
		osn.targets = make(map[string]*Target)
		return nil, ErrTargetNotFound
	}
	instance, ok := osn.targets[target]
	if !ok {
		return nil, ErrTargetNotFound
	}
	return instance, nil
}

func (osn *OSNAdmin) DeleteTarget(target string) {
	if osn.targets == nil {
		osn.targets = make(map[string]*Target)
	}
	delete(osn.targets, target)
}

func GetTargetConnectionProfile(client k8sclient.Client, orderer *current.IBPOrderer) (*current.OrdererConnectionProfile, error) {
	var err error

	// consensus connection info
	cm := &corev1.ConfigMap{}
	err = client.Get(
		context.TODO(),
		types.NamespacedName{
			Name:      orderer.GetName() + "-connection-profile",
			Namespace: orderer.GetNamespace(),
		},
		cm,
	)
	if err != nil {
		return nil, err
	}

	connectionProfile := &current.OrdererConnectionProfile{}
	if err := json.Unmarshal(cm.BinaryData["profile.json"], connectionProfile); err != nil {
		return nil, err
	}

	return connectionProfile, nil
}

// Join orderers into channel
func (osn *OSNAdmin) Join(target string, blockBytes []byte) error {
	instance, err := osn.GetTarget(target)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(JoinChannelAPI, instance.URL)
	req, err := createJoinRequest(url, blockBytes)
	if err != nil {
		return err
	}
	res, err := instance.Client.Do(req)
	if err != nil {
		return err
	}
	if !checkJoinResponse(res) {
		return errors.New("Join failed")
	}
	return nil
}

func createJoinRequest(url string, blockBytes []byte) (*http.Request, error) {
	joinBody := new(bytes.Buffer)
	writer := multipart.NewWriter(joinBody)
	part, err := writer.CreateFormFile("config-block", "config.block")
	if err != nil {
		return nil, err
	}
	part.Write(blockBytes)
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, joinBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}

// github.com/hyperledger/fabric/orderer/common/channelparticipation/restapi.go#343
func checkJoinResponse(res *http.Response) bool {
	switch res.StatusCode {
	// 201
	case http.StatusCreated:
		return true
	// 405
	case http.StatusMethodNotAllowed:
		return true
	}
	return false
}

// List all channels
func (osn *OSNAdmin) List(target string) (*http.Response, error) {
	instance, err := osn.GetTarget(target)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf(ListChannelsAPI, instance.URL)
	return instance.Client.Get(url)
}

// Query a chanenl
func (osn *OSNAdmin) Query(target string, channelID string) (*http.Response, error) {
	instance, err := osn.GetTarget(target)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf(QueryChannelAPI, instance.URL, channelID)
	return instance.Client.Get(url)
}

// ChainReady checks whether channel is ready
func (osn *OSNAdmin) WaitForChannel(target string, channelID string, duration time.Duration) error {
	instance, err := osn.GetTarget(target)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(QueryChannelAPI, instance.URL, channelID)
	timeout := time.After(duration)
	for {
		select {
		case <-timeout:
			return errors.New("timeout excceeded")
		default:
		}
		res, err := instance.Client.Get(url)
		if err != nil {
			return err
		}
		if res.StatusCode == http.StatusOK {
			return nil
		}
	}
}
