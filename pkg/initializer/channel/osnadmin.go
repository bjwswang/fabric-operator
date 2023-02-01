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

// OSNAdmin wraps a client to call orderering service's admin api
type OSNAdmin struct {
	client     k8sclient.Client
	ordererorg string
	targets    map[string]*Target
}

// Target wraps connection info of a orderer node
type Target struct {
	URL    string
	Client *http.Client
}

func NewOSNAdmin(client k8sclient.Client, ordererorg string, targetOrderers ...current.IBPOrderer) (*OSNAdmin, error) {
	var err error

	if ordererorg == "" {
		return nil, errors.Errorf("osnadmin must have ordererorg configured")
	}

	// get ordererorg' msp crypto
	organization := &current.Organization{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: ordererorg}, organization)
	if err != nil {
		return nil, errors.Wrapf(err, "get ordererorg %s", ordererorg)
	}

	osn := &OSNAdmin{
		client:     client,
		ordererorg: ordererorg,
		targets:    make(map[string]*Target),
	}

	// add all orderers into osn targets
	for index := range targetOrderers {
		err = osn.AddTarget(&targetOrderers[index])
		if err != nil {
			return nil, err
		}
	}

	return osn, nil
}

/* Target management */

func (osn *OSNAdmin) AddTarget(targetOrderer *current.IBPOrderer) error {
	var err error

	// get connection profile
	connProfile, err := GetTargetConnectionProfile(osn.client, targetOrderer)
	if err != nil {
		return err
	}
	url := connProfile.Endpoints.Admin
	tlsServerCert := connProfile.TLS.SignCerts

	// load orderering service's server tls cert
	tlsServerCertPem, err := base64.StdEncoding.DecodeString(tlsServerCert)
	if err != nil {
		return err
	}
	certPool, err := util.LoadToCertPool(tlsServerCertPem)
	if err != nil {
		return err
	}

	// Load ordererorg's admin user tls key & cert from orderer org's msp crypto
	ordererorgMSP := &corev1.Secret{}
	err = osn.client.Get(context.TODO(), types.NamespacedName{Namespace: osn.ordererorg, Name: fmt.Sprintf("%s-msp-crypto", osn.ordererorg)}, ordererorgMSP)
	if err != nil {
		return errors.Wrapf(err, "get ordererorg %s msp crypto", osn.ordererorg)
	}
	tlsClientKeyPem := ordererorgMSP.Data["admin-tls-keystore"]
	tlsClientCertPem := ordererorgMSP.Data["admin-tls-signcert"]
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

/* Channel management */

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
		return errors.Errorf("Join failed: %s", res.Body)
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

// List all channels which target has joined
func (osn *OSNAdmin) List(target string) (*http.Response, error) {
	instance, err := osn.GetTarget(target)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf(ListChannelsAPI, instance.URL)
	return instance.Client.Get(url)
}

// Query a channel from target
func (osn *OSNAdmin) Query(target string, channelID string) (*http.Response, error) {
	instance, err := osn.GetTarget(target)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf(QueryChannelAPI, instance.URL, channelID)
	resp, err := instance.Client.Get(url)
	if err != nil {
		return nil, err
	}
	return resp, err
}

// WaitForChannel wait until channel is ready
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

// GetTargetConnectionProfile helps retrived target orderer's connection profile which contains its tls server cert
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
