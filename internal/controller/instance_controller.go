/*
Copyright 2026.

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
	"fmt"
	"maps"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	gatusiov1alpha1 "github.com/Jannik-Hm/Gatus-Operator/api/v1alpha1"
	"github.com/Jannik-Hm/Gatus-Operator/internal/config"
)

// InstanceReconciler reconciles a Instance object
type InstanceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config *config.Config
}

// +kubebuilder:rbac:groups=gatus.io,resources=instances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gatus.io,resources=instances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=gatus.io,resources=instances/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Instance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.3/pkg/reconcile
func (r *InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var instance gatusiov1alpha1.Instance
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			log.Info("Instance resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Instance")
		return ctrl.Result{}, err
	}

	correspondingDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
	}

	// TODO:
	// // create/update config
	// op, err := controllerutil.CreateOrUpdate(ctx, r.Client, correspondingDeployment, func() error { return MutateDeployment(&instance, correspondingDeployment, r.Scheme) })

	// if err != nil {
	// 	log.Error(err, "Failed to create/update Deployment")
	// 	return ctrl.Result{}, err
	// }

	// if op != controllerutil.OperationResultNone {
	// 	log.Info("Deployment reconciled", "Operation", op)
	// }

	// create/update deployment
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, correspondingDeployment, func() error { return MutateDeployment(&instance, correspondingDeployment, r.Scheme) })

	if err != nil {
		log.Error(err, "Failed to create/update Deployment")
		return ctrl.Result{}, err
	}

	if op != controllerutil.OperationResultNone {
		log.Info("Deployment reconciled", "Operation", op)
	}

	// create/update service
	if instance.Spec.Service != nil && instance.Spec.Service.Enabled {
		correspondingService := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name,
				Namespace: req.Namespace,
			},
		}

		op, err = controllerutil.CreateOrUpdate(ctx, r.Client, correspondingService, func() error { return MutateService(&instance, correspondingService, r.Scheme) })

		if err != nil {
			log.Error(err, "Failed to create/update Deployment")
			return ctrl.Result{}, err
		}

		if op != controllerutil.OperationResultNone {
			log.Info("Deployment reconciled", "Operation", op)
		}
	} else {
		var service corev1.Service
		err := r.Get(ctx, req.NamespacedName, &service)

		if err != nil && !apierrors.IsNotFound(err) {
			log.Error(err, "Failed to check if service exists")
			return ctrl.Result{}, err
		} else if err == nil {
			if err := r.Delete(ctx, &service); err != nil {
				log.Error(err, "Failed to delete service")
				return ctrl.Result{}, err
			}
		}
	}

	var currentDeploy appsv1.Deployment
	if err := r.Get(ctx, req.NamespacedName, &currentDeploy); err == nil {
		instance.Status.Replicas = currentDeploy.Status.ReadyReplicas

		instance.Status.DeploymentName = new(string)
		*instance.Status.DeploymentName = currentDeploy.Name

		// Update "Available" condition based on deployment status
		conditionStatus := metav1.ConditionFalse
		reason := "DeploymentProgressing"
		if currentDeploy.Status.ReadyReplicas > 0 {
			conditionStatus = metav1.ConditionTrue
			reason = "DeploymentReady"
		}

		meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
			Type:               "Available",
			Status:             conditionStatus,
			Reason:             reason,
			Message:            fmt.Sprintf("Deployment has %d ready replicas", currentDeploy.Status.ReadyReplicas),
			ObservedGeneration: instance.Generation,
		})
	}

	if err := r.Status().Update(ctx, &instance); err != nil {
		log.Error(err, "Failed to update Instance status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatusiov1alpha1.Instance{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Named("instance").
		Complete(r)
}

func getDeploymentLabels(instance *gatusiov1alpha1.Instance) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     "gatus",
		"app.kubernetes.io/instance": instance.Name,
		"managed-by":                 "gatus-operator",
	}
}

func MutateDeployment(instance *gatusiov1alpha1.Instance, obj *appsv1.Deployment, scheme *runtime.Scheme) error {
	// Check if the object already exists and has a different owner
	if obj.GetResourceVersion() != "" { // Object exists
		owner := metav1.GetControllerOf(obj)
		if owner == nil || owner.UID != instance.UID {
			return fmt.Errorf("deployment %s already exists and is not managed by this operator", obj.GetName())
		}
	}

	// Set/Ensure the Controller Reference
	if err := controllerutil.SetControllerReference(instance, obj, scheme); err != nil {
		return err
	}

	var deploymentLabels = getDeploymentLabels(instance)

	if instance.Spec.Image == nil {
		return fmt.Errorf("Image is empty, this means that the defaulting webhook is not working properly.")
	}
	image := *instance.Spec.Image

	obj.Spec.Replicas = instance.Spec.Replicas

	obj.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: maps.Clone(deploymentLabels),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "gatus",
					Image: image,
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: 8080,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					SecurityContext: &corev1.SecurityContext{
						ReadOnlyRootFilesystem: ptr.To(true),
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{
							Path: "/health",
							Port: intstr.FromString("http"),
						}},
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{
							Path: "/health",
							Port: intstr.FromString("http"),
						}},
					},
				},
			},
			SecurityContext: &corev1.PodSecurityContext{
				RunAsNonRoot: ptr.To(true),
				RunAsUser:    ptr.To[int64](65534),
				RunAsGroup:   ptr.To[int64](65534),
				FSGroup:      ptr.To[int64](65534),
			},
		},
	}

	// We only set the selector IF the object is being created.
	// If it exists, we leave the selector alone to avoid immutability errors.
	if obj.ObjectMeta.CreationTimestamp.IsZero() {
		obj.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: maps.Clone(deploymentLabels),
		}
	}

	return nil
}

func MutateService(instance *gatusiov1alpha1.Instance, obj *corev1.Service, scheme *runtime.Scheme) error {
	// Check if the object already exists and has a different owner
	if obj.GetResourceVersion() != "" { // Object exists
		owner := metav1.GetControllerOf(obj)
		if owner == nil || owner.UID != instance.UID {
			return fmt.Errorf("deployment %s already exists and is not managed by this operator", obj.GetName())
		}
	}

	// Set/Ensure the Controller Reference
	if err := controllerutil.SetControllerReference(instance, obj, scheme); err != nil {
		return err
	}

	var deploymentLabels = getDeploymentLabels(instance)

	var serviceLabels = map[string]string{}
	maps.Copy(serviceLabels, instance.Spec.Service.ServiceLabels)
	maps.Copy(serviceLabels, deploymentLabels)

	obj.Annotations = maps.Clone(instance.Spec.Service.ServiceAnnotations)
	obj.Labels = maps.Clone(serviceLabels)

	obj.Spec = corev1.ServiceSpec{
		Selector: deploymentLabels,
		Ports: []corev1.ServicePort{
			{
				Name:       "http",
				Protocol:   corev1.ProtocolTCP,
				Port:       80,
				TargetPort: intstr.FromString("http"),
			},
		},
		Type:           instance.Spec.Service.ServiceType,
		IPFamilyPolicy: instance.Spec.Service.IPFamilyPolicy,
		IPFamilies:     instance.Spec.Service.IPFamilies,
	}

	return nil
}
