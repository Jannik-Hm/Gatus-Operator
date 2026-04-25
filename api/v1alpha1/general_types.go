package v1alpha1

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	InstanceReferenceKey string = ".spec.instance"
)

type InstanceReference struct {
	// The Gatus Instance that this Config should be applied to
	// +required
	Instances []InstanceRefFields `json:"instances"`
}

type InstanceRefFields struct {
	// Name of the Instance
	// +required
	Name string `json:"name"`

	// Namespace of the Instance
	// +optional
	Namespace *string `json:"namespace,omitempty"`
}

// +kubebuilder:object:generate=false
type InstanceReferencer interface {
	client.Object
	GetInstances() []client.ObjectKey
}

func MapRessourceToInstance(ctx context.Context, obj client.Object) []reconcile.Request {
	ref, ok := obj.(InstanceReferencer)
	if !ok {
		return nil
	}

	instances := ref.GetInstances()
	reconciles := make([]reconcile.Request, len(instances))
	for index, instance := range instances {
		reconciles[index] = reconcile.Request{
			NamespacedName: instance,
		}
	}
	return reconciles
}
