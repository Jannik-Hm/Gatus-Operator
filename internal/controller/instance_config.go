package controller

import (
	"context"
	"crypto/sha256"
	"fmt"
	"maps"
	"strings"

	annotatedressources "github.com/Jannik-Hm/Gatus-Operator/internal/annotated_ressources"
	gatusconfig "github.com/Jannik-Hm/Gatus-Operator/internal/gatus_config"
	"go.yaml.in/yaml/v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	gatusiov1alpha1 "github.com/Jannik-Hm/Gatus-Operator/api/v1alpha1"
)

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
		obj := annotatedressources.AnnotatedIngress{Ingress: item}
		cfgs, err := obj.GetEndpointConfigs(*r.Config)
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
		cfgs, err := route.GetEndpointConfigs(*r.Config)
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

	disabledAnnotationWithPrefix       string = annotationPrefix + annotatedressources.DisabledAnnotation
	nameAnnotationWithPrefix           string = annotationPrefix + annotatedressources.NameAnnotation
	instancesAnnotationWithPrefix      string = annotationPrefix + annotatedressources.InstancesAnnotation
	groupAnnotationWithPrefix          string = annotationPrefix + annotatedressources.GroupAnnotation
	configOverrideAnnotationWithPrefix string = annotationPrefix + annotatedressources.ConfigOverrideAnnotation

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
		disabled, ok := annotations[annotatedressources.DisabledAnnotation]

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
	instances, ok := annotations[annotatedressources.InstancesAnnotation]
	disabled := annotations[annotatedressources.DisabledAnnotation]
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
