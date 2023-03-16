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
	"reflect"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	config "github.com/IBM-Blockchain/fabric-operator/operatorconfig"
	"github.com/IBM-Blockchain/fabric-operator/pkg/global"
	k8sclient "github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/common"
	k8schaincode "github.com/IBM-Blockchain/fabric-operator/pkg/offering/k8s/chaincode"
	"github.com/IBM-Blockchain/fabric-operator/pkg/operatorerrors"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	log = logf.Log.WithName("controller_chaincode")
)

// Add creates a new Vote Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, config *config.Config) error {
	r, err := newReconciler(mgr, config)
	if err != nil {
		return err
	}
	return add(mgr, r)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, cfg *config.Config) (*ReconcileChaincode, error) {
	c := k8sclient.New(mgr.GetClient(), &global.ConfigSetter{Config: cfg.Operator.Globals})
	scheme := mgr.GetScheme()

	cc := &ReconcileChaincode{
		client: c,
		scheme: scheme,
		Config: cfg,
	}

	switch cfg.Offering {
	case offering.K8S:
		cc.Offering = k8schaincode.New(c, scheme, cfg)
	default: // todo: maybe add OPENSHIFT support?
		return nil, errors.Errorf("offering %s not supported in Organization controller", cfg.Offering)
	}

	return cc, nil
}

// add adds a new Controller to mgr with r as the reconcile Reconciler
func add(mgr manager.Manager, r *ReconcileChaincode) error {
	predicateFuncs := predicate.Funcs{
		CreateFunc: r.CreateFunc,
		UpdateFunc: r.UpdateFunc,
		DeleteFunc: r.DeleteFunc,
	}

	c, err := controller.New("chaincode-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	if err = c.Watch(&source.Kind{Type: &current.Chaincode{}}, &handler.EnqueueRequestForObject{}, predicateFuncs); err != nil {
		return err
	}

	proposalPredicateFuncs := predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			return false
		},
		UpdateFunc: r.ProposalUpdateFunc,
		DeleteFunc: r.ProposalDeleteFunc,
	}
	// Watch for changes to secondary resource proposal
	if err = c.Watch(&source.Kind{Type: &current.Proposal{}}, &handler.EnqueueRequestForObject{}, proposalPredicateFuncs); err != nil {
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &ReconcileChaincode{}

//go:generate counterfeiter -o mocks/chaincodereconcile.go -fake-name ChaincodeReconcile . chaincodeReconcile

type chaincodeReconcile interface {
	Reconcile(chaincode *current.Chaincode) (common.Result, error)
}

// ReconcileChaincode reconciles a Vote object
type ReconcileChaincode struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client k8sclient.Client
	scheme *runtime.Scheme

	Offering chaincodeReconcile
	Config   *config.Config
}

// Reconcile reads that state of the cluster for a Vote object and makes changes based on the state read
// and what is in the Vote.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileChaincode) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	var err error

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Chaincode")

	instance := &current.Chaincode{}
	err = r.client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if instance.Labels == nil {
		instance.Labels = make(map[string]string)
	}
	update := false
	if instance.Labels[current.ChaincodeChannelLabel] != instance.Spec.Channel {
		update = true
		instance.Labels[current.ChaincodeChannelLabel] = instance.Spec.Channel
	}
	if instance.Labels[current.ChaincodeIDLabel] != instance.Spec.ID {
		update = true
		instance.Labels[current.ChaincodeIDLabel] = instance.Spec.ID
	}
	if instance.Labels[current.ChaincodeVersionLabel] != instance.Spec.Version {
		update = true
		instance.Labels[current.ChaincodeVersionLabel] = instance.Spec.Version
	}
	if instance.Labels[current.ChaincodeUsedEndorsementPolicy] != instance.Spec.EndorsePolicyRef.Name {
		update = true
		instance.Labels[current.ChaincodeUsedEndorsementPolicy] = instance.Spec.EndorsePolicyRef.Name
	}
	if update {
		reqLogger.Info(fmt.Sprintf("update chaincode %s's labels", instance.GetName()))
		return reconcile.Result{Requeue: true}, r.client.Update(context.TODO(), instance)
	}

	result, e := r.Offering.Reconcile(instance)
	if e != nil {
		return reconcile.Result{}, operatorerrors.IsBreakingError(errors.Wrapf(err, "Chaincode instance '%s' encountered error", instance.GetName()), "stopping reconcile loop", log)
	}

	reqLogger.Info(fmt.Sprintf("Finished reconciling Vote '%s'", instance.GetName()))
	return result.Result, nil
}

