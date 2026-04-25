package controller

import (
	"context"
	"fmt"

	annotatedressources "github.com/Jannik-Hm/Gatus-Operator/internal/annotated_ressources"
	networkingv1 "k8s.io/api/networking/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *InstanceReconciler) getAnnotatedIngresses(ctx context.Context, req ctrl.Request) (annotatedressources.IngressMap, error) {
	ingresses := annotatedressources.IngressMap{}

	if r.Config.WatchIngresses {
		var annotated_ingresses networkingv1.IngressList
		if err := r.List(ctx, &annotated_ingresses,
			client.MatchingFields{
				instancesAnnotationWithPrefix: req.Namespace + "/" + req.Name,
				disabledAnnotationWithPrefix:  "false",
			},
		); err != nil {
			return nil, fmt.Errorf("unable to list directly annotated Ingresses: %w", err)
		}

		for _, ingress := range annotated_ingresses.Items {
			ingresses.AddUnique(&ingress, nil)
		}
	}

	if r.Config.WatchIngressClasses {
		var annotated_ingress_classes networkingv1.IngressClassList
		if err := r.List(ctx, &annotated_ingress_classes,
			client.MatchingFields{
				instancesAnnotationWithPrefix: req.Namespace + "/" + req.Name,
				disabledAnnotationWithPrefix:  "false",
			},
		); err != nil {
			return nil, fmt.Errorf("unable to list directly annotated IngressClasses: %w", err)
		}

		// Note: default ingress class is only respected if ingress controller adds ingress class field to ingress via admission webhook

		// get ingresses attached to annotated classes
		for _, ingress_class := range annotated_ingress_classes.Items {
			var attached_ingresses networkingv1.IngressList
			if err := r.List(ctx, &attached_ingresses,
				client.MatchingFields{
					ingressClassRef:              ingress_class.Name,
					disabledAnnotationWithPrefix: "false",
				},
			); err != nil {
				return nil, fmt.Errorf("unable to list indirectly annotated HTTProutes: %w", err)
			}
			for _, ingress := range attached_ingresses.Items {
				ingresses.AddUnique(&ingress, &ingress_class)
			}
		}
	}

	return ingresses, nil
}
