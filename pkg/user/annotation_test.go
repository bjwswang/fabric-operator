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

package user

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Annotation", func() {
	var organization = "org1"
	var org1admin = "org1admin"
	var idNotExist = "id_not_exist"
	It("Test build ID", func() {
		id := BuildAdminID("admin")
		Expect(id.Type).To(Equal(ADMIN))
		Expect(id.Attributes["hf.Type"]).To(Equal(string(ADMIN)))

		id = BuildClientID("client")
		Expect(id.Type).To(Equal(CLIENT))
		Expect(id.Attributes["hf.Type"]).To(Equal(string(CLIENT)))

		id = BuildPeerID("peer")
		Expect(id.Type).To(Equal(PEER))
		Expect(id.Attributes["hf.Type"]).To(Equal(string(PEER)))

		id = BuildOrdererID("orderer")
		Expect(id.Type).To(Equal(ORDERER))
		Expect(id.Attributes["hf.Type"]).To(Equal(string(ORDERER)))
	})

	It("Test BlockchainAnnotation", func() {
		annotation := NewBlockchainAnnotation(organization, BuildAdminID(org1admin))
		Expect(annotation.Organization).To(Equal(organization))

		id, err := annotation.GetID(org1admin)
		Expect(err).To(BeNil())
		Expect(id.Type).To(Equal(ADMIN))

		_, err = annotation.GetID(idNotExist)
		Expect(err).To(Equal(ErrIDNotExist))

		Expect(annotation.RemoveID(org1admin)).To(BeNil())
		Expect(annotation.RemoveID(idNotExist)).To(Equal(ErrIDNotExist))

		Expect(annotation.SetID(BuildClientID(idNotExist))).To(BeNil())
		Expect(annotation.RemoveID(idNotExist)).To(BeNil())

		annotation = nil
		_, err = annotation.GetID(org1admin)
		Expect(err).To(Equal(ErrNilAnnotation))
		Expect(annotation.SetID(BuildClientID(idNotExist))).To(Equal(ErrNilAnnotation))
		Expect(annotation.RemoveID(org1admin)).To(Equal(ErrNilAnnotation))

	})

	Context("Test blockchain annotation list", func() {
		var list *BlockchainAnnotationList
		BeforeEach(func() {
			list = NewBlockchainAnnotationList()
		})
		It("marshal and unmarshal", func() {
			marshalled, err := list.Marshal()
			Expect(err).To(BeNil())
			err = list.Unmarshal(marshalled)
			Expect(err).To(BeNil())

			err = list.Unmarshal([]byte{})
			Expect(err).To(BeNil())

			list = nil
			_, err = list.Marshal()
			Expect(err).To(Equal(ErrNilAnnotationList))
			err = list.Unmarshal(marshalled)
			Expect(err).To(Equal(ErrNilAnnotationList))
		})
		It("get/set/delete annotations", func() {
			annotation := NewBlockchainAnnotation(organization, BuildAdminID(org1admin))
			Expect(list.SetAnnotation(organization, *annotation)).To(BeNil())

			retrievedAnnotation, err := list.GetAnnotation(organization)
			Expect(err).To(BeNil())
			Expect(retrievedAnnotation.Organization).To(Equal(annotation.Organization))

			Expect(list.DeleteAnnotation(organization)).To(BeNil())

			_, err = list.GetAnnotation(organization)
			Expect(err).To(Equal(ErrAnnotationNotExist))
		})
	})
})
