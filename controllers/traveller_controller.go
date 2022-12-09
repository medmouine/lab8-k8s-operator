/*
Copyright 2022.

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
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	oplab "github.com/example-inc/lab8-operator/api/v1"
)

type subReconciler interface {
	Reconcile(*oplab.Traveller) error
}

// TravellerReconciler reconciles a Traveller object
type TravellerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=example.com,resources=travellers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.com,resources=travellers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=example.com,resources=travellers/finalizers,verbs=update
func (r *TravellerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("Traveller", req.NamespacedName)

	var err error
	instance := &oplab.Traveller{}
	if stop := r.reconcileTraveller(req, instance, err); stop {
		return reconcile.Result{}, err
	}

	deplReconciler := deploymentReconciler{
		client: r.Client,
		Scheme: r.Scheme,
		log:    &logger,
	}

	svcReconciler := serviceReconciler{
		client: r.Client,
		Scheme: r.Scheme,
		log:    &logger,
	}

	for _, sr := range []subReconciler{deplReconciler, svcReconciler} {
		if err = sr.Reconcile(instance); err != nil {
			return reconcile.Result{}, err
		}
	}

	logger.Info("Skip reconcile: Deployment and service already exists")
	return reconcile.Result{}, nil
}

func (r *TravellerReconciler) reconcileTraveller(req ctrl.Request, instance *oplab.Traveller, err error) bool {
	err = r.Get(context.TODO(), req.NamespacedName, instance)

	if err != nil {
		if errors.IsNotFound(err) {
			err = nil
		}
		return true
	}

	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *TravellerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&oplab.Traveller{}).
		Complete(r)
}
