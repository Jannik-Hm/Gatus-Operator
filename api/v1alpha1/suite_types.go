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
	"maps"

	gatusconfig "github.com/Jannik-Hm/Gatus-Operator/internal/gatus_config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SuiteSpec defines the desired state of Suite
type SuiteSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	InstanceReference `json:",inline"`

	// +required
	Config SuiteConfigSpec `json:"config"`
}

type SuiteConfigSpec struct {
	// Whether to monitor the suite.
	// Defaults to true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Name of the suite.
	// Must be unique.
	// +required
	Name string `json:"name"`

	// Group name.
	// Used to group multiple suites together on the dashboard.
	// +optional
	Group *string `json:"group,omitempty"`

	// Duration to wait between suite executions.
	// +optional
	Interval *string `json:"interval,omitempty"`

	// Maximum duration for the entire suite execution.
	// +optional
	Timeout *string `json:"timeout,omitempty"`

	// Initial context values that can be referenced by endpoints. (see https://github.com/TwiN/gatus/tree/master#using-context-in-endpoints)
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	Context map[string]runtime.RawExtension `json:"context,omitempty"`

	// List of endpoints to execute sequentially.
	// +required
	Endpoints []*SuiteEndpointSpec `json:"endpoints,omitempty"`
}

func (spec *SuiteConfigSpec) ToGatusConfig() *gatusconfig.GatusSuiteConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusSuiteConfig{
		Enabled:   spec.Enabled,
		Name:      spec.Name,
		Group:     spec.Group,
		Interval:  spec.Interval,
		Timeout:   spec.Timeout,
		Context:   maps.Clone(spec.Context),
		Endpoints: ToGatusConfigList(spec.Endpoints),
	}
}

type SuiteEndpointSpec struct {
	EndpointConfigSpec `json:",inline"`

	// Suites only

	// Whether to execute this endpoint even if previous endpoints in the suite failed.
	// +optional
	AlwaysRun *bool `json:"always-run,omitempty"`

	// Map of values to extract from the response and store in the suite context (stored even on failure).
	// +optional
	Store map[string]string `json:"store,omitempty"`
}

func (spec *SuiteEndpointSpec) ToGatusConfig() *gatusconfig.GatusEndpointConfig {
	if spec == nil {
		return nil
	}
	result := spec.EndpointConfigSpec.ToGatusConfig()

	result.AlwaysRun = spec.AlwaysRun
	result.Store = maps.Clone(result.Store)

	return result
}

// SuiteStatus defines the observed state of Suite.
type SuiteStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the Suite resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Suite is the Schema for the suites API
type Suite struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Suite
	// +required
	Spec SuiteSpec `json:"spec"`

	// status defines the observed state of Suite
	// +optional
	Status SuiteStatus `json:"status,omitzero"`
}

func (crd *Suite) GetInstances() []client.ObjectKey {
	instances := make([]client.ObjectKey, len(crd.Spec.Instances))

	for index, instance := range crd.Spec.Instances {
		namespace := crd.GetNamespace()
		if instance.Namespace != nil {
			namespace = *instance.Namespace
		}

		instances[index] = types.NamespacedName{
			Name:      instance.Name,
			Namespace: namespace,
		}
	}
	return instances
}

// +kubebuilder:object:root=true

// SuiteList contains a list of Suite
type SuiteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Suite `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Suite{}, &SuiteList{})
}
