package controllers

import (
	"context"
	oplab "github.com/example-inc/lab8-operator/api/v1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type serviceReconciler struct {
	client client.Client
	Scheme *runtime.Scheme
	log    *logr.Logger
}

func (s serviceReconciler) Reconcile(parent *oplab.Traveller) error {
	svc := &corev1.Service{}
	err := s.client.Get(context.TODO(), types.NamespacedName{
		Name:      parent.Name,
		Namespace: parent.Namespace,
	}, svc)

	if errors.IsNotFound(err) {
		s.log.Info("Creating service")
		err = s.createService(parent)
		if err != nil {
			s.log.Error(err, "Failed to create service")
			return err
		}
	}

	return err
}

func (s serviceReconciler) getServiceDefinition(parent *oplab.Traveller) *corev1.Service {
	labels := map[string]string{
		"app":             "visitors",
		"visitorssite_cr": parent.Name,
		"tier":            "backend",
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "lab8-service",
			Namespace: parent.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       80,
				TargetPort: intstr.FromInt(8080),
				NodePort:   30685,
			}},
			Type: corev1.ServiceTypeNodePort,
		},
	}

}

func (s serviceReconciler) createService(parent *oplab.Traveller) error {
	svc := s.getServiceDefinition(parent)
	if err := controllerutil.SetControllerReference(parent, svc, s.Scheme); err != nil {
		s.log.Error(err, "Failed to set owner reference on service")
		return err
	}
	return s.client.Create(context.TODO(), svc)
}
