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
	initiatorOrg := &current.Organization{ObjectMeta: v1.ObjectMeta{Name: instance.GetInitiatorMember()}}
	initiatorNamespace := initiatorOrg.GetUserNamespace()
	orderer.Namespace = initiatorNamespace
	orderer.Name = instance.Name
	orderer.Spec = instance.Spec.OrderSpec

	orderer.Spec.Domain = o.IngressDomain
	orderer.Spec.MSPID = instance.GetInitiatorMember()
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
	orderer.Spec.OrgName = instance.GetInitiatorMember()

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
	initiatorOrg := &current.Organization{ObjectMeta: v1.ObjectMeta{Name: instance.GetInitiatorMember()}}
	initiatorNamespace := initiatorOrg.GetUserNamespace()
	profile, err := o.getCAConnectionProfileData(initiatorNamespace, instance.GetInitiatorMember())
	if err != nil {
		return errors.Wrap(err, "failed to get ca cm connection-profile")
	}

	user, err := o.getEnrollUser(initiatorNamespace, instance.GetInitiatorMember())
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
	if orderer.Spec.ClusterSecret == nil {
		orderer.Spec.ClusterSecret = make([]*current.SecretSpec, instance.Spec.OrderSpec.ClusterSize)
	}
	for i := range orderer.Spec.ClusterSecret {
		if orderer.Spec.ClusterSecret[i] == nil {
			orderer.Spec.ClusterSecret[i] = &current.SecretSpec{}
		}
		v := orderer.Spec.ClusterSecret[i]
		enrollID := fmt.Sprintf("%s%d", instance.Name, i)
		if v.Enrollment == nil {
			v.Enrollment = &current.EnrollmentSpec{}
		}
		if v.Enrollment.Component == nil {
			v.Enrollment.Component = &current.Enrollment{}
		}
		if v.Enrollment.Component.CAHost == "" {
			v.Enrollment.Component.CAHost = host
		}
		if v.Enrollment.Component.CAPort == "" {
			v.Enrollment.Component.CAPort = caURL.Port()
		}
		if v.Enrollment.Component.CAName == "" {
			v.Enrollment.Component.CAName = "ca"
		}
		if v.Enrollment.Component.CATLS == nil {
			v.Enrollment.Component.CATLS = &current.CATLS{CACert: profile.TLS.Cert}
		}
		if v.Enrollment.Component.EnrollID == "" {
			v.Enrollment.Component.EnrollID = enrollID
		}
		if v.Enrollment.Component.EnrollToken == "" {
			v.Enrollment.Component.EnrollToken = instance.Spec.InitialToken
		}
		if v.Enrollment.Component.EnrollSecret == "" {
			v.Enrollment.Component.EnrollSecret = enrollID
		}
		if v.Enrollment.Component.EnrollUser == "" {
			v.Enrollment.Component.EnrollUser = user
		}

		if v.Enrollment.TLS == nil {
			v.Enrollment.TLS = &current.Enrollment{}
		}
		if v.Enrollment.TLS.CAHost == "" {
			v.Enrollment.TLS.CAHost = host
		}
		if v.Enrollment.TLS.CAPort == "" {
			v.Enrollment.TLS.CAPort = caURL.Port()
		}
		if v.Enrollment.TLS.CAName == "" {
			v.Enrollment.TLS.CAName = "ca"
		}
		if v.Enrollment.TLS.CATLS == nil {
			v.Enrollment.TLS.CATLS = &current.CATLS{CACert: profile.TLS.Cert}
		}
		if v.Enrollment.TLS.EnrollID == "" {
			v.Enrollment.TLS.EnrollID = enrollID
		}
		if v.Enrollment.TLS.EnrollToken == "" {
			v.Enrollment.TLS.EnrollToken = instance.Spec.InitialToken
		}
		if v.Enrollment.TLS.EnrollSecret == "" {
			v.Enrollment.TLS.EnrollSecret = enrollID
		}
		if v.Enrollment.TLS.EnrollUser == "" {
			v.Enrollment.TLS.EnrollUser = user
		}
		if v.Enrollment.TLS.CSR == nil {
			v.Enrollment.TLS.CSR = &current.CSR{Hosts: make([]string, 0)}
		}
		initiatorOrg := &current.Organization{ObjectMeta: v1.ObjectMeta{Name: instance.GetInitiatorMember()}}
		initiatorNamespace := initiatorOrg.GetUserNamespace()
		clusterHosts := []string{
			instance.GetInitiatorMember(),
			instance.GetInitiatorMember() + "." + initiatorNamespace,
			instance.GetInitiatorMember() + "." + initiatorNamespace + ".svc.cluster.local",
		}
		hosts := v.Enrollment.TLS.CSR.Hosts
		for _, h := range clusterHosts {
			hosts = util.AppendStringIfMissing(hosts, h)
		}
		v.Enrollment.TLS.CSR.Hosts = hosts
		orderer.Spec.ClusterSecret[i] = v
	}

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
