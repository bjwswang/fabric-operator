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
	"reflect"
	"strings"

	iam "github.com/IBM-Blockchain/fabric-operator/api/iam/v1alpha1"
	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	k8sclient "github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/user"
	"github.com/go-test/deep"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func (r *ReconcileOrganization) CreateFunc(e event.CreateEvent) bool {
	var reconcile bool
	organization := e.Object.(*current.Organization)
	log.Info(fmt.Sprintf("Create event detected for organization '%s'", organization.GetName()))
	reconcile = r.PredictOrganizationCreate(organization)
	if reconcile {
		log.Info(fmt.Sprintf("Create event triggering reconcile for creating organization '%s'", organization.GetName()))
	}
	return reconcile
}

func (r *ReconcileOrganization) PredictOrganizationCreate(organization *current.Organization) bool {
	update := Update{}
	if organization.HasType() {
		log.Info(fmt.Sprintf("Operator restart detected, running update flow on existing organization '%s'", organization.GetName()))

		cm, err := r.GetSpecState(organization)
		if err != nil {
			log.Info(fmt.Sprintf("Failed getting saved organization spec '%s', triggering create: %s", organization.GetName(), err.Error()))
			return true
		}

		specBytes := cm.BinaryData["spec"]
		existingOrg := &current.Organization{}
		err = yaml.Unmarshal(specBytes, &existingOrg.Spec)
		if err != nil {
			log.Info(fmt.Sprintf("Unmarshal failed for saved organization spec '%s', triggering create: %s", organization.GetName(), err.Error()))
			return true
		}

		diff := deep.Equal(organization.Spec, existingOrg.Spec)
		if diff != nil {
			log.Info(fmt.Sprintf("Organization '%s' spec was updated while operator was down", organization.GetName()))
			log.Info(fmt.Sprintf("Difference detected: %s", diff))
			update.specUpdated = true
		}

		if organization.Spec.Admin != existingOrg.Spec.Admin {
			update.adminUpdated = true
			update.adminTransfered = existingOrg.Spec.Admin
		}

		// https://github.com/bestchains/fabric-operator/issues/14#issuecomment-1371948917
		if organization.Spec.AdminToken != existingOrg.Spec.AdminToken && organization.Spec.AdminToken != "" {
			update.tokenUpdated = update.adminUpdated
			if !update.tokenUpdated {
				secretNotExist := false
				secret := corev1.Secret{}
				if err = r.client.Get(context.TODO(), types.NamespacedName{Name: existingOrg.GetName() + "-msg-crypto",
					Namespace: existingOrg.GetName()}, &secret); err != nil {
					log.Error(err, fmt.Sprintf("get secret %s error", existingOrg.GetName()+"-msg-crypto"))
					if k8serrors.IsNotFound(err) {
						secretNotExist = true
					}
				}
				update.tokenUpdated = secretNotExist
			}
		}

		added, removed := current.DifferClients(existingOrg.Spec.Clients, organization.Spec.Clients)
		if len(added) != 0 || len(removed) != 0 {
			update.clientsUpdated = true
			update.clientsRemoved = strings.Join(removed, ",")
		}

		r.PushUpdate(organization.GetName(), update)
		return true
	}

	update.specUpdated = true
	update.adminUpdated = true
	update.clientsUpdated = true
	if organization.Spec.AdminToken != "" {
		update.tokenUpdated = true
	}

	r.PushUpdate(organization.GetName(), update)
	return true
}

func (r *ReconcileOrganization) UpdateFunc(e event.UpdateEvent) bool {
	oldOrg := e.ObjectOld.(*current.Organization)
	newOrg := e.ObjectNew.(*current.Organization)
	log.Info(fmt.Sprintf("Update event detected for organization '%s'", oldOrg.GetName()))

	return r.PredictOrganizationUpdate(oldOrg, newOrg)
}