func (r ReconcileChaincode) CreateFunc(e event.CreateEvent) bool {
	x := e.Object.(*current.Chaincode)
	log.Info(fmt.Sprintf("chaincode:%s create proposal for creating chaincode", x.GetName()))
	return true
}

func (r *ReconcileChaincode) UpdateFunc(e event.UpdateEvent) bool {
	oldCC := e.ObjectOld.(*current.Chaincode)
	newCC := e.ObjectNew.(*current.Chaincode)

	log.Info(fmt.Sprintf("new chaincode phase: %s, old: %s", newCC.Status.Phase, oldCC.Status.Phase))
	if newCC.Status.Phase == current.ChaincodePhasePending || newCC.Status.Phase == current.ChaincodePhaseUnapproved {
		return false
	}

	r1 := !reflect.DeepEqual(oldCC.Spec, newCC.Spec) || !reflect.DeepEqual(oldCC.Status, newCC.Status)
	return r1 || newCC.Status.Phase == "" || newCC.Status.Phase == current.ChaincodePhaseApproved
}

func (r *ReconcileChaincode) ProposalUpdateFunc(e event.UpdateEvent) bool {
	newProposal := e.ObjectNew.(*current.Proposal)
	if newProposal.Labels == nil || newProposal.Labels[current.ChaincodeProposalLabel] == "" {
		return false
	}

	cr := &current.Chaincode{}
	if err := r.client.Get(context.TODO(),
		types.NamespacedName{Name: newProposal.Labels[current.ChaincodeProposalLabel]}, cr); err != nil {
		log.Error(err, fmt.Sprintf("failed to get chaincode %s ", newProposal.Labels[current.ChaincodeProposalLabel]))
		return false
	}
	log.Info(fmt.Sprintf("new proposal phase %s", newProposal.Status.Phase))
	if newProposal.Status.Phase != current.ProposalFinished {
		return false
	}

	for i := len(newProposal.Status.Conditions) - 1; i >= 0; i-- {
		cond := newProposal.Status.Conditions[i]
		if cond.Type == current.ProposalFailed {
			cr.Status.Phase = current.ChaincodePhaseUnapproved
			break
		}
		if cond.Type == current.ProposalSucceeded {
			cr.Status.Phase = current.ChaincodePhaseApproved
			break
		}
	}
	log.Info(fmt.Sprintf("proposal:%s done, chaincode status: %s", newProposal.GetName(), cr.Status.Phase))

	if cr.Status.Phase != current.ChaincodePhaseApproved && cr.Status.Phase != current.ChaincodePhaseUnapproved {
		log.Error(fmt.Errorf("expect %s or %s, but got %s",
			current.ChaincodePhaseApproved, current.ChaincodePhaseUnapproved, cr.Status.Phase), "")
		return false
	}

	if cr.Status.Phase == current.ChaincodePhaseApproved {
		chaincodeBuildName := cr.Spec.ExternalBuilder
		upgrade := false
		if newProposal.Spec.UpgradeChaincode != nil {
			upgrade = true
			chaincodeBuildName = newProposal.Spec.UpgradeChaincode.ExternalBuilder
		}

		if !upgrade && len(cr.Status.History) > 0 {
			log.Error(fmt.Errorf("already deployed chaincode is not allowed to be redeployed"), "")
			return false
		}

		image, digest, version, id, err := r.PickUpImageFromBuilder(chaincodeBuildName)
		if err != nil {
			log.Error(err, "the proposal passed, but failed to get the mirror information.")
			return false
		}
		originSpec := cr.Spec
		cr.Spec.Images.Name = image
		cr.Spec.Images.Digest = digest
		cr.Spec.Version = version
		cr.Spec.ID = id
		cr.Spec.ExternalBuilder = chaincodeBuildName

		if err := r.client.Patch(context.TODO(), cr, nil, k8sclient.PatchOption{
			Resilient: &k8sclient.ResilientPatch{
				Retry:    3,
				Into:     &current.Chaincode{},
				Strategy: client.MergeFrom,
			},
		}); err != nil {
			log.Error(err, fmt.Sprintf("failed to patch chaincode %s spec", cr.GetName()))
			return false
		}

		cr = &current.Chaincode{}
		if err := r.client.Get(context.TODO(),
			types.NamespacedName{Name: newProposal.Labels[current.ChaincodeProposalLabel]}, cr); err != nil {
			log.Error(err, fmt.Sprintf("failed to get chaincode %s ", newProposal.Labels[current.ChaincodeProposalLabel]))
			return false
		}
		cr.Status.Phase = current.ChaincodePhaseApproved
		if upgrade {
			if cr.Status.History == nil {
				cr.Status.History = make([]current.ChaincodeHistory, 0)
			}
			appendHistory := true
			for idx, item := range cr.Status.History {
				if item.Image.Name == originSpec.Images.Name &&
					item.Image.Digest == originSpec.Images.Digest &&
					item.Image.PullSecret == originSpec.Images.PullSecret &&
					item.Version == originSpec.Version &&
					item.ExternalBuilder == originSpec.ExternalBuilder {
					cr.Status.History[idx].UpgradeTime = v1.Now()
					appendHistory = false
					break
				}
			}
			if appendHistory {
				cr.Status.History = append(cr.Status.History, current.ChaincodeHistory{
					Version:         cr.Spec.Version,
					Image:           cr.Spec.Images,
					ExternalBuilder: chaincodeBuildName,
					UpgradeTime:     v1.Now(),
				})
			}
			cr.Status.Conditions = make([]current.ChaincodeCondition, 0)
			cr.Status.Sequence++
		}
	}

	if err := r.client.PatchStatus(context.TODO(), cr, nil, k8sclient.PatchOption{
		Resilient: &k8sclient.ResilientPatch{
			Retry:    3,
			Into:     &current.Chaincode{},
			Strategy: client.MergeFrom,
		},
	}); err != nil {
		log.Error(err, fmt.Sprintf("failed to patch chaincode %s status", cr.GetName()))
	}
	return false
}

