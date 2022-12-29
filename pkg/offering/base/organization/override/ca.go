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

package override

import (
	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/pkg/initializer/ca/config"
	"github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources"
	"github.com/IBM-Blockchain/fabric-operator/pkg/util"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	cav1 "github.com/IBM-Blockchain/fabric-operator/pkg/apis/ca/v1"
)

func (o *Override) CertificateAuthority(object v1.Object, ca *current.IBPCA, action resources.Action) error {
	instance := object.(*current.Organization)
	switch action {
	case resources.Create, resources.Update:
		return o.CreateOrUpdateCA(instance, ca)
	}

	return nil
}

func (o *Override) CreateOrUpdateCA(instance *current.Organization, ca *current.IBPCA) error {
	var err error
	namespaced := instance.GetCA()
	ca.Namespace = namespaced.Namespace
	ca.Name = namespaced.Name

	// merge with override
	ca.Spec = instance.Spec.CASpec

	ca.Spec.Domain = o.IngressDomain

	if o.IAMEnabled {
		err = o.OverrideCAConfig(instance, ca)
		if err != nil {
			return err
		}
		err = o.OverrideTLSCAConfig(instance, ca)
		if err != nil {
			return err
		}
	}

	ca.OwnerReferences = []v1.OwnerReference{
		{
			Kind:       "Organization",
			APIVersion: "ibp.com/v1beta1",
			Name:       instance.GetName(),
			UID:        instance.GetUID(),
		},
	}

	return nil
}

func (o *Override) OverrideCAConfig(instance *current.Organization, ca *current.IBPCA) error {
	if ca.Spec.ConfigOverride == nil {
		ca.Spec.ConfigOverride = &current.ConfigOverride{}
	}

	if ca.Spec.ConfigOverride.CA == nil {
		ca.Spec.ConfigOverride.CA = &runtime.RawExtension{}
	}

	configOverride, err := config.ReadFrom(&ca.Spec.ConfigOverride.CA.Raw)
	if err != nil {
		configOverride = &config.Config{
			ServerConfig: &cav1.ServerConfig{},
		}
	}

	configOverride.ServerConfig.CAConfig.IAM.Enabled = &o.IAMEnabled
	configOverride.ServerConfig.CAConfig.IAM.URL = o.IAMServer

	configOverride.ServerConfig.CAConfig.Organization = instance.GetName()

	raw, err := util.ConvertToJsonMessage(configOverride.ServerConfig)
	if err != nil {
		return err
	}

	ca.Spec.ConfigOverride.CA.Raw = *raw

	return nil
}

func (o *Override) OverrideTLSCAConfig(instance *current.Organization, ca *current.IBPCA) error {
	if ca.Spec.ConfigOverride == nil {
		ca.Spec.ConfigOverride = &current.ConfigOverride{}
	}

	if ca.Spec.ConfigOverride.TLSCA == nil {
		ca.Spec.ConfigOverride.TLSCA = &runtime.RawExtension{}
	}

	configOverride, err := config.ReadFrom(&ca.Spec.ConfigOverride.TLSCA.Raw)
	if err != nil {
		configOverride = &config.Config{
			ServerConfig: &cav1.ServerConfig{},
		}
	}

	configOverride.ServerConfig.CAConfig.IAM.Enabled = &o.IAMEnabled
	configOverride.ServerConfig.CAConfig.IAM.URL = o.IAMServer

	configOverride.ServerConfig.CAConfig.Organization = instance.GetName()

	raw, err := util.ConvertToJsonMessage(configOverride.ServerConfig)
	if err != nil {
		return err
	}

	ca.Spec.ConfigOverride.TLSCA.Raw = *raw

	return nil
}
