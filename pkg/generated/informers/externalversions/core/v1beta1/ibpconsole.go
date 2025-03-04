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

// IBPConsoleInformer provides access to a shared informer and lister for
// IBPConsoles.
type IBPConsoleInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1beta1.IBPConsoleLister
}

type iBPConsoleInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewIBPConsoleInformer constructs a new informer for IBPConsole type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewIBPConsoleInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredIBPConsoleInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredIBPConsoleInformer constructs a new informer for IBPConsole type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredIBPConsoleInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.Ibp().IBPConsoles(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.Ibp().IBPConsoles(namespace).Watch(context.TODO(), options)
			},
		},
		&apiv1beta1.IBPConsole{},
		resyncPeriod,
		indexers,
	)
}

func (f *iBPConsoleInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredIBPConsoleInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *iBPConsoleInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&apiv1beta1.IBPConsole{}, f.defaultInformer)
}

func (f *iBPConsoleInformer) Lister() v1beta1.IBPConsoleLister {
	return v1beta1.NewIBPConsoleLister(f.Informer().GetIndexer())
}
