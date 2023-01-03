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
	"encoding/json"
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BlockchainAnnotationKey is key in User's annotations to store blockchains related content
const BlockchainAnnotationKey = "bestchains"

var (
	ErrNilAnnotationList  = errors.New("nil annotation list")
	ErrNilAnnotation      = errors.New("nil blockchain annotation")
	ErrAnnotationNotExist = errors.New("annotation not exists")
	ErrIDNotExist         = errors.New("id not exists")
)

// BlockchainAnnotationList stores blockchain related content which enbale fabric-ca with IAM
type BlockchainAnnotationList struct {
	// List stores User's BlockchainAnnotation in different organizations
	List                    map[string]BlockchainAnnotation `json:"list,omitempty"`
	CreationTimestamp       metav1.Time                     `json:"creationTimestamp,omitempty"`
	LastAppliedTimestamp    metav1.Time                     `json:"lastAppliedTimestamp,omitempty"`
	LastDeletetionTimestamp metav1.Time                     `json:"lastDeletetionTimestamp,omitempty"`
}

func NewBlockchainAnnotationList() *BlockchainAnnotationList {
	return &BlockchainAnnotationList{
		List:                 make(map[string]BlockchainAnnotation),
		CreationTimestamp:    metav1.Now(),
		LastAppliedTimestamp: metav1.Now(),
	}
}

func (annotations *BlockchainAnnotationList) Marshal() ([]byte, error) {
	if annotations == nil {
		return nil, ErrNilAnnotationList
	}
	return json.Marshal(annotations)
}

func (annotations *BlockchainAnnotationList) Unmarshal(raw []byte) error {
	if annotations == nil {
		return ErrNilAnnotationList
	}
	// return nil when raw is empty
	// - raw is empty when no blockchain annotation appened yet
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, annotations)
}

func (annotations *BlockchainAnnotationList) GetAnnotation(k string) (BlockchainAnnotation, error) {
	if annotations == nil {
		return BlockchainAnnotation{}, ErrNilAnnotationList
	}
	if annotations.List == nil {
		annotations.List = make(map[string]BlockchainAnnotation)
		return BlockchainAnnotation{}, ErrAnnotationNotExist
	}
	annotation, ok := annotations.List[k]
	if !ok {
		return BlockchainAnnotation{}, ErrAnnotationNotExist
	}
	return annotation, nil
}

func (annotations *BlockchainAnnotationList) SetAnnotation(k string, annotation BlockchainAnnotation) error {
	if annotations == nil {
		return ErrNilAnnotationList
	}
	if annotations.List == nil {
		annotations.List = make(map[string]BlockchainAnnotation)
	}
	annotations.List[k] = annotation
	annotations.LastAppliedTimestamp = metav1.Now()
	return nil
}

func (annotations *BlockchainAnnotationList) DeleteAnnotation(k string) error {
	if annotations == nil {
		return ErrNilAnnotationList
	}
	if annotations.List == nil {
		annotations.List = make(map[string]BlockchainAnnotation)
		return ErrAnnotationNotExist
	}
	_, ok := annotations.List[k]
	if !ok {
		return ErrAnnotationNotExist
	}
	delete(annotations.List, k)
	annotations.LastDeletetionTimestamp = metav1.Now()
	return nil
}

// BlockchainAnnotation defines blockchain-related fields
type BlockchainAnnotation struct {
	// Organization defines which organization this annotation is for
	Organization string `json:"organization,omitempty"`
	// IDs stores all Fabric-CA identities under this User's government
	IDs                   map[string]ID `json:"ids,omitempty"`
	CreationTimestamp     metav1.Time   `json:"creationTimestamp,omitempty"`
	LastAppliedTimestamp  metav1.Time   `json:"lastAppliedTimestamp,omitempty"`
	LastDeletionTimestamp metav1.Time   `json:"lastDeletionTimestamp,omitempty"`
}

