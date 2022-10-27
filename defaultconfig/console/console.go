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

package console

import (
	"github.com/IBM-Blockchain/fabric-operator/pkg/apis/deployer"
	"github.com/IBM-Blockchain/fabric-operator/pkg/util"
)

func GetImages() *deployer.ConsoleImages {
	return &deployer.ConsoleImages{
		ConsoleImage:       util.GetRegistyServer() + "fabric-console",
		ConsoleTag:         "latest",
		ConsoleInitImage:   util.GetRegistyServer() + "ubi-minimal",
		ConsoleInitTag:     "latest",
		ConfigtxlatorImage: util.GetRegistyServer() + "fabric-tools",
		ConfigtxlatorTag:   "2.2.5",
		DeployerImage:      util.GetRegistyServer() + "fabric-deployer",
		DeployerTag:        "latest",
		CouchDBImage:       util.GetRegistyServer() + "couchdb",
		CouchDBTag:         "3.2.2",
	}
}