func (r *ReconcileChaincode) DeleteFunc(e event.DeleteEvent) bool {
	cc := e.Object.(*current.Chaincode)
	if !cc.Spec.License.Accept {
		log.Info(fmt.Sprintf("chaincode:%s don't accept license", cc.GetName()))
		return false
	}
	log.Info(fmt.Sprintf("chaincode:%s deleted\n", cc.GetName()))

	go func() {
		pl := &current.ProposalList{}
		x := client.MatchingLabels(map[string]string{current.ChaincodeProposalLabel: cc.GetName()})
		if err := r.client.List(context.TODO(), pl, x); err != nil {
			log.Error(err, fmt.Sprintf("failed to list proposals when delete chaincode %s", cc.GetName()))
			return
		}
		for _, p := range pl.Items {
			r.client.Delete(context.TODO(), &p)
		}
	}()
	return false
}

func (r *ReconcileChaincode) ProposalDeleteFunc(e event.DeleteEvent) bool {
	proposal := e.Object.(*current.Proposal)
	log.Info(fmt.Sprintf("proposal:%s deleted", proposal.GetName()))
	return false
}

func (r *ReconcileChaincode) PickUpImageFromBuilder(builderName string) (string, string, string, string, error) {
	image, digest, version, id := "", "", "", ""
	if builderName == "" {
		return image, digest, version, id, fmt.Errorf("empty chiancodeBuilder name")
	}
	builder := &current.ChaincodeBuild{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: builderName}, builder); err != nil {
		log.Error(err, "trying to get the chaincode image, but failed to get the chaincodebuild")
		return image, digest, version, id, err
	}

	if len(builder.Status.PipelineRunResults) != 2 {
		return image, digest, version, id, fmt.Errorf("expect only 2 elements, but have %d", len(builder.Status.PipelineRunResults))
	}
	version = builder.Spec.Version
	id = builder.Spec.ID
	for _, item := range builder.Status.PipelineRunResults {
		if item.Name == current.IMAGE_URL {
			image = item.Value
		}
		if item.Name == current.IMAGE_DIGEST {
			digest = item.Value
		}
	}
	return image, digest, version, id, nil
}
