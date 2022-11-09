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

package integration

import "github.com/IBM-Blockchain/fabric-operator/pkg/util"

var (
	FabricCAVersion    = "1.5.3"
	FabricVersion      = "2.2.5"
	FabricVersion24    = "2.4.7"
	InitImage          = util.GetRegistyServer() + "ubi-minimal"
	InitTag            = "latest"
	CaImage            = util.GetRegistyServer() + "fabric-ca"
	CaTag              = FabricCAVersion
	PeerImage          = util.GetRegistyServer() + "fabric-peer"
	PeerTag            = FabricVersion24
	OrdererImage       = util.GetRegistyServer() + "fabric-orderer"
	OrdererTag         = FabricVersion24
	Orderer14Tag       = "1.4.12"
	Orderer24Tag       = FabricVersion24
	ConfigtxlatorImage = util.GetRegistyServer() + "fabric-tools"
	ConfigtxlatorTag   = FabricVersion24
	CouchdbImage       = util.GetRegistyServer() + "couchdb"
	CouchdbTag         = "3.2.2"
	GrpcwebImage       = util.GetRegistyServer() + "grpc-web"
	GrpcwebTag         = "latest"
	ConsoleImage       = util.GetRegistyServer() + "fabric-console"
	ConsoleTag         = "latest"
	DeployerImage      = util.GetRegistyServer() + "fabric-deployer"
	DeployerTag        = "latest-amd64"
)
