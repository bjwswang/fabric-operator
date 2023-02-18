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

package chaincode

import "testing"

func TestPodName(t *testing.T) {
	msgID, peerID, chaincodeID := "org2", "org2peer1", "chaincode1:79ecefe2e6ce879134996982dbda6d9427318cb2edb006af66657e28c5043d23"
	expectPodName := "cc-org2-org2peer1chaincode1-79ecefe2e6ce879134996982dbda6d94273"
	if name := PodName(msgID, peerID, chaincodeID); name != expectPodName {
		t.Fatalf("expect %s get %s", expectPodName, name)
	}
}
