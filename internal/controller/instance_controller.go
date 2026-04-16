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
	"maps"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/yaml"

	gatusiov1alpha1 "github.com/Jannik-Hm/Gatus-Operator/api/v1alpha1"
	annotatedressources "github.com/Jannik-Hm/Gatus-Operator/internal/annotated_ressources"
	"github.com/Jannik-Hm/Gatus-Operator/internal/config"
	gatusconfig "github.com/Jannik-Hm/Gatus-Operator/internal/gatus_config"
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
// +kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses,verbs=get;list;watch

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
						// TODO: rollback deployment
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
	}

	if err := r.Status().Update(ctx, &instance); err != nil {
		log.Error(err, "Failed to update Instance status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

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

	if err := r.registerAnnotationIndices(mgr, &networkingv1.Ingress{}); err != nil {
		return fmt.Errorf("Could not register Ingress Indices: %s", err)
	}

	if err := r.registerAnnotationIndices(mgr, &gatewayv1.HTTPRoute{}); err != nil {
		return fmt.Errorf("Could not register HTTPRoute Indices: %s", err)
	}

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
		return fmt.Errorf("Could not register Gateway Indices: %s", err)
	}

	controller := ctrl.NewControllerManagedBy(mgr).
		For(&gatusiov1alpha1.Instance{}, builder.WithPredicates(predicate.GenerationChangedPredicate{}))

	// managed ressources
	controller.Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{})

	// CRDs
	controller.
		Watches(&gatusiov1alpha1.Endpoint{}, handler.EnqueueRequestsFromMapFunc(gatusiov1alpha1.MapRessourceToInstance)).
		Watches(&gatusiov1alpha1.ExternalEndpoint{}, handler.EnqueueRequestsFromMapFunc(gatusiov1alpha1.MapRessourceToInstance)).
		Watches(&gatusiov1alpha1.Announcement{}, handler.EnqueueRequestsFromMapFunc(gatusiov1alpha1.MapRessourceToInstance)).
		Watches(&gatusiov1alpha1.Suite{}, handler.EnqueueRequestsFromMapFunc(gatusiov1alpha1.MapRessourceToInstance))

	// annotations
	controller.
		Watches(&networkingv1.Ingress{}, handler.EnqueueRequestsFromMapFunc(mapLabelsToInstances), builder.WithPredicates(predicate.GenerationChangedPredicate{})). // TODO: add flag for this watcher
		Watches(&gatewayv1.HTTPRoute{}, handler.EnqueueRequestsFromMapFunc(mapLabelsToInstances), builder.WithPredicates(predicate.GenerationChangedPredicate{})).  // TODO: add flag for this watcher
		Watches(&gatewayv1.Gateway{}, handler.EnqueueRequestsFromMapFunc(r.mapGatewayToInstances), builder.WithPredicates(predicate.GenerationChangedPredicate{}))  // TODO: add flag for this watcher
		// TODO: IngressClass?

	return controller.Named("instance").
		Complete(r)
}

func (r *InstanceReconciler) registerInstanceRefIndices(mgr ctrl.Manager, obj gatusiov1alpha1.InstanceReferencer) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), obj, gatusiov1alpha1.InstanceNameReferenceKey, func(rawObj client.Object) []string {
		// grab the endpoint object, extract the instance...
		endpoint := rawObj.(gatusiov1alpha1.InstanceReferencer)
		return []string{endpoint.GetInstanceName()}
	}); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), obj, gatusiov1alpha1.InstanceNamespaceReferenceKey, func(rawObj client.Object) []string {
		// grab the endpoint object, extract the instance...
		endpoint := rawObj.(gatusiov1alpha1.InstanceReferencer)
		if endpoint.GetInstanceNamespace() != nil {
			return []string{*endpoint.GetInstanceNamespace()}
		}
		// fallback to own namespace
		return []string{endpoint.GetNamespace()}
	}); err != nil {
		return err
	}
	return nil
}

func getInstanceLabels(instance *gatusiov1alpha1.Instance) map[string]string {
	return map[string]string{
		"app.kubernetes.io/instance": instance.Name,
		"managed-by":                 "gatus-operator",
	}
}

func getDeploymentLabels(instance *gatusiov1alpha1.Instance) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     "gatus",
		"app.kubernetes.io/instance": instance.Name,
		"managed-by":                 "gatus-operator",
	}
}

func mutateDeployment(instance *gatusiov1alpha1.Instance, obj *appsv1.Deployment, scheme *runtime.Scheme, configmap *corev1.ConfigMap) error {
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

	obj.Spec.ProgressDeadlineSeconds = ptr.To[int32](120) // rollout should fail after 120 seconds (gatus has quick startup and fails fast if cfg is wrong)
	// TODO: either fine tune this setting or make it available to configure via env var

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
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "gatus-config",
							MountPath: "/config",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "gatus-config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: configmap.Name,
							},
						},
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

