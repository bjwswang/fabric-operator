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
	"encoding/base64"
	"path/filepath"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/pkg/initializer/orderer/configtx"
	"github.com/IBM-Blockchain/fabric-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (i *Initializer) ConfigureApplication(instance *current.Channel, profile *configtx.Profile) error {
	for _, member := range instance.Spec.Members {
		organization, err := i.GetApplicationOrganization(instance, member.Name)
		if err != nil {
			return err
		}
		profile.Application.Organizations = append(profile.Application.Organizations, organization)
	}
	return nil
}

func (i *Initializer) GetApplicationOrganization(instance *current.Channel, member string) (*configtx.Organization, error) {
	var err error

	organization := &current.Organization{}
	err = i.Client.Get(context.TODO(), types.NamespacedName{Name: member}, organization)
	if err != nil {
		return nil, err
	}

	msp := &corev1.Secret{}
	err = i.Client.Get(context.TODO(), organization.GetMSPCrypto(), msp)
	if err != nil {
		return nil, err
	}

	// save cers under /msp/dir
	caCerts := msp.Data["org-ca-signcert"]
	caCertPem, err := base64.StdEncoding.DecodeString(string(caCerts))
	if err != nil {
		return nil, err
	}
	err = util.WriteFile(filepath.Join(i.GetOrgMSPDir(instance, member), "cacerts", "ca-signcert.pem"), caCertPem, 0777)
	if err != nil {
		return nil, err
	}
	// caInterCerts := msp.Data["caintercerts"]
	// err = util.WriteFile(filepath.Join(i.GetOrgMSPDir(instance, member), "intermidiate", "intermidiate_ca.pem"), caInterCerts, 0777)
	// if err != nil {
	// 	return nil, err
	// }
	tlscaCerts := msp.Data["org-tlsca-signcert"]
	tlscaCertPem, err := base64.StdEncoding.DecodeString(string(tlscaCerts))
	if err != nil {
		return nil, err
	}
	err = util.WriteFile(filepath.Join(i.GetOrgMSPDir(instance, member), "tlscacerts", "tlsca-signcert.pem"), tlscaCertPem, 0777)
	if err != nil {
		return nil, err
	}
	// tlscaInterCerts := msp.Data["caintercerts"]
	// err = util.WriteFile(filepath.Join(i.GetOrgMSPDir(instance, member), "intermidiate_tlsca_certs", "intermidiate_ca.pem"), tlscaInterCerts, 0777)
	// if err != nil {
	// 	return nil, err
	// }

	ouConfig := msp.Data["ou-config"]
	err = util.WriteFile(filepath.Join(i.GetOrgMSPDir(instance, member), "config.yaml"), ouConfig, 0777)
	if err != nil {
		return nil, err
	}

	org := configtx.DefaultOrganization(organization.GetName())
	org.MSPDir = i.GetOrgMSPDir(instance, member)

	return org, nil
}
