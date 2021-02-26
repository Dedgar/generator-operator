/*
Copyright 2020 Doug Edgar.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pingcap/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	managedv1alpha1 "github.com/dedgar/generator-operator/api/v1alpha1"
	"github.com/dedgar/generator-operator/k8s"
	corev1 "k8s.io/api/core/v1"
)

// ProxyServiceReconciler reconciles a ProxyService object
type ProxyServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=managed.openshift.io,resources=loggerservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=managed.openshift.io,resources=loggerservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=managed.openshift.io,resources=loggerservices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ProxyService object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *ProxyServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqProxy := r.Log.WithValues("loggerservice", req.NamespacedName)
	reqProxy.Info("Reconciling ProxyService")

	// Fetch the ProxyService instance
	instance := &managedv1alpha1.ProxyService{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Service object
	service := k8s.ProxyService(instance)

	// Set Proxy instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, service, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this DaemonSet already exists
	svcFound := &corev1.Service{}

	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, svcFound)
	if err != nil && errors.IsNotFound(err) {
		reqProxy.Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.Client.Create(context.TODO(), service)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Service created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Service already exists - don't requeue
	reqProxy.Info("Skip reconcile: Service already exists", "Service.Namespace", svcFound.Namespace, "Service.Name", svcFound.Name)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProxyServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&managedv1alpha1.ProxyService{}).
		Owns(&managedv1alpha1.Scanner{}).
		Complete(r)
}
