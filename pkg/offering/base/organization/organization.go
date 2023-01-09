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
	"fmt"
	"strings"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	config "github.com/IBM-Blockchain/fabric-operator/operatorconfig"
	"github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources"
	"github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources/manager"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/base/organization/override"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/common"
	"github.com/IBM-Blockchain/fabric-operator/pkg/operatorerrors"
	bcrbac "github.com/IBM-Blockchain/fabric-operator/pkg/rbac"
	"github.com/IBM-Blockchain/fabric-operator/pkg/user"
	"github.com/IBM-Blockchain/fabric-operator/version"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("base_organization")

//go:generate counterfeiter -o mocks/update.go -fake-name Update . Update

type Update interface {
	SpecUpdated() bool
	AdminUpdated() bool
	TokenUpdated() bool
	AdminTransfered() string
	ClientsUpdated() bool
	ClientsRemoved() string
}

type Override interface {
	AdminRole(v1.Object, *rbacv1.Role, resources.Action) error
	ClientRole(v1.Object, *rbacv1.Role, resources.Action) error
	AdminRoleBinding(v1.Object, *rbacv1.RoleBinding, resources.Action) error
	ClientRoleBinding(v1.Object, *rbacv1.RoleBinding, resources.Action) error

	AdminClusterRole(v1.Object, *rbacv1.ClusterRole, resources.Action) error
	AdminClusterRoleBinding(v1.Object, *rbacv1.ClusterRoleBinding, resources.Action) error
	ClientClusterRole(v1.Object, *rbacv1.ClusterRole, resources.Action) error
	ClientClusterRoleBinding(v1.Object, *rbacv1.ClusterRoleBinding, resources.Action) error

	CertificateAuthority(v1.Object, *current.IBPCA, resources.Action) error
}

//go:generate counterfeiter -o mocks/initializer.go -fake-name InitializerOrganization . InitializerOrganization

type InitializerOrganization interface{}

type Organization interface {
	PreReconcileChecks(instance *current.Organization, update Update) error
	Initialize(instance *current.Organization, update Update) error
	ReconcileManagers(instance *current.Organization, update Update) error
	CheckStates(instance *current.Organization) (common.Result, error)
	Reconcile(instance *current.Organization, update Update) (common.Result, error)
}

var _ Organization = (*BaseOrganization)(nil)

const (
	KIND = "ORGANIZATION"
)

type BaseOrganization struct {
	Client controllerclient.Client
	Scheme *runtime.Scheme

	Config *config.Config

	Initializer InitializerOrganization

	Override Override

	AdminRoleManager         resources.Manager
	AdminRoleBindingManager  resources.Manager
	ClientRoleManager        resources.Manager
	ClientRoleBindingManager resources.Manager

	AdminClusterRoleManager         resources.Manager
	AdminClusterRoleBindingManager  resources.Manager
	ClientClusterRoleManager        resources.Manager
	ClientClusterRoleBindingManager resources.Manager

	CAManager resources.Manager
}

func New(client controllerclient.Client, scheme *runtime.Scheme, config *config.Config) *BaseOrganization {
	o := &override.Override{
		Client:        client,
		IngressDomain: config.Operator.IngressDomain,
		IAMEnabled:    config.Operator.IAM.Enabled,
		IAMServer:     config.Operator.IAM.Server,
	}
	base := &BaseOrganization{
		Client:   client,
		Scheme:   scheme,
		Config:   config,
		Override: o,
	}

	base.Initializer = NewInitializer(config.OrganizationInitConfig, scheme, client, base.GetLabels)

	base.CreateManagers()

	return base
}

