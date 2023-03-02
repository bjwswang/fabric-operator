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

import (
	"context"
	"fmt"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	config "github.com/IBM-Blockchain/fabric-operator/operatorconfig"
	"github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logf.Log.WithName("base_chaincode")

// The chaincode installation is more complex and will be printed out with the installation steps.
const (
	stepPrefix = "-->"
	// the max length of chaincode.Status.Condition
	maxCondHistory = 10
)

// Override What relevant resources will be modified
type Override interface {
	// TODO: Reserved
	Hello()
}

type Chaincode interface {
	Reconcile(*current.Chaincode) (common.Result, error)

	Package(*current.Chaincode) error
	Install(*current.Chaincode) error
	Approve(*current.Chaincode) error
	Commit(*current.Chaincode) error

	// R check if the chaincode is already running
	R(*current.Chaincode) error
}

type baseChaincode struct {
	client   controllerclient.Client
	override Override

	Scheme *runtime.Scheme
	Config *config.Config
}

func (c *baseChaincode) chaincodeProcess(instance *current.Chaincode) error {
	nextCond := current.NextCond(instance)
	switch nextCond {
	case current.ChaincodeCondPackaged:
		return c.Package(instance)
	case current.ChaincodeCondInstalled:
		return c.Install(instance)
	case current.ChaincodeCondApproved:
		return c.Approve(instance)
	case current.ChaincodeCondCommitted:
		return c.Commit(instance)
	case current.ChaincodeCondRunning:
		return c.R(instance)
	}
	return nil
}

func (c *baseChaincode) PatchStatus(ctx context.Context, instance client.Object) error {
	return c.client.PatchStatus(context.TODO(), instance, nil, controllerclient.PatchOption{
		Resilient: &controllerclient.ResilientPatch{
			Retry:    3,
			Into:     &current.Chaincode{},
			Strategy: client.MergeFrom,
		},
	})
}

func (c *baseChaincode) Reconcile(instance *current.Chaincode) (common.Result, error) {
	if instance.Spec.Channel == "" {
		log.Error(fmt.Errorf("instance %s's channel is empty", instance.GetName()),
			" channel name can't be empty")
		return common.Result{Result: reconcile.Result{Requeue: true}},
			fmt.Errorf("channel name can't be empty")
	}
	if instance.Status.Phase == "" {
		instance.Status.Phase = current.ChaincodePhasePending
		instance.Status.Sequence = 1
		err := c.PatchStatus(context.TODO(), instance)
		return common.Result{}, err
	}
	if instance.Status.Phase == current.ChaincodePhasePending || instance.Status.Phase == current.ChaincodePhaseRunning {
		return common.Result{}, nil
	}

	if instance.Status.Phase == current.ChaincodePhaseUnapproved {
		log.Info("proposal has not been approved, stop processing.")
		return common.Result{}, nil
	}

	if instance.Status.Phase == current.ChaincodePhaseApproved {
		log.Info("proposal has been approved, cool...")
		if err := c.chaincodeProcess(instance); err != nil {
			log.Error(err, "chaincode process error")
			return common.Result{Result: reconcile.Result{Requeue: true}}, err
		}
		return common.Result{}, nil
	}

	return common.Result{}, nil
}

func (c *baseChaincode) Package(instance *current.Chaincode) error {
	method := fmt.Sprintf("%s base.chaincode.Package", stepPrefix)

	conditions := instance.Status.Conditions
	if conditions == nil {
		conditions = make([]current.ChaincodeCondition, 0, maxCondHistory)
	}
	expectCond := current.ChaincodeCondition{
		Type:               current.ChaincodeCondPackaged,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             "chaincode packaged successfully",
		Message:            "chaincode packaged successfully",
	}
	if _, err := c.PackageForK8s(instance); err != nil {
		log.Info("an error occurred in the packing process")
		expectCond.Type = current.ChaincodeCondError
		expectCond.Status = metav1.ConditionFalse
		expectCond.Reason = err.Error()
		expectCond.Message = "an error occurred in the packing process"
		expectCond.NextStage = current.ChaincodeCondPackaged
	}
	log.Info(fmt.Sprintf("%s package stage with condition %+v", method, expectCond))
	if len(conditions) == maxCondHistory {
		for i := 0; i < maxCondHistory-1; i++ {
			conditions[i] = conditions[i+1]
		}
		conditions[maxCondHistory-1] = expectCond
	} else {
		conditions = append(conditions, expectCond)
	}

	instance.Status.Conditions = conditions
	return c.PatchStatus(context.Background(), instance)
}

func (c *baseChaincode) Install(instance *current.Chaincode) error {
	method := fmt.Sprintf("%s base.chaincode.Install", stepPrefix)
	conditions := instance.Status.Conditions
	expectCond := current.ChaincodeCondition{
		Type:               current.ChaincodeCondInstalled,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             "chaincode installed successfully",
		Message:            "chaincode installed successfully",
	}
	reason, err := c.InstallChaincode(instance)
	expectCond.Reason = reason
	if err != nil {
		log.Info("an error occurred in the intalling process")
		expectCond.Type = current.ChaincodeCondError
		expectCond.Status = metav1.ConditionFalse
		expectCond.LastTransitionTime = metav1.Now()
		expectCond.Message = "ans error occurred in the installing process"
		expectCond.NextStage = current.ChaincodeCondInstalled
	}
	log.Info(fmt.Sprintf("%s install stage with condition %+v", method, expectCond))
	if len(conditions) == maxCondHistory {
		for i := 0; i < maxCondHistory-1; i++ {
			conditions[i] = conditions[i+1]
		}
		conditions[maxCondHistory-1] = expectCond
	} else {
		conditions = append(conditions, expectCond)
	}

	instance.Status.Conditions = conditions
	return c.PatchStatus(context.Background(), instance)
}

func (c *baseChaincode) Approve(instance *current.Chaincode) error {
	method := fmt.Sprintf("%s base.chaincode.Approve", stepPrefix)
	conditions := instance.Status.Conditions
	expectCond := current.ChaincodeCondition{
		Type:               current.ChaincodeCondApproved,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             "chaincode approved successfully",
		Message:            "chaincode approved successfully",
	}
	reason, err := c.ApproveChaincode(instance)
	expectCond.Reason = reason
	if err != nil {
		log.Info("an error occurred in the approving process")
		expectCond.Type = current.ChaincodeCondError
		expectCond.Status = metav1.ConditionFalse
		expectCond.LastTransitionTime = metav1.Now()
		expectCond.Message = "ans error occurred in the approving process"
		expectCond.NextStage = current.ChaincodeCondApproved
	}
	log.Info(fmt.Sprintf("%s approve stage with condition %+v", method, expectCond))

	if len(conditions) == maxCondHistory {
		for i := 0; i < maxCondHistory-1; i++ {
			conditions[i] = conditions[i+1]
		}
		conditions[maxCondHistory-1] = expectCond
	} else {
		conditions = append(conditions, expectCond)
	}

	instance.Status.Conditions = conditions
	return c.PatchStatus(context.Background(), instance)
}

func (c *baseChaincode) Commit(instance *current.Chaincode) error {
	method := fmt.Sprintf("%s base.chaincode.Commit", stepPrefix)
	conditions := instance.Status.Conditions
	expectCond := current.ChaincodeCondition{
		Type:               current.ChaincodeCondCommitted,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             "chaincode committed successfully",
		Message:            "chaincode committed successfully",
	}
	reason, err := c.CommitChaincode(instance)
	expectCond.Reason = reason
	if err != nil {
		log.Info("an error occurred in the commiting process")
		expectCond.Type = current.ChaincodeCondError
		expectCond.Status = metav1.ConditionFalse
		expectCond.LastTransitionTime = metav1.Now()
		expectCond.Message = "ans error occurred in the committing process"
		expectCond.NextStage = current.ChaincodeCondCommitted
	}

	log.Info(fmt.Sprintf("%s commit stage with condition %+v", method, expectCond))
	if len(conditions) == maxCondHistory {
		for i := 0; i < maxCondHistory-1; i++ {
			conditions[i] = conditions[i+1]
		}
		conditions[maxCondHistory-1] = expectCond
	} else {
		conditions = append(conditions, expectCond)
	}

	instance.Status.Conditions = conditions
	return c.PatchStatus(context.Background(), instance)
}

// R 检查失败，将type修改error，然后将Running状态放到NextStage
// 成功，将type修改为Running,同时,更新CR.Phase为Running即可.
func (c *baseChaincode) R(instance *current.Chaincode) error {
	method := fmt.Sprintf("%s base.chaincode.R", stepPrefix)
	conditions := instance.Status.Conditions
	expectCond := current.ChaincodeCondition{
		Type:               current.ChaincodeCondRunning,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             "chaincode is running",
		Message:            "chaincode is running",
	}
	prePhase := instance.Status.Phase
	instance.Status.Phase = current.ChaincodePhaseRunning
	reason, err := c.RunningChecker(instance)

	expectCond.Reason = reason
	if err != nil {
		expectCond.Type = current.ChaincodeCondError
		expectCond.Status = metav1.ConditionFalse
		expectCond.LastTransitionTime = metav1.Now()
		expectCond.Message = reason
		expectCond.NextStage = current.ChaincodeCondRunning
		instance.Status.Phase = prePhase
	}
	log.Info(fmt.Sprintf("%s check running stage with condition %+v", method, expectCond))
	if len(conditions) == maxCondHistory {
		for i := 0; i < maxCondHistory-1; i++ {
			conditions[i] = conditions[i+1]
		}
		conditions[maxCondHistory-1] = expectCond
	} else {
		conditions = append(conditions, expectCond)
	}

	instance.Status.Conditions = conditions
	return c.PatchStatus(context.Background(), instance)
}

var _ Chaincode = &baseChaincode{}

func New(client controllerclient.Client, o Override, scheme *runtime.Scheme, conf *config.Config) Chaincode {
	return &baseChaincode{client: client, override: o, Scheme: scheme, Config: conf}
}
