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

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/pkg/initializer/orderer/configtx"
	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/orderer/etcdraft"
	"github.com/hyperledger/fabric/bccsp"
	fmsp "github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (i *Initializer) ConfigureOrderer(instance *current.Channel, profile *configtx.Profile, ordererorg string, parentOrderer *current.IBPOrderer, clusterNodes *current.IBPOrdererList) (map[string]*msp.MSPConfig, error) {
	var err error
	err = i.AddHostPortToProfile(profile, parentOrderer, clusterNodes)
	if err != nil {
		return nil, err
	}

	org := configtx.DefaultOrganization(ordererorg)
	org.MSPDir = i.GetOrgMSPDir(instance, ordererorg)
	err = profile.AddOrgToOrderer(org)
	if err != nil {
		return nil, err
	}

	conf := profile.Orderer
	mspConfigs := map[string]*msp.MSPConfig{}
	// only one orderer organization for now
	for _, org := range conf.Organizations {
		mspConfigs[org.Name], err = i.GetOrdererMSPConfig(parentOrderer, org.ID)
		if err != nil {
			return nil, err
		}
	}
	return mspConfigs, nil
}

func (i *Initializer) AddHostPortToProfile(profile *configtx.Profile, parent *current.IBPOrderer, clusterNodes *current.IBPOrdererList) error {
	log.Info("Adding hosts to genesis block")
	ns := parent.GetNamespace()
	parentName := parent.GetName()

	for _, node := range clusterNodes.Items {
		if node.Status.Type != current.Deployed {
			return errors.Errorf("consensus node {name:%s,namespace:%s} not deployed yet", node.GetName(), node.GetNamespace())
		}

		n := types.NamespacedName{
			Name:      fmt.Sprintf("tls-%s-signcert", node.Name),
			Namespace: ns,
		}

		// To avoid the race condition of the TLS signcert secret not existing, need to poll for it's
		// existence before proceeding
		tlsSecret := &corev1.Secret{}
		err := i.Client.Get(context.TODO(), n, tlsSecret)
		if err != nil {
			return errors.Wrapf(err, "failed to find secret '%s'", n.Name)
		}

		nodeName := fmt.Sprintf("node%d", *node.Spec.NodeNumber)
		domain := node.Spec.Domain
		fqdn := ns + "-" + parentName + nodeName + "-orderer" + "." + domain

		log.Info(fmt.Sprintf("Adding consentor domain '%s' to genesis block", fqdn))

		profile.AddOrdererAddress(fmt.Sprintf("%s:%d", fqdn, 443))
		consentors := &etcdraft.Consenter{
			Host:          fqdn,
			Port:          443,
			ClientTlsCert: tlsSecret.Data["cert.pem"],
			ServerTlsCert: tlsSecret.Data["cert.pem"],
		}
		err = profile.AddRaftConsentingNode(consentors)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Initializer) GetOrdererMSPConfig(instance *current.IBPOrderer, ID string) (*msp.MSPConfig, error) {
	isIntermediate := false
	admincert := [][]byte{}
	n := types.NamespacedName{
		Name:      fmt.Sprintf("ecert-%s%s%d-admincerts", instance.Name, NODE, 1),
		Namespace: instance.Namespace,
	}
	adminCert := &corev1.Secret{}
	err := i.Client.Get(context.TODO(), n, adminCert)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, err
		}
	}
	for _, cert := range adminCert.Data {
		admincert = append(admincert, cert)
	}

	cacerts := [][]byte{}
	n.Name = fmt.Sprintf("ecert-%s%s%d-cacerts", instance.Name, NODE, 1)
	caCerts := &corev1.Secret{}
	err = i.Client.Get(context.TODO(), n, caCerts)
	if err != nil {
		return nil, err
	}
	for _, cert := range caCerts.Data {
		cacerts = append(cacerts, cert)
	}

	intermediateCerts := [][]byte{}
	interCerts := &corev1.Secret{}
	n.Name = fmt.Sprintf("ecert-%s%s%d-intercerts", instance.Name, NODE, 1)
	err = i.Client.Get(context.TODO(), n, interCerts)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, err
		}
	}
	for _, cert := range interCerts.Data {
		isIntermediate = true
		intermediateCerts = append(intermediateCerts, cert)
	}

	cryptoConfig := &msp.FabricCryptoConfig{
		SignatureHashFamily:            bccsp.SHA2,
		IdentityIdentifierHashFunction: bccsp.SHA256,
	}

	tlsCACerts := [][]byte{}
	n.Name = fmt.Sprintf("tls-%s%s%d-cacerts", instance.Name, NODE, 1)
	tlsCerts := &corev1.Secret{}
	err = i.Client.Get(context.TODO(), n, tlsCerts)
	if err != nil {
		return nil, err
	}
	for _, cert := range tlsCerts.Data {
		tlsCACerts = append(tlsCACerts, cert)
	}

	tlsIntermediateCerts := [][]byte{}
	tlsInterCerts := &corev1.Secret{}
	n.Name = fmt.Sprintf("tls-%s%s%d-intercerts", instance.Name, NODE, 1)
	err = i.Client.Get(context.TODO(), n, tlsInterCerts)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, err
		}
	}
	for _, cert := range tlsInterCerts.Data {
		tlsIntermediateCerts = append(tlsIntermediateCerts, cert)
	}

	fmspconf := &msp.FabricMSPConfig{
		Admins:               admincert,
		RootCerts:            cacerts,
		IntermediateCerts:    intermediateCerts,
		Name:                 ID,
		CryptoConfig:         cryptoConfig,
		TlsRootCerts:         tlsCACerts,
		TlsIntermediateCerts: tlsIntermediateCerts,
		FabricNodeOus: &msp.FabricNodeOUs{
			Enable: true,
			ClientOuIdentifier: &msp.FabricOUIdentifier{
				OrganizationalUnitIdentifier: "client",
				Certificate:                  cacerts[0],
			},
			PeerOuIdentifier: &msp.FabricOUIdentifier{
				OrganizationalUnitIdentifier: "peer",
				Certificate:                  cacerts[0],
			},
			AdminOuIdentifier: &msp.FabricOUIdentifier{
				OrganizationalUnitIdentifier: "admin",
				Certificate:                  cacerts[0],
			},
			OrdererOuIdentifier: &msp.FabricOUIdentifier{
				OrganizationalUnitIdentifier: "orderer",
				Certificate:                  cacerts[0],
			},
		},
	}

	if isIntermediate {
		fmspconf.FabricNodeOus.ClientOuIdentifier.Certificate = intermediateCerts[0]
		fmspconf.FabricNodeOus.PeerOuIdentifier.Certificate = intermediateCerts[0]
		fmspconf.FabricNodeOus.AdminOuIdentifier.Certificate = intermediateCerts[0]
		fmspconf.FabricNodeOus.OrdererOuIdentifier.Certificate = intermediateCerts[0]
	}

	fmpsjs, err := proto.Marshal(fmspconf)
	if err != nil {
		return nil, err
	}

	mspconf := &msp.MSPConfig{Config: fmpsjs, Type: int32(fmsp.FABRIC)}

	return mspconf, nil
}
