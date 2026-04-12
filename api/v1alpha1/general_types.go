package v1alpha1

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	InstanceNameReferenceKey      string = ".spec.instance.name"
	InstanceNamespaceReferenceKey string = ".spec.instance.namespace"
)

type InstanceReference struct {
	// The Gatus Instance that this Config should be applied to
	// +required
	Instance InstanceRefFields `json:"instance"`
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
	GetInstanceName() string
	GetInstanceNamespace() *string
}

func MapRessourceToInstance(ctx context.Context, obj client.Object) []reconcile.Request {
	ref, ok := obj.(InstanceReferencer)
	if !ok || ref.GetInstanceName() == "" {
		return nil
	}

	var namespace string
	tmp_ns := ref.GetInstanceNamespace()
	if tmp_ns != nil {
		namespace = *tmp_ns
	} else {
		namespace = obj.GetNamespace()
	}

	return []reconcile.Request{
		{
			NamespacedName: client.ObjectKey{
				Name:      ref.GetInstanceName(),
				Namespace: namespace,
			},
		},
	}
}
