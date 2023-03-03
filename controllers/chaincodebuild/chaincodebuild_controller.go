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

package chaincodebuild

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	ctrl "sigs.k8s.io/controller-runtime"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	config "github.com/IBM-Blockchain/fabric-operator/operatorconfig"
	"github.com/IBM-Blockchain/fabric-operator/pkg/global"
	k8sclient "github.com/IBM-Blockchain/fabric-operator/pkg/k8s/controllerclient"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering"
	"github.com/IBM-Blockchain/fabric-operator/pkg/operatorerrors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	baseccb "github.com/IBM-Blockchain/fabric-operator/pkg/offering/base/chaincodebuild"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/common"
	k8sccb "github.com/IBM-Blockchain/fabric-operator/pkg/offering/k8s/chaincodebuild"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	KIND = "ChaincodeBuild"
	// Labels added in ChaincodeBuild
	CHAINCODEBUILD_NETWORK_LABEL   = "bestchains.chaincodebuild.network"
	CHAINCODEBUILD_ID_LABEL        = "bestchains.chaincodebuild.id"
	CHAINCODEBUILD_VERSION_LABEL   = "bestchains.chaincodebuild.version"
	CHAINCODEBUILD_INITIATOR_LABEL = "bestchains.chaincodebuild.initiator"
)

var log = logf.Log.WithName("controller_chaincodebuild")

// Add creates a new ChaincodeBuild Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, cfg *config.Config) error {
	r, err := newReconciler(mgr, cfg)
	if err != nil {
		return err
	}
	return add(mgr, r)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, cfg *config.Config) (*ReconcileChaincodeBuild, error) {
	client := k8sclient.New(mgr.GetClient(), &global.ConfigSetter{Config: cfg.Operator.Globals})
	scheme := mgr.GetScheme()

	chaincodeBuild := &ReconcileChaincodeBuild{
		client: client,
		scheme: scheme,
		Config: cfg,
		update: map[string][]Update{},
		mutex:  &sync.Mutex{},
	}

	switch cfg.Offering {
	case offering.K8S:
		chaincodeBuild.Offering = k8sccb.New(client, scheme, cfg)
	default:
		return nil, errors.Errorf("offering %s not supported in ChaincodeBuild controller", cfg.Offering)
	}

	return chaincodeBuild, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r *ReconcileChaincodeBuild) error {
	c, err := controller.New("chaincodebuild-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ChaincodeBuild
	predicateFuncs := predicate.Funcs{
		CreateFunc: r.CreateFunc,
	}

	err = c.Watch(&source.Kind{Type: &current.ChaincodeBuild{}}, &handler.EnqueueRequestForObject{}, predicateFuncs)
	if err != nil {
		return err
	}

	pipelineFuncs := predicate.Funcs{
		UpdateFunc: r.PipelineRunUpdateFunc,
	}
	err = c.Watch(&source.Kind{Type: &pipelinev1beta1.PipelineRun{}}, &handler.EnqueueRequestForObject{}, pipelineFuncs)
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileChaincodeBuild{}

//go:generate counterfeiter -o mocks/Reconcile.go -fake-name ChaincodeBuildReconcile . chaincodeBuildReconcile
//counterfeiter:generate . chaincodeBuildReconcile
type chaincodeBuildReconcile interface {
	Reconcile(*current.ChaincodeBuild, baseccb.Update) (common.Result, error)
}

// ReconcileChaincodeBuild reconciles a ChaincodeBuild object
type ReconcileChaincodeBuild struct {
	client k8sclient.Client
	scheme *runtime.Scheme

	Offering chaincodeBuildReconcile
	Config   *config.Config

	update map[string][]Update
	mutex  *sync.Mutex
}

func (r *ReconcileChaincodeBuild) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&current.ChaincodeBuild{}).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ChaincodeBuild object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
// +kubebuilder:rbac:groups="",resources=events;configmaps;secrets,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=ibp.com,resources=chaincodebuilds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ibp.com,resources=chaincodebuilds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ibp.com,resources=chaincodebuilds/finalizers,verbs=update
func (r *ReconcileChaincodeBuild) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	var err error
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	reqLogger.Info("Reconciling ChaincodeBuild")

	instance := &current.ChaincodeBuild{}
	err = r.client.Get(context.TODO(), request.NamespacedName, instance)
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

	if r.SetLabels(instance) {
		log.Info(fmt.Sprintf("update chaincodebuild %s's labels", instance.GetName()))
		return reconcile.Result{Requeue: true}, r.client.Update(context.TODO(), instance)
	}

	update := r.GetUpdateStatus(instance)
	reqLogger.Info(fmt.Sprintf("Reconciling ChaincodeBuild '%s' with update values of [ %+v ]", instance.GetName(), update.GetUpdateStackWithTrues()))

	result, err := r.Offering.Reconcile(instance, r.PopUpdate(instance.GetName()))
	if err != nil {
		if setStatuErr := r.SetErrorStatus(instance, err); setStatuErr != nil {
			return reconcile.Result{}, operatorerrors.IsBreakingError(setStatuErr, "failed to update status", log)
		}
		return reconcile.Result{}, operatorerrors.IsBreakingError(errors.Wrapf(err, "ChaincodeBuild instance '%s' encountered error", instance.GetName()), "stopping reconcile loop", log)
	} else {
		setStatusErr := r.SetStatus(instance, result.Status)
		if setStatusErr != nil {
			return reconcile.Result{}, operatorerrors.IsBreakingError(setStatusErr, "failed to update status", log)
		}
	}

	if result.Requeue {
		r.PushUpdate(instance.GetName(), *update)
	}

	reqLogger.Info(fmt.Sprintf("Finished reconciling ChaincodeBuild '%s' with update values of [ %+v ]", instance.GetName(), update.GetUpdateStackWithTrues()))

	// If the stack still has items that require processing, keep reconciling
	// until the stack has been cleared
	_, found := r.update[instance.GetName()]
	if found {
		if len(r.update[instance.GetName()]) > 0 {
			return reconcile.Result{
				Requeue: true,
			}, nil
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileChaincodeBuild) SetLabels(instance *current.ChaincodeBuild) bool {
	if instance.Labels == nil {
		instance.Labels = make(map[string]string)
	}

	labelUpdate := false
	if instance.Labels[CHAINCODEBUILD_NETWORK_LABEL] != instance.Spec.Network {
		labelUpdate = true
		instance.Labels[CHAINCODEBUILD_NETWORK_LABEL] = instance.Spec.Network
	}
	if instance.Labels[CHAINCODEBUILD_INITIATOR_LABEL] != instance.Spec.Initiator {
		labelUpdate = true
		instance.Labels[CHAINCODEBUILD_INITIATOR_LABEL] = instance.Spec.Initiator
	}
	if instance.Labels[CHAINCODEBUILD_ID_LABEL] != instance.Spec.ID {
		labelUpdate = true
		instance.Labels[CHAINCODEBUILD_ID_LABEL] = instance.Spec.ID
	}
	if instance.Labels[CHAINCODEBUILD_VERSION_LABEL] != instance.Spec.Version {
		labelUpdate = true
		instance.Labels[CHAINCODEBUILD_VERSION_LABEL] = instance.Spec.Version
	}

	return labelUpdate
}

func (r *ReconcileChaincodeBuild) SetStatus(instance *current.ChaincodeBuild, reconcileStatus *current.CRStatus) error {
	var err error

	log.Info(fmt.Sprintf("Setting status for '%s'", instance.GetName()))

	if err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}, instance); err != nil {
		return err
	}

	if err = r.SaveSpecState(instance); err != nil {
		return errors.Wrap(err, "failed to save spec state")
	}

	status := instance.Status.CRStatus

	// Check if reconcile loop returned an updated status that differs from exisiting status.
	// If so, set status to the reconcile status.
	if reconcileStatus != nil {
		if instance.Status.Type != reconcileStatus.Type || instance.Status.Reason != reconcileStatus.Reason || instance.Status.Message != reconcileStatus.Message {
			status.Type = reconcileStatus.Type
			status.Status = current.True
			status.Reason = reconcileStatus.Reason
			status.Message = reconcileStatus.Message
			status.LastHeartbeatTime = metav1.Now()

			instance.Status = current.ChaincodeBuildStatus{
				CRStatus: status,
			}

			log.Info(fmt.Sprintf("Updating status of ChaincodeBuild custom resource to %s phase", instance.Status.Type))
			err = r.client.PatchStatus(context.TODO(), instance, nil, k8sclient.PatchOption{
				Resilient: &k8sclient.ResilientPatch{
					Retry:    2,
					Into:     &current.ChaincodeBuild{},
					Strategy: client.MergeFrom,
				},
			})
			if err != nil {
				return err
			}

			return nil
		}
	}

	return nil
}