func (organization *BaseOrganization) CreateManagers() {
	override := organization.Override
	mgr := manager.New(organization.Client, organization.Scheme)

	organization.AdminRoleManager = mgr.CreateRoleManager(bcrbac.AdminSuffix, override.AdminRole, organization.GetLabels, organization.Config.OrganizationInitConfig.AdminRoleFile)
	organization.AdminRoleBindingManager = mgr.CreateRoleBindingManager(bcrbac.AdminSuffix, override.AdminRoleBinding, organization.GetLabels, organization.Config.OrganizationInitConfig.RoleBindingFile)

	organization.ClientRoleManager = mgr.CreateRoleManager(bcrbac.ClientSuffix, override.ClientRole, organization.GetLabels, organization.Config.OrganizationInitConfig.ClientRoleFile)
	organization.ClientRoleBindingManager = mgr.CreateRoleBindingManager(bcrbac.ClientSuffix, override.ClientRoleBinding, organization.GetLabels, organization.Config.OrganizationInitConfig.RoleBindingFile)

	organization.AdminClusterRoleManager = mgr.CreateClusterRoleManager(bcrbac.AdminSuffix, override.AdminClusterRole, organization.GetLabels, organization.Config.OrganizationInitConfig.ClusterRoleFile)
	organization.AdminClusterRoleBindingManager = mgr.CreateClusterRoleBindingManager(bcrbac.AdminSuffix, override.AdminClusterRoleBinding, organization.GetLabels, organization.Config.OrganizationInitConfig.ClusterRoleBindingFile)

	organization.ClientClusterRoleManager = mgr.CreateClusterRoleManager(bcrbac.ClientSuffix, override.ClientClusterRole, organization.GetLabels, organization.Config.OrganizationInitConfig.ClusterRoleFile)
	organization.ClientClusterRoleBindingManager = mgr.CreateClusterRoleBindingManager(bcrbac.ClientSuffix, override.ClientClusterRoleBinding, organization.GetLabels, organization.Config.OrganizationInitConfig.ClusterRoleBindingFile)

	organization.CAManager = mgr.CreateCAManager("", override.CertificateAuthority, organization.GetLabels, organization.Config.OrganizationInitConfig.CAFile)
}

// Reconcile on Organization upon Update
func (organization *BaseOrganization) Reconcile(instance *current.Organization, update Update) (common.Result, error) {
	var err error

	if err = organization.PreReconcileChecks(instance, update); err != nil {
		return common.Result{}, errors.Wrap(err, "failed on prereconcile checks")
	}

	if err = organization.Initialize(instance, update); err != nil {
		return common.Result{}, operatorerrors.Wrap(err, operatorerrors.OrganizationInitilizationFailed, "failed to initialize organization")
	}

	// TODO: define managers
	if err = organization.ReconcileManagers(instance, update); err != nil {
		return common.Result{}, errors.Wrap(err, "failed to reconcile managers")
	}

	return organization.CheckStates(instance)
}

// PreReconcileChecks on Organization upon Update
func (organization *BaseOrganization) PreReconcileChecks(instance *current.Organization, update Update) error {
	log.Info(fmt.Sprintf("PreReconcileChecks on Organization %s", instance.GetName()))

	if !instance.HasAdmin() {
		return errors.New("organization admin is empty")
	}

	return nil
}

// Initialize on Organization upon Update
func (organization *BaseOrganization) Initialize(instance *current.Organization, update Update) error {
	log.Info(fmt.Sprintf("Checking if organization '%s' needs initialization", instance.GetName()))
	return nil
}

// ReconcileManagers on Organization upon Update
func (organization *BaseOrganization) ReconcileManagers(instance *current.Organization, update Update) error {
	var err error

	err = organization.CreateNamespace(instance)
	if err != nil {
		return err
	}

	// Deploy CA
	err = organization.CAManager.Reconcile(instance, true)
	if err != nil {
		return err
	}

	err = organization.ReconcileRBAC(instance, update)
	if err != nil {
		return err
	}

	err = organization.ReconcileUsers(instance, update)
	if err != nil {
		return err
	}

	return nil
}