func mutateService(instance *gatusiov1alpha1.Instance, obj *corev1.Service, scheme *runtime.Scheme) error {
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
	var instanceLabels = getInstanceLabels(instance)

	var serviceLabels = map[string]string{}
	maps.Copy(serviceLabels, instance.Spec.Service.ServiceLabels)
	maps.Copy(serviceLabels, instanceLabels)

	obj.Annotations = maps.Clone(instance.Spec.Service.ServiceAnnotations)
	obj.Labels = maps.Clone(serviceLabels)

	obj.Spec = corev1.ServiceSpec{
		Selector: maps.Clone(deploymentLabels),
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

func (r *InstanceReconciler) generateConfigString(ctx context.Context, req ctrl.Request, instance *gatusiov1alpha1.Instance) (string, error) {
	log := logf.FromContext(ctx)

	var endpoints gatusiov1alpha1.EndpointList
	if err := r.List(ctx, &endpoints,
		client.MatchingFields{
			gatusiov1alpha1.InstanceNameReferenceKey:      req.Name,
			gatusiov1alpha1.InstanceNamespaceReferenceKey: req.Namespace,
		},
	); err != nil {
		log.Error(err, "unable to list endpoints")
	}

	var externalEndpoints gatusiov1alpha1.ExternalEndpointList
	if err := r.List(ctx, &externalEndpoints,
		client.MatchingFields{
			gatusiov1alpha1.InstanceNameReferenceKey:      req.Name,
			gatusiov1alpha1.InstanceNamespaceReferenceKey: req.Namespace,
		},
	); err != nil {
		log.Error(err, "unable to list externalEndpoints")
	}

	var announcements gatusiov1alpha1.AnnouncementList
	if err := r.List(ctx, &announcements,
		client.MatchingFields{
			gatusiov1alpha1.InstanceNameReferenceKey:      req.Name,
			gatusiov1alpha1.InstanceNamespaceReferenceKey: req.Namespace,
		},
	); err != nil {
		log.Error(err, "unable to list announcements")
	}

	var suites gatusiov1alpha1.SuiteList
	if err := r.List(ctx, &suites,
		client.MatchingFields{
			gatusiov1alpha1.InstanceNameReferenceKey:      req.Name,
			gatusiov1alpha1.InstanceNamespaceReferenceKey: req.Namespace,
		},
	); err != nil {
		log.Error(err, "unable to list suits")
	}

	// TODO: add fetchers for ingressClass

	var annotated_ingresses networkingv1.IngressList
	if err := r.List(ctx, &annotated_ingresses,
		client.MatchingFields{
			instancesAnnotationWithPrefix: req.Namespace + "/" + req.Name,
			disabledAnnotationWithPrefix:  "false",
		},
	); err != nil {
		log.Error(err, "unable to list directly annotated Ingresses")
	}

	http_routes := annotatedressources.HttpRouteMap{}

	var annotated_http_routes gatewayv1.HTTPRouteList
	if err := r.List(ctx, &annotated_http_routes,
		client.MatchingFields{
			instancesAnnotationWithPrefix: req.Namespace + "/" + req.Name,
			disabledAnnotationWithPrefix:  "false",
		},
	); err != nil {
		log.Error(err, "unable to list directly annotated HTTProutes")
	}

	for _, route := range annotated_http_routes.Items {
		http_routes.AddUnique(&route, nil)
	}

	var annotated_gateways gatewayv1.GatewayList
	if err := r.List(ctx, &annotated_gateways,
		client.MatchingFields{
			instancesAnnotationWithPrefix: req.Namespace + "/" + req.Name,
			disabledAnnotationWithPrefix:  "false",
		},
	); err != nil {
		log.Error(err, "unable to list directly annotated Gateways")
	}

	// get routes attached to annotated gateways
	for _, gateway := range annotated_gateways.Items {
		var attached_routes gatewayv1.HTTPRouteList
		if err := r.List(ctx, &attached_routes,
			client.MatchingFields{
				gatewayParentRefSpec:         fmt.Sprintf("%s/%s", gateway.Namespace, gateway.Name),
				disabledAnnotationWithPrefix: "false",
			},
		); err != nil {
			log.Error(err, "unable to list indirectly annotated HTTProutes")
		}
		for _, route := range attached_routes.Items {
			http_routes.AddUnique(&route, []*gatewayv1.Gateway{&gateway})
		}
	}

	gatus_config := gatusconfig.Config{
		Metrics:      instance.Spec.GatusConfig.Metrics,
		Storage:      instance.Spec.GatusConfig.Storage.ToGatusConfig(),
		Alerting:     instance.Spec.GatusConfig.Alerting,
		Security:     instance.Spec.GatusConfig.Security.ToGatusConfig(),
		Concurrency:  instance.Spec.GatusConfig.Concurrency,
		Web:          instance.Spec.GatusConfig.Web.ToGatusConfig(),
		Ui:           instance.Spec.GatusConfig.Ui.ToGatusConfig(),
		Maintenance:  instance.Spec.GatusConfig.Maintenance.ToGatusConfig(),
		Connectivity: instance.Spec.GatusConfig.Connectivity.ToGatusConfig(),
	}

	// TODO: sort lists

	gatus_config.Endpoints = make([]gatusconfig.GatusEndpointConfig, 0, len(endpoints.Items)+len(annotated_ingresses.Items)+len(http_routes))
	for _, item := range endpoints.Items {
		gatus_config.Endpoints = append(gatus_config.Endpoints, *item.Spec.Config.ToGatusConfig())
	}
	for _, item := range annotated_ingresses.Items {
		obj := annotatedressources.AnnotatedIngress(item)
		cfgs, err := annotationsToGatusConfigs(&obj)
		if err != nil {
			log.Error(err, "Could not parse Ingress annotations")
		}
		if cfgs == nil {
			continue
		}
		for _, cfg := range cfgs {
			gatus_config.Endpoints = append(gatus_config.Endpoints, *cfg)
		}
	}
	for _, route := range http_routes {
		err := route.GetMissingRouteGateways(ctx, r)
		if err != nil {
			log.Error(err, "Could not get parent Gateway spec")
		}
		cfgs, err := annotationsToGatusConfigs(route)
		if err != nil {
			log.Error(err, "Could not parse HTTPRoute annotations")
		}
		if cfgs == nil {
			continue
		}
		for _, cfg := range cfgs {
			gatus_config.Endpoints = append(gatus_config.Endpoints, *cfg)
		}
	}

	gatus_config.ExternalEndpoints = make([]gatusconfig.GatusExternalEndpointConfig, 0, len(externalEndpoints.Items))
	for _, item := range externalEndpoints.Items {
		gatus_config.ExternalEndpoints = append(gatus_config.ExternalEndpoints, *item.Spec.Config.ToGatusConfig())
	}

	gatus_config.Announcements = make([]gatusconfig.GatusAnnouncementConfig, 0, len(announcements.Items))
	for _, item := range announcements.Items {
		gatus_config.Announcements = append(gatus_config.Announcements, *item.Spec.Config.ToGatusConfig())
	}

	gatus_config.Suites = make([]gatusconfig.GatusSuiteConfig, 0, len(suites.Items))
	for _, item := range suites.Items {
		gatus_config.Suites = append(gatus_config.Suites, *item.Spec.Config.ToGatusConfig())
	}

	yaml, err := yaml.Marshal(gatus_config)

	if err != nil {
		return "", fmt.Errorf("Error converting gatus config into yaml: %s", err)
	}
	return string(yaml), nil
}

func mutateConfig(instance *gatusiov1alpha1.Instance, obj *corev1.ConfigMap, scheme *runtime.Scheme, configYaml string) error {
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

	var instanceLabels = getInstanceLabels(instance)

	obj.Labels = maps.Clone(instanceLabels)

	hash := sha256.Sum256([]byte(configYaml))

	obj.Annotations = map[string]string{
		"config-hash": fmt.Sprintf("%x", hash),
	}

	obj.Data = map[string]string{
		"config.yaml": configYaml,
	}

	return nil
}

const (
	annotationPrefix string = ".metadata.annotations."

	disabledAnnotation       string = "gatus.io/disabled"
	nameAnnotation           string = "gatus.io/name"
	instancesAnnotation      string = "gatus.io/instances"
	groupAnnotation          string = "gatus.io/group"
	configOverrideAnnotation string = "gatus.io/config"

	disabledAnnotationWithPrefix       string = annotationPrefix + disabledAnnotation
	nameAnnotationWithPrefix           string = annotationPrefix + nameAnnotation
	instancesAnnotationWithPrefix      string = annotationPrefix + instancesAnnotation
	groupAnnotationWithPrefix          string = annotationPrefix + groupAnnotation
	configOverrideAnnotationWithPrefix string = annotationPrefix + configOverrideAnnotation

	gatewayParentRefSpec string = "gatewayParentRef"
)

func (r *InstanceReconciler) registerAnnotationIndices(mgr ctrl.Manager, obj client.Object) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), obj, instancesAnnotationWithPrefix, func(rawObj client.Object) []string {
		// grab the object, extract the instances annotation...
		instance_keys := parseInstancesAnnotation(rawObj)
		indices := make([]string, 0, len(instance_keys))
		for _, value := range instance_keys {
			indices = append(indices, fmt.Sprintf("%s/%s", value.Namespace, value.Name))
		}
		return indices
	}); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), obj, disabledAnnotationWithPrefix, func(rawObj client.Object) []string {
		annotations := rawObj.GetAnnotations()
		disabled, ok := annotations[disabledAnnotation]

		disabled = strings.ToLower(disabled)
		if !ok || disabled != "true" {
			return []string{"false"} // default to false
		}
		return []string{disabled}
	}); err != nil {
		return err
	}
	return nil
}

