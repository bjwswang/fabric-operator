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

// Code generated by informer-gen. DO NOT EDIT.

package v1beta1

import (
	"context"
	time "time"

	apiv1beta1 "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	versioned "github.com/IBM-Blockchain/fabric-operator/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/IBM-Blockchain/fabric-operator/pkg/generated/informers/externalversions/internalinterfaces"
	v1beta1 "github.com/IBM-Blockchain/fabric-operator/pkg/generated/listers/core/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// OrganizationInformer provides access to a shared informer and lister for
// Organizations.
type OrganizationInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1beta1.OrganizationLister
}

type organizationInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewOrganizationInformer constructs a new informer for Organization type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewOrganizationInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredOrganizationInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredOrganizationInformer constructs a new informer for Organization type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredOrganizationInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.Ibp().Organizations().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.Ibp().Organizations().Watch(context.TODO(), options)
			},
		},
		&apiv1beta1.Organization{},
		resyncPeriod,
		indexers,
	)
}

func (f *organizationInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredOrganizationInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *organizationInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&apiv1beta1.Organization{}, f.defaultInformer)
}

func (f *organizationInformer) Lister() v1beta1.OrganizationLister {
	return v1beta1.NewOrganizationLister(f.Informer().GetIndexer())
}
