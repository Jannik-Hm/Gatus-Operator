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

package v1alpha1

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	gatusiov1alpha1 "github.com/Jannik-Hm/Gatus-Operator/api/v1alpha1"
	"github.com/Jannik-Hm/Gatus-Operator/internal/config"
)

// nolint:unused
// log is for logging in this package.
var instancelog = logf.Log.WithName("instance-resource")

// SetupInstanceWebhookWithManager registers the webhook for Instance in the manager.
func SetupInstanceWebhookWithManager(mgr ctrl.Manager, cfg *config.Config) error {
	return ctrl.NewWebhookManagedBy(mgr, &gatusiov1alpha1.Instance{}).
		WithDefaulter(&InstanceCustomDefaulter{
			DefaultImage: cfg.DefaultGatusImage,
		}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-gatus-io-v1alpha1-instance,mutating=true,failurePolicy=fail,sideEffects=None,groups=gatus.io,resources=instances,verbs=create;update,versions=v1alpha1,name=minstance-v1alpha1.kb.io,admissionReviewVersions=v1

// InstanceCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Instance when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type InstanceCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
	DefaultImage string
}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Instance.
func (d *InstanceCustomDefaulter) Default(_ context.Context, obj *gatusiov1alpha1.Instance) error {
	instancelog.Info("Defaulting for Instance", "name", obj.GetName())

	if obj.Spec.Image == nil {
		obj.Spec.Image = new(string)
		*obj.Spec.Image = d.DefaultImage
	}

	return nil
}