func parseInstancesAnnotation(obj client.Object) []client.ObjectKey {
	annotations := obj.GetAnnotations()
	instances, ok := annotations[instancesAnnotation]
	disabled := annotations[disabledAnnotation]
	if !ok || instances == "" || disabled == "true" {
		return nil
	}

	objectKeys := make([]client.ObjectKey, 0)

	for _, value := range strings.Split(instances, ",") {
		value = strings.TrimSpace(value)

		if value == "" {
			continue
		}

		parts := strings.Split(value, "/")
		key := client.ObjectKey{}

		switch len(parts) {
		case 1:
			key.Namespace = obj.GetNamespace()
			key.Name = parts[0]
		case 2:
			key.Namespace = parts[0]
			key.Name = parts[1]
		default:
			// skip invalid format, maybe log this in the future?
			continue
		}
		objectKeys = append(objectKeys, key)
	}

	return objectKeys
}

func mapLabelsToInstances(ctx context.Context, obj client.Object) []reconcile.Request {
	if obj == nil {
		return nil
	}

	instance_keys := parseInstancesAnnotation(obj)
	if len(instance_keys) == 0 {
		return nil
	}

	var reconcile_requests []reconcile.Request = make([]reconcile.Request, 0, len(instance_keys))

	for _, instance_key := range instance_keys {
		reconcile_requests = append(reconcile_requests, reconcile.Request{
			NamespacedName: instance_key,
		})
	}

	return reconcile_requests
}

