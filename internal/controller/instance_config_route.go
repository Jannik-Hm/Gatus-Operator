package controller

import (
	"context"
	"fmt"

	annotatedressources "github.com/Jannik-Hm/Gatus-Operator/internal/annotated_ressources"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func (r *InstanceReconciler) getAnnotatedHTTPRoutes(ctx context.Context, req ctrl.Request) (annotatedressources.HttpRouteMap, error) {
	http_routes := annotatedressources.HttpRouteMap{}

	if r.Config.WatchHTTPRoutes {
		var annotated_http_routes gatewayv1.HTTPRouteList
		if err := r.List(ctx, &annotated_http_routes,
			client.MatchingFields{
				instancesAnnotationWithPrefix: req.Namespace + "/" + req.Name,
				disabledAnnotationWithPrefix:  "false",
			},
		); err != nil {
			return nil, fmt.Errorf("unable to list directly annotated HTTProutes: %w", err)
		}

		for _, route := range annotated_http_routes.Items {
			http_routes.AddUnique(&route, nil)
		}
	}

	if r.Config.WatchGateways {
		var annotated_gateways gatewayv1.GatewayList
		if err := r.List(ctx, &annotated_gateways,
			client.MatchingFields{
				instancesAnnotationWithPrefix: req.Namespace + "/" + req.Name,
				disabledAnnotationWithPrefix:  "false",
			},
		); err != nil {
			return nil, fmt.Errorf("unable to list directly annotated Gateways: %w", err)
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
				return nil, fmt.Errorf("unable to list indirectly annotated HTTProutes: %w", err)
			}
			for _, route := range attached_routes.Items {
				http_routes.AddUnique(&route, []*gatewayv1.Gateway{&gateway})
			}
		}
	}

	return http_routes, nil
}