func (r *ReconcileChaincodeBuild) SetErrorStatus(instance *current.ChaincodeBuild, reconcileErr error) error {
	var err error

	if err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}, instance); err != nil {
		return err
	}

	if err = r.SaveSpecState(instance); err != nil {
		return errors.Wrap(err, "failed to save spec state")
	}

	log.Info(fmt.Sprintf("Setting error status for '%s'", instance.GetName()))

	status := instance.Status.CRStatus
	status.Type = current.Error
	status.Status = current.True
	status.Reason = "errorOccurredDuringReconcile"
	status.Message = reconcileErr.Error()
	status.LastHeartbeatTime = metav1.Now()
	status.ErrorCode = operatorerrors.GetErrorCode(reconcileErr)

	instance.Status = current.ChaincodeBuildStatus{
		CRStatus: status,
	}

	log.Info(fmt.Sprintf("Updating status of ChaincodeBuild custom resource to %s phase", instance.Status.Type))
	if err = r.client.PatchStatus(context.TODO(), instance, nil, k8sclient.PatchOption{
		Resilient: &k8sclient.ResilientPatch{
			Retry:    2,
			Into:     &current.ChaincodeBuild{},
			Strategy: client.MergeFrom,
		},
	}); err != nil {
		return err
	}

	return nil

}

func (r *ReconcileChaincodeBuild) SaveSpecState(instance *current.ChaincodeBuild) error {
	data, err := yaml.Marshal(instance.Spec)
	if err != nil {
		return err
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("chan-%s-spec", instance.GetName()),
			Namespace: r.Config.Operator.Namespace,
			Labels:    instance.GetLabels(),
		},
		BinaryData: map[string][]byte{
			"spec": data,
		},
	}

	err = r.client.CreateOrUpdate(context.TODO(), cm, k8sclient.CreateOrUpdateOption{
		Owner:  instance,
		Scheme: r.scheme,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *ReconcileChaincodeBuild) GetSpecState(instance *current.ChaincodeBuild) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}
	nn := types.NamespacedName{
		Name:      fmt.Sprintf("chan-%s-spec", instance.GetName()),
		Namespace: r.Config.Operator.Namespace,
	}

	err := r.client.Get(context.TODO(), nn, cm)
	if err != nil {
		return nil, err
	}

	return cm, nil
}
