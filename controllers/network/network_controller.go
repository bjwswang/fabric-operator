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

package network

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

	bcrbac "github.com/IBM-Blockchain/fabric-operator/pkg/rbac"

	basenet "github.com/IBM-Blockchain/fabric-operator/pkg/offering/base/network"
	"github.com/IBM-Blockchain/fabric-operator/pkg/offering/common"
	k8snet "github.com/IBM-Blockchain/fabric-operator/pkg/offering/k8s/network"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KIND                    = "Network"
	NETWORK_INITIATOR_LABEL = "bestchains.network.initiator"
)

var log = logf.Log.WithName("controller_network")

// Add creates a new Network Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, cfg *config.Config) error {
	r, err := newReconciler(mgr, cfg)
	if err != nil {
		return err
	}
	return add(mgr, r)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, cfg *config.Config) (*ReconcileNetwork, error) {
	client := k8sclient.New(mgr.GetClient(), &global.ConfigSetter{Config: cfg.Operator.Globals})
	scheme := mgr.GetScheme()

	network := &ReconcileNetwork{
		client:      client,
		scheme:      scheme,
		Config:      cfg,
		update:      map[string][]Update{},
		mutex:       &sync.Mutex{},
		rbacManager: bcrbac.NewRBACManager(client, nil),
	}

	switch cfg.Offering {
	case offering.K8S:
		network.Offering = k8snet.New(client, scheme, cfg)
	default:
		return nil, errors.Errorf("offering %s not supported in Network controller", cfg.Offering)
	}

	return network, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r *ReconcileNetwork) error {
	// Create a new controller
	predicateFuncs := predicate.Funcs{
		CreateFunc: r.CreateFunc,
		UpdateFunc: r.UpdateFunc,
		DeleteFunc: r.DeleteFunc,
	}

	c, err := controller.New("network-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Network
	err = c.Watch(&source.Kind{Type: &current.Network{}}, &handler.EnqueueRequestForObject{}, predicateFuncs)
	if err != nil {
		return err
	}

	channelPredictFuncs := predicate.Funcs{
		CreateFunc: r.ChannelCreateFunc,
		DeleteFunc: r.ChannelDeleteFunc,
	}

	err = c.Watch(&source.Kind{Type: &current.Channel{}}, &handler.EnqueueRequestForObject{}, channelPredictFuncs)
	if err != nil {
		return err
	}

	ordererPredictFuncs := predicate.Funcs{
		UpdateFunc: r.OrdererUpdateFunc,
	}

	err = c.Watch(&source.Kind{Type: &current.IBPOrderer{}}, &handler.EnqueueRequestForOwner{
		OwnerType:    &current.Network{},
		IsController: true,
	}, ordererPredictFuncs)
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileNetwork{}

//go:generate counterfeiter -o mocks/networkReconcile.go -fake-name NetworkReconcile . networkReconcile
//counterfeiter:generate . networkReconcile
type networkReconcile interface {
	Reconcile(*current.Network, basenet.Update) (common.Result, error)
}

// ReconcileNetwork reconciles a Network object
type ReconcileNetwork struct {
	client k8sclient.Client
	scheme *runtime.Scheme

	Offering networkReconcile
	Config   *config.Config

	update map[string][]Update
	mutex  *sync.Mutex

	rbacManager *bcrbac.Manager
}

func (r *ReconcileNetwork) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&current.Network{}).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Network object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
// +kubebuilder:rbac:groups="",resources=events;configmaps;secrets,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=ibp.com,resources=Networks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ibp.com,resources=Networks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ibp.com,resources=Networks/finalizers,verbs=update
func (r *ReconcileNetwork) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	var err error
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	reqLogger.Info("Reconciling Network")

	instance := &current.Network{}
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

	if instance.Labels == nil {
		instance.Labels = make(map[string]string)
	}

	if _, ok := instance.Labels[NETWORK_INITIATOR_LABEL]; !ok {
		instance.Labels[NETWORK_INITIATOR_LABEL] = instance.Spec.Initiator
		err = r.client.Update(context.TODO(), instance)
		return reconcile.Result{Requeue: true}, err
	}

	update := r.GetUpdateStatus(instance)
	reqLogger.Info(fmt.Sprintf("Reconciling Network '%s' with update values of [ %+v ]", instance.GetName(), update.GetUpdateStackWithTrues()))

	result, err := r.Offering.Reconcile(instance, r.PopUpdate(instance.GetName()))
	if err != nil {
		if setStatuErr := r.SetErrorStatus(instance, err); setStatuErr != nil {
			return reconcile.Result{}, operatorerrors.IsBreakingError(setStatuErr, "failed to update status", log)
		}
		return reconcile.Result{}, operatorerrors.IsBreakingError(errors.Wrapf(err, "Network instance '%s' encountered error", instance.GetName()), "stopping reconcile loop", log)
	} else {
		setStatusErr := r.SetStatus(instance, result.Status)
		if setStatusErr != nil {
			return reconcile.Result{}, operatorerrors.IsBreakingError(setStatusErr, "failed to update status", log)
		}
	}

	if result.Requeue {
		r.PushUpdate(instance.GetName(), *update)
	}

	reqLogger.Info(fmt.Sprintf("Finished reconciling Network '%s' with update values of [ %+v ]", instance.GetName(), update.GetUpdateStackWithTrues()))

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

// TODO: NetworkStatus relevant
func (r *ReconcileNetwork) SetStatus(instance *current.Network, reconcileStatus *current.CRStatus) error {
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
			status.LastHeartbeatTime = v1.Now()

			instance.Status = current.NetworkStatus{
				CRStatus: status,
			}

			log.Info(fmt.Sprintf("Updating status of Network custom resource to %s phase", instance.Status.Type))
			err = r.client.PatchStatus(context.TODO(), instance, nil, k8sclient.PatchOption{
				Resilient: &k8sclient.ResilientPatch{
					Retry:    2,
					Into:     &current.Network{},
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

func (r *ReconcileNetwork) SetErrorStatus(instance *current.Network, reconcileErr error) error {
	var err error

	if err = r.SaveSpecState(instance); err != nil {
		return errors.Wrap(err, "failed to save spec state")
	}

	log.Info(fmt.Sprintf("Setting error status for '%s'", instance.GetName()))

	status := instance.Status.CRStatus
	status.Type = current.Error
	status.Status = current.True
	status.Reason = "errorOccurredDuringReconcile"
	status.Message = reconcileErr.Error()
	status.LastHeartbeatTime = v1.Now()
	status.ErrorCode = operatorerrors.GetErrorCode(reconcileErr)

	instance.Status = current.NetworkStatus{
		CRStatus: status,
	}

	log.Info(fmt.Sprintf("Updating status of Network custom resource to %s phase", instance.Status.Type))
	if err = r.client.PatchStatus(context.TODO(), instance, nil, k8sclient.PatchOption{
		Resilient: &k8sclient.ResilientPatch{
			Retry:    2,
			Into:     &current.Network{},
			Strategy: client.MergeFrom,
		},
	}); err != nil {
		return err
	}

	return nil

}

func (r *ReconcileNetwork) SaveSpecState(instance *current.Network) error {
	data, err := yaml.Marshal(instance.Spec)
	if err != nil {
		return err
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("net-%s-spec", instance.GetName()),
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

func (r *ReconcileNetwork) GetSpecState(instance *current.Network) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}
	nn := types.NamespacedName{
		Name:      fmt.Sprintf("net-%s-spec", instance.GetName()),
		Namespace: r.Config.Operator.Namespace,
	}

	err := r.client.Get(context.TODO(), nn, cm)
	if err != nil {
		return nil, err
	}

	return cm, nil
}
