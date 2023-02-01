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

package organization

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/pkg/initializer/common"
	commonconfig "github.com/IBM-Blockchain/fabric-operator/pkg/initializer/common/config"
	"github.com/IBM-Blockchain/fabric-operator/pkg/initializer/common/secretmanager"
	orginit "github.com/IBM-Blockchain/fabric-operator/pkg/initializer/organization"
	k8sclient "github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/util"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	// maxRetryCount in WaitCA
	maxRetryCount = 10
)

//go:generate counterfeiter -o mocks/crypto.go -fake-name Crypto . Crypto

type Crypto interface {
	GetCrypto() (*commonconfig.Response, error)
	PingCA() error
	Validate() error
}

type Initializer struct {
	Config *orginit.Config

	Scheme *runtime.Scheme
	Client k8sclient.Client

	GetLabels func(instance v1.Object) map[string]string

	SecretManager *secretmanager.SecretManager
}

func NewInitializer(config *orginit.Config, scheme *runtime.Scheme, client k8sclient.Client, labels func(instance v1.Object) map[string]string) *Initializer {
	secretManager := secretmanager.New(client, scheme, labels)

	return &Initializer{
		Config:        config,
		Client:        client,
		Scheme:        scheme,
		GetLabels:     labels,
		SecretManager: secretManager,
	}
}

func (i *Initializer) GetAdminStoragePath(instance *current.Organization) string {
	return filepath.Join("/", i.Config.StoragePath, instance.GetName(), instance.Spec.Admin, util.GenerateRandomString(5))
}

func (i *Initializer) RemoveAdminStoragePath(dir string) error {
	err := os.RemoveAll(dir)
	if err != nil {
		return err
	}
	return nil
}

func (i *Initializer) ReadNodeOUConfigFile() ([]byte, error) {
	return util.ReadFile(i.Config.NodeOUConfigFile)
}

func (i *Initializer) ReconcileCrypto(instance *current.Organization) error {
	var err error
	err = i.WaitForCA(instance)
	if err != nil {
		return err
	}
	// Read CA's connection profile
	profile, err := i.GetCAConnectinProfile(instance)
	if err != nil {
		return err
	}
	// Eroll admin to get clientauth & tls certificates
	resp, err := i.EnrollAdmin(instance, profile)
	if err != nil {
		return err
	}

	// save admin clientauth&tls certs and org msp certs to secret
	clientAuthResp := resp.ClientAuth
	tlsResp := resp.TLS

	s := &corev1.Secret{}
	s.Name = instance.GetMSPCrypto().Name
	s.Namespace = instance.GetUserNamespace()
	s.Data = make(map[string][]byte)

	s.Data["admin-signcert"] = clientAuthResp.SignCert
	s.Data["admin-keystore"] = clientAuthResp.Keystore
	s.Data["admin-tls-signcert"] = tlsResp.SignCert
	s.Data["admin-tls-keystore"] = tlsResp.Keystore

	s.Data["org-ca-signcert"] = []byte(profile.CA.SignCerts)
	s.Data["org-tlsca-signcert"] = []byte(profile.TLSCA.SignCerts)

	nodeOUConfig, err := i.ReadNodeOUConfigFile()
	if err != nil {
		return err
	}
	s.Data["ou-config"] = nodeOUConfig

	err = i.Client.CreateOrUpdate(context.TODO(), s)
	if err != nil {
		return err
	}

	return nil
}
func (i *Initializer) WaitForCA(instance *current.Organization) error {
	var err error
	caDeploy := appsv1.Deployment{}
	deployment := instance.GetNamespaced()
	err = wait.Poll(10*time.Second, 10*maxRetryCount*time.Second, func() (bool, error) {
		log.Info(fmt.Sprintf("WaitForCA: poll deployment %s status", deployment.String()))
		err := i.Client.Get(context.TODO(), deployment, &caDeploy)
		if err != nil {
			if !k8serrors.IsNotFound(err) {
				return false, err
			}
			return false, nil
		}

		if caDeploy.Status.AvailableReplicas != *caDeploy.Spec.Replicas {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return errors.Errorf("exceed the maximum number of retry %d", maxRetryCount)
	}
	return nil
}
func (i *Initializer) EnrollAdmin(instance *current.Organization, profile *current.CAConnectionProfile) (*commonconfig.CryptoResponse, error) {
	cryptos := &commonconfig.Cryptos{}

	// client auth & tls in enrollment spec
	enrollmentSpec, err := i.GetAdminEnrollmentSpec(instance, profile)
	if err != nil {
		return nil, err
	}

	adminStoraePath := i.GetAdminStoragePath(instance)
	if err := common.GetCommonEnrollers(cryptos, enrollmentSpec, adminStoraePath); err != nil {
		return nil, err
	}
	defer i.RemoveAdminStoragePath(adminStoraePath)

	resp, err := cryptos.GenerateCryptoResponse()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (i *Initializer) GetAdminEnrollmentSpec(instance *current.Organization, profile *current.CAConnectionProfile) (*current.EnrollmentSpec, error) {
	caURL, err := url.Parse(profile.Endpoints.API)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse ca url")
	}

	return &current.EnrollmentSpec{
		ClientAuth: &current.Enrollment{
			CAName: "ca",
			CAHost: caURL.Hostname(),
			CAPort: caURL.Port(),
			CATLS: &current.CATLS{
				CACert: profile.TLS.Cert,
			},
			EnrollID:    instance.Spec.Admin,
			EnrollUser:  instance.Spec.Admin,
			EnrollToken: instance.Spec.AdminToken,
		},
		TLS: &current.Enrollment{
			CAName: "tlsca",
			CAHost: caURL.Hostname(),
			CAPort: caURL.Port(),
			CATLS: &current.CATLS{
				CACert: profile.TLS.Cert,
			},
			EnrollID:    instance.Spec.Admin,
			EnrollUser:  instance.Spec.Admin,
			EnrollToken: instance.Spec.AdminToken,
		},
	}, nil
}

func (i *Initializer) GetCAConnectinProfile(instance *current.Organization) (*current.CAConnectionProfile, error) {
	cm := &corev1.ConfigMap{}
	if err := i.Client.Get(context.TODO(), instance.GetCAConnectinProfile(), cm); err != nil {
		return nil, err
	}
	connectionProfile := &current.CAConnectionProfile{}
	if err := json.Unmarshal(cm.BinaryData["profile.json"], connectionProfile); err != nil {
		return nil, err
	}
	return connectionProfile, nil
}