// RecnocleRBAC on current organization,including:
// - Create admin role and client role
// - Create admin/client clusterrole if not exists
// - Create/update admin clusterrole binding if admin updated
// - Create/update client clusterrole binding if clients updated
func (organization *BaseOrganization) ReconcileRBAC(instance *current.Organization, update Update) error {
	var err error

	// Create AdminRole if not exist
	err = organization.AdminRoleManager.Reconcile(instance, false)
	if err != nil {
		return err
	}
	// Create AdminClusterRole if not exist
	err = organization.AdminClusterRoleManager.Reconcile(instance, false)
	if err != nil {
		return err
	}

	// Create ClientRole if not exist
	err = organization.ClientRoleManager.Reconcile(instance, false)
	if err != nil {
		return err
	}
	// Create ClientClusterRole if not exist
	err = organization.ClientClusterRoleManager.Reconcile(instance, false)
	if err != nil {
		return err
	}

	if update.AdminUpdated() {
		// reconcile AdminRoleBinding
		err = organization.AdminRoleBindingManager.Reconcile(instance, true)
		if err != nil {
			return err
		}

		// reconcile AdminClusterRoleBinding
		err = organization.AdminClusterRoleBindingManager.Reconcile(instance, true)
		if err != nil {
			return err
		}
	}

	if update.ClientsUpdated() {
		// reconcile ClientRoleBinding
		err = organization.ClientRoleBindingManager.Reconcile(instance, true)
		if err != nil {
			return err
		}
		// reconcile ClientClusterRoleBinding
		err = organization.ClientClusterRoleBindingManager.Reconcile(instance, true)
		if err != nil {
			return err
		}
	}

	return nil
}

// ReconcileUsers handles User's annotation change
func (organization *BaseOrganization) ReconcileUsers(instance *current.Organization, update Update) error {
	var err error

	// Set/Transfer Admin annotations
	if update.AdminUpdated() && organization.Config.OrganizationInitConfig.IAMEnabled {
		targetUser := instance.Spec.Admin
		transferFrom := update.AdminTransfered()
		if transferFrom != "" {
			err = user.ReconcileTransfer(organization.Client, transferFrom, targetUser, instance.GetName())
			if err != nil {
				return err
			}
		} else {
			err = user.Reconcile(organization.Client, targetUser, instance.GetName(), "", user.ADMIN, user.Add)
			if err != nil {
				return err
			}
		}
	}

	if update.ClientsUpdated() && organization.Config.OrganizationInitConfig.IAMEnabled {
		// reconcile user set
		for _, c := range instance.Spec.Clients {
			err = user.Reconcile(organization.Client, c, instance.GetName(), "", user.CLIENT, user.Add)
			if err != nil {
				return err
			}
		}

		// reconcile user remove
		if update.ClientsRemoved() != "" {
			removed := strings.Split(update.ClientsRemoved(), ",")
			for _, c := range removed {
				err = user.Reconcile(organization.Client, c, instance.GetName(), "", user.CLIENT, user.Remove)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// CheckStates on Organization
func (organization *BaseOrganization) CheckStates(instance *current.Organization) (common.Result, error) {
	return common.Result{
		Status: &current.CRStatus{
			Type:    current.Deploying,
			Version: version.Operator,
		},
	}, nil
}

// GetLabels from instance.GetLabels
func (organization *BaseOrganization) GetLabels(instance v1.Object) map[string]string {
	return instance.GetLabels()
}

func (organization *BaseOrganization) CreateNamespace(instance *current.Organization) error {
	ns := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name:   instance.GetUserNamespace(),
			Labels: instance.GetLabels(),
		},
	}
	ns.OwnerReferences = []v1.OwnerReference{bcrbac.OwnerReference(bcrbac.Organization, instance)}
	return organization.Client.CreateOrUpdate(context.TODO(), &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name:   instance.GetUserNamespace(),
			Labels: instance.GetLabels(),
		},
	})
}
