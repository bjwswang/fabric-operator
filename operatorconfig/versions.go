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

package operatorconfig

import (
	"github.com/IBM-Blockchain/fabric-operator/pkg/apis/deployer"
	"github.com/IBM-Blockchain/fabric-operator/pkg/util"
)

const (
	LatestTag       = "latest"
	FabricCAVersion = "1.5.3"
	FabricVersion   = "2.4.7"
)

func getDefaultVersions() *deployer.Versions {
	InitImage := util.GetRegistyServer() + "ubi-minimal"
	return &deployer.Versions{
		CA: map[string]deployer.VersionCA{
			"1.5.3-1": {
				Default: true,
				Version: "1.5.3-1",
				Image: deployer.CAImages{
					CAInitImage: InitImage,
					CAInitTag:   LatestTag,
					CAImage:     util.GetRegistyServer() + "fabric-ca",
					CATag:       FabricCAVersion,
				},
			},
		},
		Peer: map[string]deployer.VersionPeer{
			"2.4.7-1": {
				Default: true,
				Version: "2.4.7-1",
				Image: deployer.PeerImages{
					PeerInitImage: InitImage,
					PeerInitTag:   LatestTag,
					PeerImage:     util.GetRegistyServer() + "fabric-peer",
					PeerTag:       FabricVersion,
					CouchDBImage:  "couchdb",
					CouchDBTag:    "3.2.2",
					GRPCWebImage:  util.GetRegistyServer() + "grpc-web",
					GRPCWebTag:    LatestTag,
				},
			},
		},
		Orderer: map[string]deployer.VersionOrderer{
			"2.4.7-1": {
				Default: true,
				Version: "2.4.7-1",
				Image: deployer.OrdererImages{
					OrdererInitImage: InitImage,
					OrdererInitTag:   LatestTag,
					OrdererImage:     util.GetRegistyServer() + "fabric-orderer",
					OrdererTag:       FabricVersion,
					GRPCWebImage:     util.GetRegistyServer() + "grpc-web",
					GRPCWebTag:       LatestTag,
				},
			},
		},
	}
}
