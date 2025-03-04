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

package orderer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	k8sclient "github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/manager/resources"
	"github.com/IBM-Blockchain/fabric-operator/pkg/operatorerrors"
	"github.com/go-test/deep"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("orderer_manager")

type Manager struct {
	Client            k8sclient.Client
	Scheme            *runtime.Scheme
	OrdererFile       string
	IgnoreDifferences []string
	Name              string

	LabelsFunc   func(v1.Object) map[string]string
	OverrideFunc func(v1.Object, *current.IBPOrderer, resources.Action) error
}

func (m *Manager) GetName(instance v1.Object) string {
	name := instance.GetName()
	return GetName(name)
}

func (m *Manager) Reconcile(instance v1.Object, update bool) error {
	network := instance.(*current.Network)
	initiator := network.GetInitiatorMember()
	if initiator == "" {
		return fmt.Errorf("network:%s initiator no name setting", network.Name)
	}
	initiatorOrg := &current.Organization{ObjectMeta: v1.ObjectMeta{Name: initiator}}
	initiatorNamespace := initiatorOrg.GetUserNamespace()
	orderer := &current.IBPOrderer{}
	err := m.Client.Get(context.TODO(), types.NamespacedName{Name: network.Name, Namespace: initiatorNamespace}, orderer)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info(fmt.Sprintf("Creating orderer '%s'", network.Name))
			orderer, err = m.GetOrdererBasedOnCRFromFile(instance)
			if err != nil {
				return err
			}

			log.Info(fmt.Sprintf("Setting controller reference instance name: %s, orderer name: %s", instance.GetName(), orderer.GetName()))
			err = m.Client.Create(context.TODO(), orderer, k8sclient.CreateOption{Owner: instance, Scheme: m.Scheme})
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}

	if update {
		log.Info(fmt.Sprintf("Updating orderer node is not allowed programmatically '%s'", network.Name))
		return operatorerrors.New(operatorerrors.InvalidOrdererNodeUpdateRequest, "Updating orderer is not allowed programmatically")
	}

	return nil
}

func (m *Manager) GetOrdererBasedOnCRFromFile(instance v1.Object) (*current.IBPOrderer, error) {
	orderer, err := GetOrdererFromFile(m.OrdererFile)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error reading order configuration file: %s", m.OrdererFile))
		return nil, err
	}

	return m.BasedOnCR(instance, orderer)
}

func (m *Manager) BasedOnCR(instance v1.Object, orderer *current.IBPOrderer) (*current.IBPOrderer, error) {
	if m.OverrideFunc != nil {
		err := m.OverrideFunc(instance, orderer, resources.Create)
		if err != nil {
			return nil, operatorerrors.New(operatorerrors.InvalidOrdererNodeCreateRequest, err.Error())
		}
	}

	orderer.Name = m.GetName(instance)
	orderer.ObjectMeta.Name = m.GetName(instance)

	orderer.Labels = m.LabelsFunc(instance)

	return orderer, nil
}

func (m *Manager) CheckState(instance v1.Object) error {
	if instance == nil {
		return nil // Instance has not been reconciled yet
	}

	name := m.GetName(instance)

	// Get the latest version of the instance
	orderernode := &current.IBPOrderer{}
	err := m.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: instance.GetNamespace()}, orderernode)
	if err != nil {
		return nil
	}

	copy := orderernode.DeepCopy()
	expectedOrderernode, err := m.BasedOnCR(instance, copy)
	if err != nil {
		return err
	}

	deep.MaxDepth = 20
	deep.MaxDiff = 30
	deep.CompareUnexportedFields = true
	deep.LogErrors = true

	diff := deep.Equal(orderernode.Spec, expectedOrderernode.Spec)
	if diff != nil {
		err := m.ignoreDifferences(diff)
		if err != nil {
			return errors.Wrap(err, "orderernode has been edited manually, and does not match what is expected based on the CR")
		}
	}

	return nil
}

func (m *Manager) RestoreState(instance v1.Object) error {
	if instance == nil {
		return nil // Instance has not been reconciled yet
	}

	name := m.GetName(instance)
	orderernode := &current.IBPOrderer{}
	err := m.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: instance.GetNamespace()}, orderernode)
	if err != nil {
		return nil
	}

	orderernode, err = m.BasedOnCR(instance, orderernode)
	if err != nil {
		return err
	}

	err = m.Client.Update(context.TODO(), orderernode)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) Get(instance v1.Object) (client.Object, error) {
	if instance == nil {
		return nil, nil // Instance has not been reconciled yet
	}

	name := m.GetName(instance)
	orderernode := &current.IBPOrderer{}
	err := m.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: instance.GetNamespace()}, orderernode)
	if err != nil {
		return nil, err
	}

	return orderernode, nil
}

func (m *Manager) Exists(instance v1.Object) bool {
	_, err := m.Get(instance)
	return err == nil
}

func (m *Manager) Delete(instance v1.Object) error {
	on, err := m.Get(instance)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	if on == nil {
		return nil
	}

	err = m.Client.Delete(context.TODO(), on)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func (m *Manager) getSelectorLabels(instance v1.Object) map[string]string {
	return map[string]string{
		"app": instance.GetName(),
	}
}

func (m *Manager) ignoreDifferences(diff []string) error {
	diffs := []string{}
	for _, d := range diff {
		found := false
		for _, i := range m.differenceToIgnore() {
			regex := regexp.MustCompile(i)
			found = regex.MatchString(d)
			if found {
				break
			}
		}
		if !found {
			diffs = append(diffs, d)
			return fmt.Errorf("unexpected mismatch: %s", d)
		}
	}
	return nil
}

func (m *Manager) differenceToIgnore() []string {
	d := []string{
		"TypeMeta", "ObjectMeta",
	}
	d = append(d, m.IgnoreDifferences...)
	return d
}

func (m *Manager) SetCustomName(name string) {
	// NO-OP
}

func GetName(instanceName string, suffix ...string) string {
	if len(suffix) != 0 {
		if suffix[0] != "" {
			return fmt.Sprintf("%s-%s", instanceName, suffix[0])
		}
	}
	return instanceName
}

func GetOrdererFromFile(file string) (*current.IBPOrderer, error) {
	jsonBytes, err := ConvertYamlFileToJson(file)
	if err != nil {
		return nil, err
	}

	on := &current.IBPOrderer{}
	err = json.Unmarshal(jsonBytes, &on)
	if err != nil {
		return nil, err
	}

	return on, nil
}

func ConvertYamlFileToJson(file string) ([]byte, error) {
	absfilepath, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadFile(filepath.Clean(absfilepath))
	if err != nil {
		return nil, err
	}

	return yaml.ToJSON(bytes)
}
