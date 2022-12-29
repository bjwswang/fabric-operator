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

package rbac

import (
	"context"
	"errors"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	ErrBadSynchronizer = errors.New("bad synchronizer")
)

// Synchronizer sync RBAC based on different ResourceAction uppon k8s object
type Synchronizer func(controllerclient.Client, v1.Object, ResourceAction) error

// defaultSynchronizers stores default RBAC Synchronizer on part of resouces
var defaultSynchronizers = make(map[Resource]Synchronizer)

func init() {
	defaultSynchronizers[Federation] = SyncFederation
	defaultSynchronizers[Proposal] = SyncProposal
	defaultSynchronizers[Network] = SyncNetwork
}

func EmptySynchronizer(c controllerclient.Client, o v1.Object, ra ResourceAction) error {
	return nil
}

// SyncFederation triggers synchronization based on Federation's action(create/update/delete)
func SyncFederation(c controllerclient.Client, o v1.Object, ra ResourceAction) error {
	var err error

	federation, ok := o.(*current.Federation)
	if !ok {
		return ErrBadSynchronizer
	}

	// PolicyRule which should be appended/removed from role's rules
	targetRule := PolicyRule(Federation, []v1.Object{o}, []Verb{Get})

	// Make sure each organization sync on above rule
	for _, member := range federation.GetMembers() {
		organization := &current.Organization{}
		err = c.Get(context.TODO(), types.NamespacedName{Name: member.Name}, organization)
		if err != nil {
			return err
		}
		key := GetClusterRole(organization.GetNamespaced(), Admin)
		err = SyncClusterRole(c, key, targetRule, ra)
		if err != nil {
			return err
		}
	}

	return nil
}

// SyncFederation triggers synchronization uppon Proposal's action(create/update/delete)
func SyncProposal(c controllerclient.Client, o v1.Object, ra ResourceAction) error {
	var err error

	proposal, ok := o.(*current.Proposal)
	if !ok {
		return ErrBadSynchronizer
	}
	// PolicyRule which should be appended/removed from role's rules
	targetRule := PolicyRule(Proposal, []v1.Object{o}, []Verb{Get})

	// candidates stands for the organizations which are expected within this federation(cluster scope)
	candidates, err := proposal.GetCandidateOrganizations(context.TODO(), c)
	if err != nil {
		return err
	}

	for _, candidate := range candidates {
		organization := &current.Organization{}
		err = c.Get(context.TODO(), types.NamespacedName{Name: candidate.Name}, organization)
		if err != nil {
			return err
		}
		key := GetClusterRole(organization.GetNamespaced(), Admin)
		err = SyncClusterRole(c, key, targetRule, ra)
		if err != nil {
			return err
		}
	}

	return nil
}

// SyncNetwork triggers synchronization uppon Network's action(create/update/delete)
func SyncNetwork(c controllerclient.Client, o v1.Object, ra ResourceAction) error {
	var err error

	network, ok := o.(*current.Network)
	if !ok {
		return ErrBadSynchronizer
	}
	// PolicyRule which should be appended/removed from role's rules
	targetRule := PolicyRule(Network, []v1.Object{o}, []Verb{Get})
	// Make sure each organization sync on above rule
	for _, member := range network.GetMembers() {
		organization := &current.Organization{}
		err = c.Get(context.TODO(), types.NamespacedName{Name: member.Name}, organization)
		if err != nil {
			return err
		}
		key := GetClusterRole(organization.GetNamespaced(), Admin)
		err = SyncClusterRole(c, key, targetRule, ra)
		if err != nil {
			return err
		}
	}

	return nil
}

// SyncClusterRole defines common reconcile logic on clusterroles uppon cluster-scoped resources' action(create/update/delete)
func SyncClusterRole(c controllerclient.Client, key types.NamespacedName, rule rbacv1.PolicyRule, ra ResourceAction) error {
	clusterRole := &rbacv1.ClusterRole{}
	err := c.Get(context.TODO(), key, clusterRole)
	if err != nil {
		return err
	}
	switch ra {
	case ResourceCreate, ResourceUpdate:
		_, ok := CheckPolicyRule(clusterRole.Rules, rule)
		if !ok {
			// create if not exist
			clusterRole.Rules = append(clusterRole.Rules, rule)
		}
	case ResourceDelete:
		pos, ok := CheckPolicyRule(clusterRole.Rules, rule)
		if ok {
			// delete if exist
			clusterRole.Rules = append(clusterRole.Rules[0:pos], clusterRole.Rules[pos+1:]...)
		}
	}

	err = c.Update(context.TODO(), clusterRole)
	if err != nil {
		return err
	}

	return nil
}