func (r *ReconcileOrganization) PredictOrganizationUpdate(oldOrg *current.Organization, newOrg *current.Organization) bool {
	update := Update{}

	if reflect.DeepEqual(oldOrg.Spec, newOrg.Spec) {
		return false
	}

	if oldOrg.Spec.Admin != newOrg.Spec.Admin {
		update.adminUpdated = true
		update.adminTransfered = oldOrg.Spec.Admin
	}

	// https://github.com/bestchains/fabric-operator/issues/14#issuecomment-1371948917
	if oldOrg.Spec.AdminToken != newOrg.Spec.AdminToken && newOrg.Spec.AdminToken != "" {
		update.tokenUpdated = update.adminUpdated
		if !update.tokenUpdated {
			secretNotExist := false
			secret := corev1.Secret{}
			if err := r.client.Get(context.TODO(), types.NamespacedName{Name: oldOrg.GetName() + "-msg-crypto",
				Namespace: oldOrg.GetName()}, &secret); err != nil {
				log.Error(err, fmt.Sprintf("get secret %s error", oldOrg.GetName()+"-msg-crypto"))
				if k8serrors.IsNotFound(err) {
					secretNotExist = true
				}
			}
			update.tokenUpdated = secretNotExist
		}
	}

	added, removed := current.DifferClients(oldOrg.Spec.Clients, newOrg.Spec.Clients)
	if len(added) != 0 || len(removed) != 0 {
		update.clientsUpdated = true
		update.clientsRemoved = strings.Join(removed, ",")
	}

	r.PushUpdate(oldOrg.Name, update)

	log.Info(fmt.Sprintf("Spec update triggering reconcile on Organization custom resource %s: update [ %+v ]", oldOrg.Name, update.GetUpdateStackWithTrues()))

	return true
}

func (r *ReconcileOrganization) DeleteFunc(e event.DeleteEvent) bool {
	var err error
	organization := e.Object.(*current.Organization)
	log.Info(fmt.Sprintf("Delete event detected for organization '%s'", organization.GetName()))

	// reconcile users uppon organization delete
	if r.Config.OrganizationInitConfig.IAMEnabled {
		selector, _ := user.OrganizationSelector(organization.Name)
		userList, err := user.ListUsers(r.client, selector)
		if err != nil {
			log.Error(err, fmt.Sprintf("failed to get iam users with organization annotation key %s", organization.GetName()))
		} else {
			for i, iamuser := range userList.Items {
				idType := iamuser.Labels[user.OrganizationLabel.String(organization.Name)]
				_, err = user.ReconcileRemove(&userList.Items[i], organization.Name, "", user.IDType(idType))
				if err != nil {
					log.Error(err, fmt.Sprintf("failed to delete annotation %s for %s", organization.GetName(), iamuser.GetName()))
				}
			}
		}
		err = user.PatchUsers(r.client, userList.Items...)
		if err != nil {
			log.Error(err, fmt.Sprintf("failed to patch users for %s", organization.GetName()))
		}
	}

	// delete namespace
	ns := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: organization.GetUserNamespace(),
		},
	}
	err = r.client.Delete(context.TODO(), ns, &client.DeleteOptions{})
	if err != nil {
		log.Error(err, fmt.Sprintf("failed to delete namespace for organiation %s", organization.GetName()))
	}
	return false
}

// Federation related predict funcs
func (r *ReconcileOrganization) FederationCreateFunc(e event.CreateEvent) bool {
	var err error

	federation := e.Object.(*current.Federation)
	log.Info(fmt.Sprintf("Create event detected for federation '%s'", federation.GetName()))

	for _, m := range federation.Spec.Members {
		err = r.AddFed(m, federation)
		if err != nil {
			log.Error(err, fmt.Sprintf("Member %s in Federation %s", m.GetNamespacedName(), federation.GetName()))
		}
	}

	return false
}

func (r *ReconcileOrganization) FederationUpdateFunc(e event.UpdateEvent) bool {
	var err error

	oldFed := e.ObjectOld.(*current.Federation)
	newFed := e.ObjectNew.(*current.Federation)
	log.Info(fmt.Sprintf("Update event detected for fedeartion '%s'", oldFed.GetName()))

	oldMembers := oldFed.Spec.Members
	newMembers := newFed.Spec.Members

	added, removed := current.DifferMembers(oldMembers, newMembers)

	for _, am := range added {
		err = r.AddFed(am, newFed)
		if err != nil {
			log.Error(err, fmt.Sprintf("Member %s in Federation %s", am.GetNamespacedName(), newFed.GetName()))
		}
	}

	for _, rm := range removed {
		err = r.DeleteFed(rm, newFed)
		if err != nil {
			log.Error(err, fmt.Sprintf("Member %s in Federation %s", rm.GetNamespacedName(), newFed.GetName()))
		}
	}

	return false
}

