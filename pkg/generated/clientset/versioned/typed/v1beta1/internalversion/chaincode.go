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

// Code generated by client-gen. DO NOT EDIT.

package internalversion

import (
	"context"
	"time"

	v1beta1 "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	scheme "github.com/IBM-Blockchain/fabric-operator/pkg/generated/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ChaincodesGetter has a method to return a ChaincodeInterface.
// A group's client should implement this interface.
type ChaincodesGetter interface {
	Chaincodes() ChaincodeInterface
}

// ChaincodeInterface has methods to work with Chaincode resources.
type ChaincodeInterface interface {
	Create(ctx context.Context, chaincode *v1beta1.Chaincode, opts v1.CreateOptions) (*v1beta1.Chaincode, error)
	Update(ctx context.Context, chaincode *v1beta1.Chaincode, opts v1.UpdateOptions) (*v1beta1.Chaincode, error)
	UpdateStatus(ctx context.Context, chaincode *v1beta1.Chaincode, opts v1.UpdateOptions) (*v1beta1.Chaincode, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1beta1.Chaincode, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1beta1.ChaincodeList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.Chaincode, err error)
	ChaincodeExpansion
}

// chaincodes implements ChaincodeInterface
type chaincodes struct {
	client rest.Interface
}

// newChaincodes returns a Chaincodes
func newChaincodes(c *IbpClient) *chaincodes {
	return &chaincodes{
		client: c.RESTClient(),
	}
}

// Get takes name of the chaincode, and returns the corresponding chaincode object, and an error if there is any.
func (c *chaincodes) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.Chaincode, err error) {
	result = &v1beta1.Chaincode{}
	err = c.client.Get().
		Resource("chaincodes").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Chaincodes that match those selectors.
func (c *chaincodes) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.ChaincodeList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1beta1.ChaincodeList{}
	err = c.client.Get().
		Resource("chaincodes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested chaincodes.
func (c *chaincodes) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("chaincodes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a chaincode and creates it.  Returns the server's representation of the chaincode, and an error, if there is any.
func (c *chaincodes) Create(ctx context.Context, chaincode *v1beta1.Chaincode, opts v1.CreateOptions) (result *v1beta1.Chaincode, err error) {
	result = &v1beta1.Chaincode{}
	err = c.client.Post().
		Resource("chaincodes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(chaincode).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a chaincode and updates it. Returns the server's representation of the chaincode, and an error, if there is any.
func (c *chaincodes) Update(ctx context.Context, chaincode *v1beta1.Chaincode, opts v1.UpdateOptions) (result *v1beta1.Chaincode, err error) {
	result = &v1beta1.Chaincode{}
	err = c.client.Put().
		Resource("chaincodes").
		Name(chaincode.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(chaincode).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *chaincodes) UpdateStatus(ctx context.Context, chaincode *v1beta1.Chaincode, opts v1.UpdateOptions) (result *v1beta1.Chaincode, err error) {
	result = &v1beta1.Chaincode{}
	err = c.client.Put().
		Resource("chaincodes").
		Name(chaincode.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(chaincode).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the chaincode and deletes it. Returns an error if one occurs.
func (c *chaincodes) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("chaincodes").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *chaincodes) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("chaincodes").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched chaincode.
func (c *chaincodes) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.Chaincode, err error) {
	result = &v1beta1.Chaincode{}
	err = c.client.Patch(pt).
		Resource("chaincodes").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
