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

package fake

import (
	"context"

	v1beta1 "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeEndorsePolicies implements EndorsePolicyInterface
type FakeEndorsePolicies struct {
	Fake *FakeIbp
}

var endorsepoliciesResource = schema.GroupVersionResource{Group: "ibp.com", Version: "", Resource: "endorsepolicies"}

var endorsepoliciesKind = schema.GroupVersionKind{Group: "ibp.com", Version: "", Kind: "EndorsePolicy"}

// Get takes name of the endorsePolicy, and returns the corresponding endorsePolicy object, and an error if there is any.
func (c *FakeEndorsePolicies) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.EndorsePolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(endorsepoliciesResource, name), &v1beta1.EndorsePolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.EndorsePolicy), err
}

// List takes label and field selectors, and returns the list of EndorsePolicies that match those selectors.
func (c *FakeEndorsePolicies) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.EndorsePolicyList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(endorsepoliciesResource, endorsepoliciesKind, opts), &v1beta1.EndorsePolicyList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.EndorsePolicyList{ListMeta: obj.(*v1beta1.EndorsePolicyList).ListMeta}
	for _, item := range obj.(*v1beta1.EndorsePolicyList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested endorsePolicies.
func (c *FakeEndorsePolicies) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(endorsepoliciesResource, opts))
}

// Create takes the representation of a endorsePolicy and creates it.  Returns the server's representation of the endorsePolicy, and an error, if there is any.
func (c *FakeEndorsePolicies) Create(ctx context.Context, endorsePolicy *v1beta1.EndorsePolicy, opts v1.CreateOptions) (result *v1beta1.EndorsePolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(endorsepoliciesResource, endorsePolicy), &v1beta1.EndorsePolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.EndorsePolicy), err
}

// Update takes the representation of a endorsePolicy and updates it. Returns the server's representation of the endorsePolicy, and an error, if there is any.
func (c *FakeEndorsePolicies) Update(ctx context.Context, endorsePolicy *v1beta1.EndorsePolicy, opts v1.UpdateOptions) (result *v1beta1.EndorsePolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(endorsepoliciesResource, endorsePolicy), &v1beta1.EndorsePolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.EndorsePolicy), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeEndorsePolicies) UpdateStatus(ctx context.Context, endorsePolicy *v1beta1.EndorsePolicy, opts v1.UpdateOptions) (*v1beta1.EndorsePolicy, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(endorsepoliciesResource, "status", endorsePolicy), &v1beta1.EndorsePolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.EndorsePolicy), err
}

// Delete takes name of the endorsePolicy and deletes it. Returns an error if one occurs.
func (c *FakeEndorsePolicies) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(endorsepoliciesResource, name), &v1beta1.EndorsePolicy{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeEndorsePolicies) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(endorsepoliciesResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta1.EndorsePolicyList{})
	return err
}

// Patch applies the patch and returns the patched endorsePolicy.
func (c *FakeEndorsePolicies) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.EndorsePolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(endorsepoliciesResource, name, pt, data, subresources...), &v1beta1.EndorsePolicy{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.EndorsePolicy), err
}
