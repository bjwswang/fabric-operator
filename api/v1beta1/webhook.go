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

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/IBM-Blockchain/fabric-operator/pkg/util"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func AddWebhooks(mgr ctrl.Manager, setupLog logr.Logger) (err error) {
	operatorUser, err := getOperatorUser(mgr.GetClient())
	if err != nil {
		setupLog.Error(err, "unable to get operatorUser")
		return err
	}
	if err = registerCustomWebhook(mgr, &Vote{}, operatorUser); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Vote")
		return err
	}
	if err = registerCustomWebhook(mgr, &Proposal{}, operatorUser); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Proposal")
		return err
	}
	if err = registerCustomWebhook(mgr, &Network{}, operatorUser); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Network")
		return err
	}
	if err = registerCustomWebhook(mgr, &Federation{}, operatorUser); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Federation")
		return err
	}
	if err = registerCustomWebhook(mgr, &Organization{}, operatorUser); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Organization")
		return err
	}
	if err = registerCustomWebhook(mgr, &Channel{}, operatorUser); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Channel")
		return err
	}
	if err = registerCustomWebhook(mgr, &Chaincode{}, operatorUser); err != nil {
		setupLog.Error(err, "unable create webhook", "webhook", "Chaincode")
	}
	if err = registerCustomWebhook(mgr, &EndorsePolicy{}, operatorUser); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "EndorsePolicy")
	}
	if err = registerCustomWebhook(mgr, &ChaincodeBuild{}, operatorUser); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ChaincodeBuild")
	}
	return nil
}

func getOperatorUser(c client.Client) (operatorUser string, err error) {
	name := os.Getenv("OPERATOR_NAME")
	if name == "" {
		return "", fmt.Errorf("env OPERATOR_NAME not set")
	}
	ns, err := util.GetNamespace()
	if err != nil {
		return "", err
	}
	pod := &corev1.Pod{}
	pod.Name = name
	pod.Namespace = ns
	if err := c.Get(context.TODO(), client.ObjectKeyFromObject(pod), pod); err != nil {
		return "", err
	}
	sa := pod.Spec.ServiceAccountName
	if sa == "" {
		sa = pod.Spec.DeprecatedServiceAccount
	}
	return "system:serviceaccount:" + ns + ":" + sa, nil
}

func registerCustomWebhook(mgr ctrl.Manager, apiType runtime.Object, operatorUser string) error {
	gvk, err := apiutil.GVKForObject(apiType, mgr.GetScheme())
	if err != nil {
		return err
	}

	path := generateMutatePath(gvk)
	defaulter, ok := apiType.(defaulter)
	if ok {
		mh := mutatingHandler{operatorUser: operatorUser, defaulter: defaulter, client: mgr.GetClient()}
		dwh := &admission.Webhook{
			Handler: &mh,
		}
		mgr.GetWebhookServer().Register(path, dwh)
	}

	path = generateValidatePath(gvk)
	validator, ok := apiType.(validator)
	if ok {
		vh := validatingHandler{operatorUser: operatorUser, validator: validator, client: mgr.GetClient()}
		vwh := &admission.Webhook{
			Handler: &vh,
		}
		mgr.GetWebhookServer().Register(path, vwh)
	}

	return nil
}

func generateMutatePath(gvk schema.GroupVersionKind) string {
	return "/mutate-" + strings.ReplaceAll(gvk.Group, ".", "-") + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind)
}

func generateValidatePath(gvk schema.GroupVersionKind) string {
	return "/validate-" + strings.ReplaceAll(gvk.Group, ".", "-") + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind)
}

var _ webhook.AdmissionHandler = &validatingHandler{}

type validatingHandler struct {
	operatorUser string
	validator    validator
	client       client.Client
	decoder      *admission.Decoder
}

