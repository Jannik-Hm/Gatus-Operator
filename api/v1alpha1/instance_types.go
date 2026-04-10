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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InstanceSpec defines the desired state of Instance
type InstanceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// Name of the Gatus Instance (should be unique) and is used for reference when e.g. creating announcements, monitors, etc.
	// +required
	Name string `json:"name"`

	// Name of the Service created, defaults to instance name
	// +optional
	ServiceName *string `json:"serviceName,omitempty"`

	// Number of Gatus Instance Replicas, defaults to 1
	// +kubebuilder:default:=1
	// +kubebuilder:validation:Minimum=0
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Image to be used
	// +kubebuilder:validation:Pattern=`^([\w\.\-\/]+)(:[\w\.\-]+)?(@sha256:[a-fA-F0-0]{64})?$`
	// +optional
	Image *string `json:"image,omitempty"`

	// Service Config, if omitted, no service will be created
	// +optional
	Service *ServiceConfig `json:"service,omitempty"`

	// Global Gatus config
	// +optional
	GatusConfig GatusInstanceConfig `json:"gatus-config"`

	// TODO: ingress?

	// TODO: HTTPRoute?
}

// InstanceStatus defines the observed state of Instance.
type InstanceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the Instance resource.
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

	// Corresponding Deployment Name
	// +optional
	DeploymentName *string `json:"deploymentName,omitempty"`

	// Current Replicas
	// +kubebuilder:default:=0
	Replicas int32 `json:"replicas,omitempty"`
}

type ServiceConfig struct {
	// Enable service generation (only required when all other service settings are omitted)
	// +kubebuilder:default:=true
	Enabled bool `json:"enable"`

	// Type of the service
	// +kubebuilder:default:=ClusterIP
	ServiceType corev1.ServiceType `json:"type"`

	// Additional Service Annotations
	// +optional
	ServiceAnnotations map[string]string `json:"annotations,omitempty"`

	// Additional Service Labels
	// +optional
	ServiceLabels map[string]string `json:"labels,omitempty"`

	// IPFamilyPolicy of the service
	// +optional
	IPFamilyPolicy *corev1.IPFamilyPolicy `json:"ipFamilyPolicy,omitempty"`

	// IP Families of the service
	// +optional
	IPFamilies []corev1.IPFamily `json:"ipFamilies,omitempty"`
}

type GatusInstanceConfig struct {
	// Whether to expose metrics at /metrics.
	// +optional
	Metrics *bool `json:"metrics,omitempty"`

	// Storage Configuration (see https://github.com/TwiN/gatus/tree/master#storage)
	// default is Memory, this is not container restart persistent
	// since config changes result in a rolling deployment change, that will clear all history
	// to persist either use sqlite or postgres
	// +optional
	Storage *GatusStorageConfig `json:"storage,omitempty"` // TODO: custom struct to create VPC when type is `sqlite`

	// Alerting Configuration (see https://github.com/TwiN/gatus/tree/master#alerting)
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	Alerting *runtime.RawExtension `json:"alerting,omitempty"`

	// Security Configuration (see https://github.com/TwiN/gatus/tree/master#security)
	// Note: there will be changes to simplify secret refs
	// +optional
	Security *GatusSecurityConfig `json:"security,omitempty"` // TODO: custom struct for CRD to allow secretRefs and do env substitution

	// Maximum number of endpoints/suites to monitor concurrently
	// Set to 0 for unlimited (see https://github.com/TwiN/gatus/tree/master#concurrency)
	Concurrency *int32 `json:"concurrency,omitempty"`

	// Web Configuration
	// If you experience `431 Request Header Fields Too Large error`, increase this value (default is 8192, as of 2026-04-10)
	// +optional
	Web *GatusInstanceWebConfig `json:"web,omitempty"`

	// UI Configuration (see https://github.com/TwiN/gatus/tree/master#ui)
	// +optional
	Ui *GatusUiConfig `json:"ui,omitempty"`

	// Maintenance Configuration (see https://github.com/TwiN/gatus/tree/master#maintenance)
	// +optional
	Maintenance *GatusMaintenanceConfig `json:"maintenance,omitempty"`

	// Connectivity Configuration (see https://github.com/TwiN/gatus/tree/master#connectivity)
	// Used to check wether the connection of Gatus itself is broken
	// All endpoint executions are skipped while the connectivity checker deems connectivity to be down
	// +optional
	Connectivity *GatusConnectivityConfig `json:"connectivity,omitempty"`
}

type GatusInstanceWebConfig struct {
	ReadBufferSize *int32 `json:"read-buffer-size,omitempty"`

	// settings such as port or listen address should not be user adjustable
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Instance is the Schema for the instances API
type Instance struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Instance
	// +required
	Spec InstanceSpec `json:"spec"`

	// status defines the observed state of Instance
	// +optional
	Status InstanceStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// InstanceList contains a list of Instance
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Instance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Instance{}, &InstanceList{})
}