func annotationsToGatusConfigs(obj annotatedressources.AnnotatedRessource) ([]*gatusconfig.GatusEndpointConfig, error) {
	annotations := obj.GetAnnotations()
	var base_cfg gatusconfig.GatusEndpointConfig
	if additionalConfigString, ok := annotations[configOverrideAnnotation]; ok {
		err := yaml.Unmarshal([]byte(additionalConfigString), &base_cfg)
		if err != nil {
			return nil, fmt.Errorf("Could not parse override config annotation: %s", err)
		}
	} else {
		base_cfg = gatusconfig.GatusEndpointConfig{}
	}

	urls, err := obj.GetURLs()
	if err != nil {
		return nil, fmt.Errorf("Could not generate URLs: %s", err)
	}
	configs := make([]*gatusconfig.GatusEndpointConfig, 0, len(urls))

	for _, url := range urls {
		cfg := base_cfg

		cfg.URL = url

		if len(cfg.Conditions) == 0 {
			protocol, _, _ := strings.Cut(url, "://")
			cfg.Conditions = obj.GetConditions(protocol)
		}

		if name, ok := annotations[nameAnnotation]; ok && name != "" {
			cfg.Name = name
		} else {
			cfg.Name = obj.GetName()
		}
		// TODO: append unique string to name if len(urls) > 1

		if group, ok := annotations[groupAnnotation]; ok {
			cfg.Group = &group
		} else {
			cfg.Group = ptr.To(obj.GetNamespace())
		}
		configs = append(configs, &cfg)
	}

	return configs, nil
}

func (r *InstanceReconciler) mapGatewayToInstances(ctx context.Context, obj client.Object) []reconcile.Request {
	gateway := obj.(*gatewayv1.Gateway)

	instanceSet := make(map[client.ObjectKey]struct{})

	// parse gateway annotations
	instances := parseInstancesAnnotation(gateway)
	for _, instance := range instances {
		instanceSet[instance] = struct{}{}
	}

	// find all routes referencing this Gateway
	routeList := &gatewayv1.HTTPRouteList{}
	err := r.List(ctx, routeList, client.MatchingFields{gatewayParentRefSpec: fmt.Sprintf("%s/%s", gateway.Namespace, gateway.Name), disabledAnnotationWithPrefix: "false"}) //TODO: annotation key const + index
	if err != nil {
		return nil
	}

	// parse annotations for each route
	for _, route := range routeList.Items {
		instances := parseInstancesAnnotation(&route)

		for _, instance := range instances {
			instanceSet[instance] = struct{}{}
		}
	}

	requests := make([]reconcile.Request, 0, len(instanceSet))
	for instance := range instanceSet {
		requests = append(requests, reconcile.Request{NamespacedName: instance})
	}

	return requests
}
