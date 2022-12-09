package controllers

import (
	"context"
	oplab "github.com/example-inc/lab8-operator/api/v1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type deploymentReconciler struct {
	client client.Client
	Scheme *runtime.Scheme
	log    *logr.Logger
}

func (d deploymentReconciler) Reconcile(parent *oplab.Traveller) error {
	depl := &appsv1.Deployment{}
	err := d.client.Get(context.TODO(), types.NamespacedName{
		Name:      parent.Name,
		Namespace: parent.Namespace,
	}, depl)

	if errors.IsNotFound(err) {
		err = d.createDeployment(parent)
		if err != nil {
			d.log.Error(err, "Failed to create deployment")
		}
	}

	if err != nil {
		d.log.Error(err, "Failed to reconcile Deployment")
	}

	return nil
}

func (d deploymentReconciler) getDeplDefinition(parent *oplab.Traveller) *appsv1.Deployment {
	labels := map[string]string{
		"app":             "visitors",
		"visitorssite_cr": parent.Name,
		"tier":            "backend",
	}
	size := int32(1)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "lab8-deployment",
			Namespace: parent.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &size,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           "paulbouwer/hello-kubernetes:1.10",
						ImagePullPolicy: corev1.PullAlways,
						Name:            "lab8-pod",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "hello",
						}},
					}},
				},
			},
		},
	}
}

func (d deploymentReconciler) createDeployment(parent *oplab.Traveller) error {
	depl := d.getDeplDefinition(parent)
	if err := controllerutil.SetControllerReference(parent, depl, d.Scheme); err != nil {
		d.log.Error(err, "Failed to set owner reference")
		return err
	}
	return d.client.Create(context.TODO(), depl)
}