// InjectDecoder injects the decoder into a validatingHandler.
func (h *validatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *validatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	if h.validator == nil {
		panic("validator should never be nil")
	}

	ctx = newContextWithOperatorUser(ctx, h.operatorUser)
	user := req.UserInfo

	// Get the object in the request
	obj := h.validator.DeepCopyObject().(validator)
	if req.Operation == v1.Create {
		err := h.decoder.Decode(req, obj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		err = obj.ValidateCreate(ctx, h.client, user)
		if err != nil {
			var apiStatus apierrors.APIStatus
			if errors.As(err, &apiStatus) {
				return validationResponseFromStatus(false, apiStatus.Status())
			}
			return admission.Denied(err.Error())
		}
	}

	if req.Operation == v1.Update {
		oldObj := obj.DeepCopyObject()

		err := h.decoder.DecodeRaw(req.Object, obj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		err = h.decoder.DecodeRaw(req.OldObject, oldObj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		err = obj.ValidateUpdate(ctx, h.client, oldObj, user)
		if err != nil {
			var apiStatus apierrors.APIStatus
			if errors.As(err, &apiStatus) {
				return validationResponseFromStatus(false, apiStatus.Status())
			}
			return admission.Denied(err.Error())
		}
	}

	if req.Operation == v1.Delete {
		// In reference to PR: https://github.com/kubernetes/kubernetes/pull/76346
		// OldObject contains the object being deleted
		err := h.decoder.DecodeRaw(req.OldObject, obj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		err = obj.ValidateDelete(ctx, h.client, user)
		if err != nil {
			var apiStatus apierrors.APIStatus
			if errors.As(err, &apiStatus) {
				return validationResponseFromStatus(false, apiStatus.Status())
			}
			return admission.Denied(err.Error())
		}
	}

	return admission.Allowed("")
}

// validationResponseFromStatus returns a response for admitting a request with provided Status object.
func validationResponseFromStatus(allowed bool, status metav1.Status) admission.Response {
	resp := admission.Response{
		AdmissionResponse: v1.AdmissionResponse{
			Allowed: allowed,
			Result:  &status,
		},
	}
	return resp
}

type validator interface {
	runtime.Object
	ValidateCreate(ctx context.Context, client client.Client, user authenticationv1.UserInfo) error
	ValidateUpdate(ctx context.Context, client client.Client, old runtime.Object, user authenticationv1.UserInfo) error
	ValidateDelete(ctx context.Context, client client.Client, user authenticationv1.UserInfo) error
}

var _ admission.DecoderInjector = &mutatingHandler{}

type mutatingHandler struct {
	operatorUser string
	defaulter    defaulter
	client       client.Client
	decoder      *admission.Decoder
}

// InjectDecoder injects the decoder into a mutatingHandler.
func (h *mutatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// Handle handles admission requests.
func (h *mutatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	if h.defaulter == nil {
		panic("defaulter should never be nil")
	}

	ctx = newContextWithOperatorUser(ctx, h.operatorUser)

	// Get the object in the request
	obj := h.defaulter.DeepCopyObject().(defaulter)
	if err := h.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Default the object
	obj.Default(ctx, h.client, req.UserInfo)
	marshalled, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// Create the patch
	return admission.PatchResponseFromRaw(req.Object.Raw, marshalled)
}

type defaulter interface {
	runtime.Object
	Default(ctx context.Context, client client.Client, user authenticationv1.UserInfo)
}

func isSuperUser(ctx context.Context, user authenticationv1.UserInfo) bool {
	operatorUser, _ := operatorUserFromContext(ctx)
	return operatorUser == user.Username || util.ContainsValue("system:masters", user.Groups) || util.ContainsValue("system:serviceaccounts:kube-system", user.Groups)
}

type operatorUserContextKey struct{}

func operatorUserFromContext(ctx context.Context) (operatorUser string, err error) {
	if v, ok := ctx.Value(operatorUserContextKey{}).(string); ok {
		return v, nil
	}

	return "", errors.New("operatorUser not found in context")
}

func newContextWithOperatorUser(ctx context.Context, operatorUser string) context.Context {
	return context.WithValue(ctx, operatorUserContextKey{}, operatorUser)
}