func (r *ReconcileOrganization) FederationDeleteFunc(e event.DeleteEvent) bool {
	var err error

	federation := e.Object.(*current.Federation)
	log.Info(fmt.Sprintf("Delete event detected for federation '%s'", federation.GetName()))

	for _, m := range federation.Spec.Members {
		err = r.DeleteFed(m, federation)
		if err != nil {
			log.Error(err, fmt.Sprintf("Member %s in Federation %s", m.GetNamespacedName(), federation.GetName()))
		}
	}
	return false
}

func (r *ReconcileOrganization) AddFed(m current.Member, federation *current.Federation) error {
	var err error
	organization := &current.Organization{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      m.Name,
		Namespace: m.Namespace,
	}, organization)
	if err != nil {
		return err
	}

	conflict := organization.Status.AddFederation(current.NamespacedName{
		Name:      federation.Name,
		Namespace: federation.Namespace,
	})
	// conflict detected,do not need to PatchStatus
	if conflict {
		return errors.Errorf("federation %s already exist in organization %s", federation.GetName(), m.GetNamespacedName())
	}

	err = r.client.PatchStatus(context.TODO(), organization, nil, k8sclient.PatchOption{
		Resilient: &k8sclient.ResilientPatch{
			Retry:    2,
			Into:     &current.Organization{},
			Strategy: client.MergeFrom,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

// CA related predict funcs
func (r *ReconcileOrganization) CAUpdateFunc(e event.UpdateEvent) bool {
	var err error

	oldCA := e.ObjectOld.(*current.IBPCA)
	newCA := e.ObjectNew.(*current.IBPCA)
	log.Info(fmt.Sprintf("Update event detected for ibpca '%s'", oldCA.GetName()))

	org := &current.Organization{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name: newCA.GetOrganization().Name,
	}, org)
	if err != nil {
		log.Error(err, fmt.Sprintf("failed to get organization %s`", newCA.GetOrganization().Name))
		return false
	}
	// sync to CAStatus
	err = r.SetStatus(org, &newCA.Status.CRStatus)
	if err != nil {
		log.Error(err, fmt.Sprintf("set organization %s to %s", org.GetName(), newCA.Status.Type))
	}

	return false
}

func (r *ReconcileOrganization) UpdateStatus(organization current.NamespacedName, newStatus current.CRStatus) error {
	var err error
	org := &current.Organization{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: organization.Name}, org)
	if err != nil {
		return err
	}

	status := org.Status.CRStatus
	status.Type = newStatus.Type
	status.Status = current.True
	status.Reason = newStatus.Reason
	status.LastHeartbeatTime = metav1.Now()

	org.Status = current.OrganizationStatus{
		CRStatus: status,
	}

	err = r.client.PatchStatus(context.TODO(), org, nil, k8sclient.PatchOption{
		Resilient: &k8sclient.ResilientPatch{
			Retry:    2,
			Into:     &current.Organization{},
			Strategy: client.MergeFrom,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *ReconcileOrganization) DeleteFed(m current.Member, federation *current.Federation) error {
	var err error

	organization := &current.Organization{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      m.Name,
		Namespace: m.Namespace,
	}, organization)
	if err != nil {
		return err
	}

	exist := organization.Status.DeleteFederation(current.NamespacedName{
		Name:      federation.Name,
		Namespace: federation.Namespace,
	})

	// federation do not exist in this organization ,do not need to PatchStatus
	if !exist {
		return errors.Errorf("federation %s not exist in organization %s", federation.GetName(), m.GetNamespacedName())
	}

	err = r.client.PatchStatus(context.TODO(), organization, nil, k8sclient.PatchOption{
		Resilient: &k8sclient.ResilientPatch{
			Retry:    2,
			Into:     &current.Organization{},
			Strategy: client.MergeFrom,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *ReconcileOrganization) GetIAMUser(username string) (*iam.User, error) {
	var err error
	iamuser := &iam.User{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: username}, iamuser)
	if err != nil {
		return nil, err
	}
	return iamuser, nil
}
