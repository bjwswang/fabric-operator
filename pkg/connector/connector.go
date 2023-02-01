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

package connector

import (
	fabsdkconfig "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
)

// Connector is used to connect with blockchain network
type Connector struct {
	// profile which defines all required connection info
	profile []byte

	sdk *fabsdk.FabricSDK
}

// ProfileFunc defines how to get the profile(marshalled)
type ProfileFunc func() ([]byte, error)

func NewConnector(profile ProfileFunc) (*Connector, error) {
	binaryData, err := profile()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get connector profile")
	}
	c := &Connector{
		profile: binaryData,
	}
	err = c.init()
	if err != nil {
		return nil, errors.Wrap(err, "failed to init connector")

	}
	return c, nil
}

func (c *Connector) init() error {
	// only yaml supported for now
	sdk, err := fabsdk.New(fabsdkconfig.FromRaw(c.profile, "yaml"))
	if err != nil {
		return err
	}
	c.sdk = sdk

	return nil
}

func (c *Connector) Close() {
	if c.sdk == nil {
		return
	}
	c.sdk.Close()
	c.sdk = nil
}

func (c *Connector) SDK() *fabsdk.FabricSDK {
	return c.sdk
}
