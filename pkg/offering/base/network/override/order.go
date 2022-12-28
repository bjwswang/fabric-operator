package override

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources"
	"github.com/IBM-Blockchain/fabric-operator/pkg/util"
	"github.com/IBM-Blockchain/fabric-operator/pkg/util/pointer"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (o *Override) Orderer(object v1.Object, orderer *current.IBPOrderer, action resources.Action) error {
	instance := object.(*current.Network)
	switch action {
	case resources.Create, resources.Update:
		return o.CreateOrUpdateOrderer(instance, orderer)
	}

	return nil
}

func (o *Override) CreateOrUpdateOrderer(instance *current.Network, orderer *current.IBPOrderer) (err error) {
	orderer.Namespace = instance.GetInitiatorMember().Namespace
	orderer.Name = instance.Name
	orderer.Spec = instance.Spec.OrderSpec

	orderer.Spec.Domain = o.IngressDomain
	orderer.Spec.MSPID = instance.Name
	if orderer.Spec.NodeNumber == nil {
		n := 1
		orderer.Spec.NodeNumber = &n
	}
	if orderer.Spec.UseChannelLess == nil {
		orderer.Spec.UseChannelLess = pointer.True()
	}
	if orderer.Spec.FabricVersion == "" {
		if orderer.Spec.Images != nil {
			if ordererTag := orderer.Spec.Images.OrdererTag; ordererTag != "" && ordererTag != "latest" {
				orderer.Spec.FabricVersion = ordererTag
			}
		}
		if orderer.Spec.FabricVersion == "" {
			orderer.Spec.FabricVersion = "2.4.7"
		}
	}
	if orderer.Spec.SystemChannelName == "" {
		orderer.Spec.SystemChannelName = instance.Name
	}
	orderer.Spec.OrgName = instance.GetInitiatorMember().Name

	err = o.updateEnrollment(instance, orderer)
	if err != nil {
		return err
	}
	orderer.OwnerReferences = []v1.OwnerReference{
		{
			Kind:       "Network",
			APIVersion: "ibp.com/v1beta1",
			Name:       instance.GetName(),
			UID:        instance.GetUID(),
		},
	}

	return nil
}

func (o *Override) updateEnrollment(instance *current.Network, orderer *current.IBPOrderer) (err error) {
	profile, err := o.getCAConnectionProfileData(instance.GetInitiatorMember().Namespace, instance.GetInitiatorMember().Name)
	if err != nil {
		return errors.Wrap(err, "failed to get ca cm connection-profile")
	}

	user, err := o.getEnrollUser(instance.GetInitiatorMember().Namespace, instance.GetInitiatorMember().Name)
	if err != nil {
		return errors.Wrap(err, "failed to get enroll user")
	}
	caURL, err := url.Parse(profile.Endpoints.API)
	if err != nil {
		return errors.Wrap(err, "failed to parse ca url")
	}
	host, _, err := net.SplitHostPort(caURL.Host)
	if err != nil {
		return err
	}
	if orderer.Spec.Secret == nil {
		orderer.Spec.Secret = &current.SecretSpec{}
	}
	if orderer.Spec.Secret.Enrollment == nil {
		orderer.Spec.Secret.Enrollment = &current.EnrollmentSpec{}
	}

	if orderer.Spec.Secret.Enrollment.Component == nil {
		orderer.Spec.Secret.Enrollment.Component = &current.Enrollment{}
	}
	if orderer.Spec.Secret.Enrollment.Component.CAHost == "" {
		orderer.Spec.Secret.Enrollment.Component.CAHost = host
	}
	if orderer.Spec.Secret.Enrollment.Component.CAPort == "" {
		orderer.Spec.Secret.Enrollment.Component.CAPort = caURL.Port()
	}
	if orderer.Spec.Secret.Enrollment.Component.CAName == "" {
		orderer.Spec.Secret.Enrollment.Component.CAName = "ca"
	}
	if orderer.Spec.Secret.Enrollment.Component.CATLS == nil {
		orderer.Spec.Secret.Enrollment.Component.CATLS = &current.CATLS{CACert: profile.TLS.Cert}
	}
	if orderer.Spec.Secret.Enrollment.Component.EnrollID == "" {
		orderer.Spec.Secret.Enrollment.Component.EnrollID = instance.Name
	}
	if orderer.Spec.Secret.Enrollment.Component.EnrollSecret == "" {
		orderer.Spec.Secret.Enrollment.Component.EnrollSecret = instance.GetInitiatorMember().Name
	}
	if orderer.Spec.Secret.Enrollment.Component.EnrollUser == "" {
		orderer.Spec.Secret.Enrollment.Component.EnrollUser = user
	}

	if orderer.Spec.Secret.Enrollment.TLS == nil {
		orderer.Spec.Secret.Enrollment.TLS = &current.Enrollment{}
	}
	if orderer.Spec.Secret.Enrollment.TLS.CAHost == "" {
		orderer.Spec.Secret.Enrollment.TLS.CAHost = host
	}
	if orderer.Spec.Secret.Enrollment.TLS.CAPort == "" {
		orderer.Spec.Secret.Enrollment.TLS.CAPort = caURL.Port()
	}
	if orderer.Spec.Secret.Enrollment.TLS.CAName == "" {
		orderer.Spec.Secret.Enrollment.TLS.CAName = "ca"
	}
	if orderer.Spec.Secret.Enrollment.TLS.CATLS == nil {
		orderer.Spec.Secret.Enrollment.TLS.CATLS = &current.CATLS{CACert: profile.TLS.Cert}
	}
	if orderer.Spec.Secret.Enrollment.TLS.EnrollID == "" {
		orderer.Spec.Secret.Enrollment.TLS.EnrollID = instance.Name
	}
	if orderer.Spec.Secret.Enrollment.TLS.EnrollSecret == "" {
		orderer.Spec.Secret.Enrollment.TLS.EnrollSecret = instance.GetInitiatorMember().Name
	}
	if orderer.Spec.Secret.Enrollment.TLS.EnrollUser == "" {
		orderer.Spec.Secret.Enrollment.TLS.EnrollUser = user
	}
	if orderer.Spec.Secret.Enrollment.TLS.CSR == nil {
		orderer.Spec.Secret.Enrollment.TLS.CSR = &current.CSR{Hosts: make([]string, 0)}
	}
	clusterHosts := []string{
		instance.GetInitiatorMember().Name,
		instance.GetInitiatorMember().Name + "." + instance.GetInitiatorMember().Namespace,
		instance.GetInitiatorMember().Name + "." + instance.GetInitiatorMember().Namespace + ".svc.cluster.local",
	}
	hosts := orderer.Spec.Secret.Enrollment.TLS.CSR.Hosts
	for _, h := range clusterHosts {
		hosts = util.AppendStringIfMissing(hosts, h)
	}
	orderer.Spec.Secret.Enrollment.TLS.CSR.Hosts = hosts
	return nil
}

func (o *Override) getCAConnectionProfileData(namespace, name string) (*current.CAConnectionProfile, error) {
	cm := &corev1.ConfigMap{}
	name = name + "-connection-profile"
	if err := o.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, cm); err != nil {
		return nil, err
	}
	data := cm.BinaryData["profile.json"]
	if data == nil {
		return nil, fmt.Errorf("no profile.json in cm:%s in ns:%s", name, namespace)
	}
	profile := &current.CAConnectionProfile{}
	if err := json.Unmarshal(data, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func (o *Override) getEnrollUser(namespace, name string) (user string, err error) {
	org := &current.Organization{}
	if err := o.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, org); err != nil {
		return "", err
	}
	return org.Spec.Admin, nil
}
