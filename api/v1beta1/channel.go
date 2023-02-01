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

package v1beta1

func init() {
	SchemeBuilder.Register(&Channel{}, &ChannelList{})
}

func (channel *Channel) GetConnectionPorfile() string {
	return "chan-" + channel.GetName() + "-connection-profile"
}

func (channel *Channel) GetChannelID() string {
	return channel.GetName()
}

func (channel *Channel) GetMembers() []Member {
	return channel.Spec.Members
}

func (channel *Channel) GetPeerCondition(peer NamespacedName) (int, PeerCondition) {
	for index, p := range channel.Status.PeerConditions {
		if p.String() == peer.String() {
			return index, p
		}
	}
	return -1, PeerCondition{
		NamespacedName: peer,
	}
}

func (channel *Channel) HasType() bool {
	return channel.Status.CRStatus.Type != ""
}

func (channel *Channel) HasNetwork() bool {
	return channel.Spec.Network != ""
}

func (channel *Channel) HashMembers() bool {
	return len(channel.Spec.Members) > 0
}

func DifferChannelPeers(old []NamespacedName, new []NamespacedName) (added []NamespacedName, removed []NamespacedName) {
	// cache in map
	oldMapper := make(map[string]NamespacedName, len(old))
	for _, m := range old {
		oldMapper[m.Name] = m
	}

	// calculate differences
	for _, m := range new {

		// added: in new ,but not in old
		if _, ok := oldMapper[m.Name]; !ok {
			added = append(added, m)
			continue
		}

		// delete the intersection
		delete(oldMapper, m.Name)
	}

	for _, m := range oldMapper {
		removed = append(removed, m)
	}

	return
}
