/*
Copyright 2025.

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

package controller

import (
	"context"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appscodev1 "refat.kubebuilder.io/project/api/v1"
)

// ApiserverReconciler reconciles a Apiserver object
type ApiserverReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=appscode.refat.kubebuilder.io,resources=apiservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=appscode.refat.kubebuilder.io,resources=apiservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=appscode.refat.kubebuilder.io,resources=apiservers/finalizers,verbs=update

// For Deployment (under apps/v1)
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get

// For Service (core API group)
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Apiserver object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *ApiserverReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// TODO(user): your logic here
	var apiserver appscodev1.Apiserver
	if err := r.Get(ctx, req.NamespacedName, &apiserver); err != nil {
		log.Error(err, "unable to fetch Apiserver")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("Reconciling Apiserver", "Name", apiserver.Name, "Namespace", apiserver.Namespace)
	log.Info("Spec", "Spec", apiserver.Spec, "Status", apiserver.Status)

	//Generate Desired Deployment
	desiredDeployment := r.buildDeployment(&apiserver)

	//Getting the existing deployment
	var existingDeployment appsv1.Deployment
	err := r.Get(ctx, types.NamespacedName{
		Name:      apiserver.Name,
		Namespace: apiserver.Namespace,
	}, &existingDeployment)

	if err != nil && apierror.IsNotFound(err) {
		//Deployment does not exist --> create it
		if err := controllerutil.SetControllerReference(&apiserver, desiredDeployment, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		log.Info("Creating Deployment........")
		err = r.Create(ctx, desiredDeployment)
	} else if err == nil {
		log.Info("Deployment Exist, Updating Deployment........")
		patch := client.MergeFrom(existingDeployment.DeepCopy())
		existingDeployment.Spec = desiredDeployment.Spec

		if err := r.Patch(ctx, &existingDeployment, patch); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApiserverReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appscodev1.Apiserver{}).
		Owns(&appsv1.Deployment{}).
		Named("apiserver").
		Complete(r)
}

func (r *ApiserverReconciler) buildDeployment(apiserver *appscodev1.Apiserver) *appsv1.Deployment {
	labels := map[string]string{
		"app": apiserver.Name,
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apiserver.Name,
			Namespace: apiserver.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: apiserver.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "apiserver",
							Image: apiserver.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: apiserver.Spec.Port,
								},
							},
						},
					},
				},
			},
		},
	}
}
