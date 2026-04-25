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
	gatusconfig "github.com/Jannik-Hm/Gatus-Operator/internal/gatus_config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AnnouncementSpec defines the desired state of Announcement
type AnnouncementSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	InstanceReference `json:",inline"`

	// +required
	Config AnnouncementConfigSpec `json:"config"`
}

type AnnouncementConfigSpec struct {
	// UTC timestamp when the announcement was made
	// +required
	Timestamp metav1.Time `json:"timestamp"`

	// Type of announcement, defaults to none
	// +kubebuilder:validation:Enum:=outage;warning;information;operational;none
	// +kubebuilder:default:=none
	// +optional
	Type string `json:"type,omitempty"`

	// The message to display to users
	// +required
	Message string `json:"message"`

	// Whether to archive the announcement.
	// Archived announcements show at the bottom of the status page instead of at the top.
	// +optional
	Archived *bool `json:"archived,omitempty"`
}

func (spec *AnnouncementConfigSpec) ToGatusConfig() *gatusconfig.GatusAnnouncementConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusAnnouncementConfig{
		Timestamp: spec.Timestamp,
		Type:      gatusconfig.AnnouncementType(spec.Type),
		Message:   spec.Message,
		Archived:  spec.Archived,
	}
}

// AnnouncementStatus defines the observed state of Announcement.
type AnnouncementStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the Announcement resource.
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

// Announcement is the Schema for the announcements API
type Announcement struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Announcement
	// +required
	Spec AnnouncementSpec `json:"spec"`

	// status defines the observed state of Announcement
	// +optional
	Status AnnouncementStatus `json:"status,omitzero"`
}

func (crd *Announcement) GetInstances() []client.ObjectKey {
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

// AnnouncementList contains a list of Announcement
type AnnouncementList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Announcement `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Announcement{}, &AnnouncementList{})
}