func NewBlockchainAnnotation(organization string, ids ...ID) *BlockchainAnnotation {
	annotation := &BlockchainAnnotation{
		Organization:         organization,
		IDs:                  make(map[string]ID),
		CreationTimestamp:    metav1.Now(),
		LastAppliedTimestamp: metav1.Now(),
	}
	for _, id := range ids {
		_ = annotation.SetID(id)
	}
	return annotation
}

func (annotation *BlockchainAnnotation) GetID(id string) (ID, error) {
	if annotation == nil {
		return ID{}, ErrNilAnnotation
	}
	if annotation.IDs == nil {
		annotation.IDs = make(map[string]ID)
		return ID{}, ErrIDNotExist
	}
	storedID, ok := annotation.IDs[id]
	if !ok {
		return ID{}, ErrIDNotExist
	}
	return storedID, nil
}

func (annotation *BlockchainAnnotation) SetID(id ID) error {
	if annotation == nil {
		return ErrNilAnnotation
	}
	if annotation.IDs == nil {
		annotation.IDs = make(map[string]ID)
	}
	annotation.IDs[id.Name] = id
	annotation.LastAppliedTimestamp = metav1.Now()

	return nil
}

func (annotation *BlockchainAnnotation) RemoveID(id string) error {
	if annotation == nil {
		return ErrNilAnnotation
	}
	if annotation.IDs == nil {
		return ErrIDNotExist
	}
	_, exist := annotation.IDs[id]
	if !exist {
		return ErrIDNotExist
	}

	delete(annotation.IDs, id)
	annotation.LastDeletionTimestamp = metav1.Now()

	return nil
}

type IDType string

const (
	ADMIN   IDType = "admin"
	CLIENT  IDType = "client"
	PEER    IDType = "peer"
	ORDERER IDType = "orderer"
)

func (idType IDType) String() string {
	return string(idType)
}

// ID stands for a Fabric-CA identity
type ID struct {
	Name                 string            `json:"name"`
	Type                 IDType            `json:"type"`
	Attributes           map[string]string `json:"attributes,omitempty"`
	CreationTimestamp    metav1.Time       `json:"creationTimestamp,omitempty"`
	LastAppliedTimestamp metav1.Time       `json:"lastAppliedTimestamp,omitempty"`
}

func BuildAdminID(id string) ID {
	return ID{
		Name: id,
		Type: ADMIN,
		Attributes: map[string]string{
			"hf.EnrollmentID":            id,
			"hf.Type":                    ADMIN.String(),
			"hf.Affiliation":             "",
			"hf.Registrar.Roles":         "*",
			"hf.RegistrarDelegateRoles":  "*",
			"hf.Revoker":                 "*",
			"hf.IntermediateCA":          "true",
			"hf.GenCRL":                  "true",
			"hf.hf.Registrar.Attributes": "*",
		},
		CreationTimestamp:    metav1.Now(),
		LastAppliedTimestamp: metav1.Now(),
	}
}

func BuildClientID(id string) ID {
	return ID{
		Name: id,
		Type: CLIENT,
		Attributes: map[string]string{
			"hf.EnrollmentID": id,
			"hf.Type":         CLIENT.String(),
			"hf.Affiliation":  "",
		},
		CreationTimestamp:    metav1.Now(),
		LastAppliedTimestamp: metav1.Now(),
	}
}

func BuildPeerID(id string) ID {
	return ID{
		Name: id,
		Type: PEER,
		Attributes: map[string]string{
			"hf.EnrollmentID": id,
			"hf.Type":         PEER.String(),
			"hf.Affiliation":  "",
		},
		CreationTimestamp:    metav1.Now(),
		LastAppliedTimestamp: metav1.Now(),
	}
}

func BuildOrdererID(id string) ID {
	return ID{
		Name: id,
		Type: ORDERER,
		Attributes: map[string]string{
			"hf.EnrollmentID": id,
			"hf.Type":         ORDERER.String(),
			"hf.Affiliation":  "",
		},
		CreationTimestamp:    metav1.Now(),
		LastAppliedTimestamp: metav1.Now(),
	}
}
