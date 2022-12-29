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
	"strings"

	"github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
)

var GroupVersion = v1beta1.GroupVersion

type Resource string

const (
	// Cluster Scope
	Organization Resource = "Organization"
	Federation   Resource = "Federation"
	Proposal     Resource = "Proposal"
	Network      Resource = "Network"
	Channel      Resource = "Channel"

	// Namespaced Scope
	Votes       Resource = "Vote"
	IBPCAs      Resource = "IBPCA"
	IBPOrderers Resource = "IBPOrderer"
	IBPPeers    Resource = "IBPPeer"
)

func (resource Resource) String() string {
	return strings.ToLower(string(resource)) + "s"
}
