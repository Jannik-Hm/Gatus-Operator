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
	"crypto/sha256"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

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
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=gatus.io,resources=endpoints,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gatus.io,resources=endpoints/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=gatus.io,resources=endpoints/finalizers,verbs=update

// +kubebuilder:rbac:groups=gatus.io,resources=externalendpoints,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gatus.io,resources=externalendpoints/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=gatus.io,resources=externalendpoints/finalizers,verbs=update

// +kubebuilder:rbac:groups=gatus.io,resources=announcements,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gatus.io,resources=announcements/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=gatus.io,resources=announcements/finalizers,verbs=update

// +kubebuilder:rbac:groups=gatus.io,resources=suites,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gatus.io,resources=suites/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=gatus.io,resources=suites/finalizers,verbs=update

// +kubebuilder:rbac:groups="gateway.networking.k8s.io",resources=httproutes;gateways,verbs=get;list;watch
// +kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses;ingressclasses,verbs=get;list;watch

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

	log.Info("Reconciling Instance", "Name", req.Name, "Namespace", req.Namespace)

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

	// clean up old configmaps
	var config_maps corev1.ConfigMapList
	if err := r.List(ctx, &config_maps, client.MatchingLabels(getInstanceLabels(&instance))); err != nil && !apierrors.IsNotFound(err) {
		log.Error(err, "Failed to list existing config maps")
		return ctrl.Result{}, err
	}
	if len(config_maps.Items) > 0 {
		for _, config_map := range config_maps.Items {
			if config_map.Annotations["config-hash"] != instance.Status.CurrentConfigmapHash && config_map.Annotations["config-hash"] != instance.Status.LastSuccessfulConfigmapHash {
				if err := r.Delete(ctx, &config_map); err != nil {
					log.Error(err, "Failed to remove old config map")
					return ctrl.Result{}, err
				}
			}
		}
	}

	// create/update config
	configYaml, err := r.generateConfigString(ctx, req, &instance)

	if err != nil {
		log.Error(err, "Failed to create Configmap")
		return ctrl.Result{}, err
	}

	hash := sha256.Sum256([]byte(configYaml))

	correspondingConfigmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-config-%.10x", instance.Name, hash),
			Namespace: req.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, correspondingConfigmap, func() error { return mutateConfig(&instance, correspondingConfigmap, r.Scheme, configYaml) })

	if err != nil {
		log.Error(err, "Failed to create/update Configmap")
		return ctrl.Result{}, err
	}

	if op != controllerutil.OperationResultNone {
		log.Info("Configmap reconciled", "Operation", op)
	}
	if op == controllerutil.OperationResultUpdated || op == controllerutil.OperationResultUpdatedStatus {
		log.Info("Configmap has been updated, that likely means that somebody else messed with the spec")
	}

	// create/update deployment
	correspondingDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
	}

	op, err = controllerutil.CreateOrUpdate(ctx, r.Client, correspondingDeployment, func() error {
		return mutateDeployment(&instance, correspondingDeployment, r.Scheme, correspondingConfigmap)
	})

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

		op, err = controllerutil.CreateOrUpdate(ctx, r.Client, correspondingService, func() error { return mutateService(&instance, correspondingService, r.Scheme) })

		if err != nil {
			log.Error(err, "Failed to create/update Service")
			return ctrl.Result{}, err
		}

		if op != controllerutil.OperationResultNone {
			log.Info("Service reconciled", "Operation", op)
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
	err = r.Get(ctx, req.NamespacedName, &instance)
	if err != nil {
		log.Error(err, "Failed to re-get Instance")
		return ctrl.Result{}, err
	}
	err = r.Get(ctx, req.NamespacedName, &currentDeploy)
	if err != nil {
		log.Error(err, "Failed to re-get deployment")
		return ctrl.Result{}, err
	}

	instance.Status.Replicas = currentDeploy.Status.ReadyReplicas

	instance.Status.DeploymentName = new(string)
	*instance.Status.DeploymentName = currentDeploy.Name

	instance.Status.CurrentConfigmapHash = fmt.Sprintf("%x", hash)

	// Update "Available" condition based on deployment status
	conditionStatus := metav1.ConditionFalse
	reason := "DeploymentProgressing"
	instanceStatus := "Pending"
	var configUpdateSucceeded *bool
	if currentDeploy.Status.UpdatedReplicas > 0 && currentDeploy.Status.UpdatedReplicas == currentDeploy.Status.AvailableReplicas {
		conditionStatus = metav1.ConditionTrue
		reason = "DeploymentReady"
		instance.Status.LastSuccessfulConfigmapHash = fmt.Sprintf("%x", hash)
		configUpdateSucceeded = ptr.To(true)
		instanceStatus = "Running"
	} else {
		for _, condition := range currentDeploy.Status.Conditions {
			if condition.Type == appsv1.DeploymentProgressing && condition.Status == corev1.ConditionFalse {
				if condition.Reason == "ProgressDeadlineExceeded" {
					log.Info("Deployment rollout failed: ProgressDeadlineExceeded")
					configUpdateSucceeded = ptr.To(false)
					reason = "DeploymentRolloutFailed"
					instanceStatus = "Failed"
				}
			}
		}
	}

	instance.Status.Status = instanceStatus

	if configUpdateSucceeded == nil || *configUpdateSucceeded {
		meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
			Type:               "Available",
			Status:             conditionStatus,
			Reason:             reason,
			Message:            fmt.Sprintf("Deployment has %d ready replicas", currentDeploy.Status.ReadyReplicas),
			ObservedGeneration: instance.Generation,
		})
	} else {
		meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
			Type:               "ConfigUpdateFailed",
			Status:             conditionStatus,
			Reason:             reason,
			Message:            "Deployment update failed, likely an invalid config",
			ObservedGeneration: instance.Generation,
		})
	}

	if err := r.Status().Update(ctx, &instance); err != nil {
		log.Error(err, "Failed to update Instance status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

const (
	ingressClassRef = "ingressClass"
)

// SetupWithManager sets up the controller with the Manager.
func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := r.registerInstanceRefIndices(mgr, &gatusiov1alpha1.Endpoint{}); err != nil {
		return fmt.Errorf("Could not register Endpoint Indices: %s", err)
	}

	if err := r.registerInstanceRefIndices(mgr, &gatusiov1alpha1.ExternalEndpoint{}); err != nil {
		return fmt.Errorf("Could not register ExternalEndpoint Indices: %s", err)
	}

	if err := r.registerInstanceRefIndices(mgr, &gatusiov1alpha1.Announcement{}); err != nil {
		return fmt.Errorf("Could not register Announcement Indices: %s", err)
	}

	if err := r.registerInstanceRefIndices(mgr, &gatusiov1alpha1.Suite{}); err != nil {
		return fmt.Errorf("Could not register Suite Indices: %s", err)
	}

	if r.Config.WatchIngresses {
		if err := r.registerAnnotationIndices(mgr, &networkingv1.Ingress{}); err != nil {
			return fmt.Errorf("Could not register Ingress Indices: %s", err)
		}
	}

	if r.Config.WatchIngressClasses {
		// NOTE: since IngressClasses do not have namespaces, instance namespace always needs to be specified in `instances` annotation
		if err := r.registerAnnotationIndices(mgr, &networkingv1.IngressClass{}); err != nil {
			return fmt.Errorf("Could not register IngressClass Indices: %s", err)
		}
	}

	if r.Config.WatchIngressClasses {
		// register ingress class ref
		if err := mgr.GetCache().IndexField(context.Background(), &networkingv1.Ingress{}, ingressClassRef, func(rawObj client.Object) []string {
			route := rawObj.(*networkingv1.Ingress)
			var ingressClasses []string
			if route.Spec.IngressClassName != nil {
				ingressClasses = append(ingressClasses, *route.Spec.IngressClassName)
			} else if _, ok := route.Annotations["kubernetes.io/ingress.class"]; ok {
				ingressClasses = append(ingressClasses, route.Annotations["kubernetes.io/ingress.class"])
			} else {
				// empty string as default
				ingressClasses = append(ingressClasses, "")
			}
			return ingressClasses
		}); err != nil {
			return fmt.Errorf("Could not register Ingress Parent Ref Indices: %s", err)
		}
	}

	if r.Config.WatchHTTPRoutes {
		if err := r.registerAnnotationIndices(mgr, &gatewayv1.HTTPRoute{}); err != nil {
			return fmt.Errorf("Could not register HTTPRoute Indices: %s", err)
		}
	}

	if r.Config.WatchGateways {
		if err := r.registerAnnotationIndices(mgr, &gatewayv1.Gateway{}); err != nil {
			return fmt.Errorf("Could not register Gateway Indices: %s", err)
		}
	}

	if r.Config.WatchGateways {
		// register gateway parent ref
		if err := mgr.GetCache().IndexField(context.Background(), &gatewayv1.HTTPRoute{}, gatewayParentRefSpec, func(rawObj client.Object) []string {
			route := rawObj.(*gatewayv1.HTTPRoute)
			var gatewayNames []string
			for _, ref := range route.Spec.ParentRefs {
				if ref.Kind != nil && *ref.Kind == "Gateway" {
					namespace := route.GetNamespace()
					if ref.Namespace != nil {
						namespace = string(*ref.Namespace)
					}
					gatewayNames = append(gatewayNames, fmt.Sprintf("%s/%s", namespace, ref.Name))
				}
			}
			return gatewayNames
		}); err != nil {
			return fmt.Errorf("Could not register Gateway Parent Ref Indices: %s", err)
		}
	}

	controller := ctrl.NewControllerManagedBy(mgr).
		For(&gatusiov1alpha1.Instance{}, builder.WithPredicates(predicate.GenerationChangedPredicate{}))

	// managed ressources
	controller.Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&corev1.ConfigMap{}, builder.WithPredicates(predicate.GenerationChangedPredicate{}))

	// CRDs
	controller.
		Watches(&gatusiov1alpha1.Endpoint{}, handler.EnqueueRequestsFromMapFunc(gatusiov1alpha1.MapRessourceToInstance), builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&gatusiov1alpha1.ExternalEndpoint{}, handler.EnqueueRequestsFromMapFunc(gatusiov1alpha1.MapRessourceToInstance), builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&gatusiov1alpha1.Announcement{}, handler.EnqueueRequestsFromMapFunc(gatusiov1alpha1.MapRessourceToInstance), builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&gatusiov1alpha1.Suite{}, handler.EnqueueRequestsFromMapFunc(gatusiov1alpha1.MapRessourceToInstance), builder.WithPredicates(predicate.GenerationChangedPredicate{}))

	// annotations
	if r.Config.WatchIngresses || r.Config.WatchIngressClasses {
		controller.Watches(&networkingv1.Ingress{},
			handler.EnqueueRequestsFromMapFunc(mapLabelsToInstances),
			builder.WithPredicates(predicate.Or(
				predicate.GenerationChangedPredicate{},
				predicate.AnnotationChangedPredicate{},
			)),
		)
	}
	if r.Config.WatchIngressClasses {
		controller.Watches(&networkingv1.IngressClass{},
			handler.EnqueueRequestsFromMapFunc(mapLabelsToInstances),
			builder.WithPredicates(predicate.Or(
				predicate.GenerationChangedPredicate{},
				predicate.AnnotationChangedPredicate{},
			)),
		)
	}
	if r.Config.WatchHTTPRoutes {
		controller.Watches(
			&gatewayv1.HTTPRoute{},
			handler.EnqueueRequestsFromMapFunc(mapLabelsToInstances),
			builder.WithPredicates(predicate.Or(
				predicate.GenerationChangedPredicate{},
				predicate.AnnotationChangedPredicate{},
			)),
		)
	}
	if r.Config.WatchGateways {
		controller.Watches(&gatewayv1.Gateway{},
			handler.EnqueueRequestsFromMapFunc(r.mapGatewayToInstances),
			builder.WithPredicates(predicate.Or(
				predicate.GenerationChangedPredicate{},
				predicate.AnnotationChangedPredicate{},
			)),
		)
	}

	return controller.Named("instance").
		Complete(r)
}

func (r *InstanceReconciler) registerInstanceRefIndices(mgr ctrl.Manager, obj gatusiov1alpha1.InstanceReferencer) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), obj, gatusiov1alpha1.InstanceReferenceKey, func(rawObj client.Object) []string {
		// grab the endpoint object, extract the instance...
		endpoint := rawObj.(gatusiov1alpha1.InstanceReferencer)
		instances := endpoint.GetInstances()
		indices := make([]string, len(instances))
		for index, instance := range instances {
			indices[index] = instance.Namespace + "/" + instance.Name
		}
		return indices
	}); err != nil {
		return err
	}
	return nil
}
