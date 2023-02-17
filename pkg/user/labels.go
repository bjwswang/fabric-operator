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
	"errors"
	"fmt"
	"strings"

	bcrbac "github.com/IBM-Blockchain/fabric-operator/pkg/rbac"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	BlockchainLabelPrefix = "bestchains"
)

type UserLabel string

const (
	OrganizationLabel UserLabel = "organizaiton"
)

func (l UserLabel) String(suffix ...string) string {
	return strings.Join(append([]string{BlockchainLabelPrefix, string(OrganizationLabel)}, suffix...), ".")
}

// OrganizationSelector constructs selector with `OrganizationLabel's val is in Roles`
func OrganizationSelector(organization string) (labels.Selector, error) {
	key := OrganizationLabel.String(organization)
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(key, selection.In, bcrbac.Roles())
	if err != nil {
		return nil, err
	}
	selector = selector.Add(*requirement)
	return selector, nil
}

// UserSelector constructs selector with `t7d.io.username:xxx`
func UserSelector(preferUsername string) (labels.Selector, error) {
	if preferUsername == "" {
		return nil, errors.New("empty preferUsername")
	}
	return labels.Parse(fmt.Sprintf("t7d.io.username=%s", preferUsername))
}
