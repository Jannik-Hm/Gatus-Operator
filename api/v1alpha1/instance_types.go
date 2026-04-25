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
	"slices"

	gatusconfig "github.com/Jannik-Hm/Gatus-Operator/internal/gatus_config"
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

	// Name of the Service created, defaults to instance name
	// +optional
	ServiceName *string `json:"serviceName,omitempty"`

	// Number of Gatus Instance Replicas, defaults to 1
	// Currently limiting to 1 replica max as gatus currently does not support HA
	// +kubebuilder:default:=1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1
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
	GatusConfig GatusInstanceSpec `json:"gatus-config"`

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

	// Hash of current configmap
	// +optional
	CurrentConfigmapHash string `json:"currentConfigmapHash,omitempty"`

	// Hash of last successful configmap
	// +optional
	LastSuccessfulConfigmapHash string `json:"lastSuccessfulConfigmapHash,omitempty"`

	// Current status
	// +kubebuilder:validation:Enum:=Pending;Running;Failed
	// +kubebuilder:default:=Pending
	Status string `json:"status"`
}

type ServiceConfig struct {
	// Enable service generation (only required when all other service settings are omitted)
	// +kubebuilder:default:=true
	// +optional
	Enabled bool `json:"enable"`

	// Type of the service
	// +kubebuilder:default:=ClusterIP
	// +optional
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

type GatusInstanceSpec struct {
	// Whether to expose metrics at /metrics.
	// +optional
	Metrics *bool `json:"metrics,omitempty"`

	// Storage Configuration (see https://github.com/TwiN/gatus/tree/master#storage)
	// default is Memory, this is not container restart persistent
	// since config changes result in a rolling deployment change, that will clear all history
	// to persist either use sqlite or postgres
	// +optional
	Storage *GatusStorageSpec `json:"storage,omitempty"` // TODO: custom struct to create VPC when type is `sqlite`

	// Alerting Configuration (see https://github.com/TwiN/gatus/tree/master#alerting)
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	Alerting *runtime.RawExtension `json:"alerting,omitempty"`

	// Security Configuration (see https://github.com/TwiN/gatus/tree/master#security)
	// Note: there will be changes to simplify secret refs
	// +optional
	Security *GatusSecuritySpec `json:"security,omitempty"` // TODO: custom struct for CRD to allow secretRefs and do env substitution

	// Maximum number of endpoints/suites to monitor concurrently
	// Set to 0 for unlimited (see https://github.com/TwiN/gatus/tree/master#concurrency)
	Concurrency *int32 `json:"concurrency,omitempty"`

	// Web Configuration
	// If you experience `431 Request Header Fields Too Large error`, increase this value (default is 8192, as of 2026-04-10)
	// +optional
	Web *GatusInstanceWebSpec `json:"web,omitempty"`

	// UI Configuration (see https://github.com/TwiN/gatus/tree/master#ui)
	// +optional
	Ui *GatusUiSpec `json:"ui,omitempty"`

	// Maintenance Configuration (see https://github.com/TwiN/gatus/tree/master#maintenance)
	// +optional
	Maintenance *GatusMaintenanceSpec `json:"maintenance,omitempty"`

	// Connectivity Configuration (see https://github.com/TwiN/gatus/tree/master#connectivity)
	// Used to check wether the connection of Gatus itself is broken
	// All endpoint executions are skipped while the connectivity checker deems connectivity to be down
	// +optional
	Connectivity *GatusConnectivitySpec `json:"connectivity,omitempty"`
}

type GatusStorageSpec struct {
	// +kubebuilder:validation:Enum:=memory;sqlite;postgres
	// +optional
	Type              *gatusconfig.GatusStorageType `json:"type,omitempty"`
	Path              *string                       `json:"path,omitempty"`
	Caching           *bool                         `json:"caching,omitempty"`
	MaximumResultsNum *int32                        `json:"maximum-number-of-results,omitempty"`
	MaximumEventsNum  *int32                        `json:"maximum-number-of-events,omitempty"`
}

func (spec *GatusStorageSpec) ToGatusConfig() *gatusconfig.GatusStorageConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusStorageConfig{
		Type:              spec.Type,
		Path:              spec.Path,
		Caching:           spec.Caching,
		MaximumResultsNum: spec.MaximumResultsNum,
		MaximumEventsNum:  spec.MaximumEventsNum,
	}
}

type GatusSecuritySpec struct {
	Basic *GatusSecurityBasicSpec `json:"basic,omitempty"`
	OIDC  *GatusSecurityOIDCSpec  `json:"oidc,omitempty"`
}

func (spec *GatusSecuritySpec) ToGatusConfig() *gatusconfig.GatusSecurityConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusSecurityConfig{
		Basic: spec.Basic.ToGatusConfig(),
		OIDC:  spec.OIDC.ToGatusConfig(),
	}
}

type GatusSecurityBasicSpec struct {
	Username string `json:"username"`               //TODO: secret refs
	PassHash string `json:"password-bcrypt-base64"` //TODO: secret refs
}

func (spec *GatusSecurityBasicSpec) ToGatusConfig() *gatusconfig.GatusSecurityBasicConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusSecurityBasicConfig{
		Username: spec.Username,
		PassHash: spec.PassHash,
	}
}

type GatusSecurityOIDCSpec struct {
	IssuerURL       string   `json:"issuer-url"`                 //TODO: secret refs
	RedirectURL     string   `json:"redirect-url"`               //TODO: secret refs
	ClientID        string   `json:"client-id"`                  //TODO: secret refs
	ClientSecret    string   `json:"client-secret"`              //TODO: secret refs
	Scopes          []string `json:"scopes,omitempty"`           //TODO: secret refs
	AllowedSubjects []string `json:"allowed-subjects,omitempty"` //TODO: secret refs
	SessionTTL      *string  `json:"session-ttl,omitempty"`
}

func (spec *GatusSecurityOIDCSpec) ToGatusConfig() *gatusconfig.GatusSecurityOIDCConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusSecurityOIDCConfig{
		IssuerURL:       spec.IssuerURL,
		RedirectURL:     spec.RedirectURL,
		ClientID:        spec.ClientID,
		ClientSecret:    spec.ClientSecret,
		Scopes:          slices.Clone(spec.Scopes),
		AllowedSubjects: slices.Clone(spec.AllowedSubjects),
		SessionTTL:      spec.SessionTTL,
	}
}

type GatusInstanceWebSpec struct {
	ReadBufferSize *int32 `json:"read-buffer-size,omitempty"`

	// settings such as port or listen address should not be user adjustable
}

func (spec *GatusInstanceWebSpec) ToGatusConfig() *gatusconfig.GatusWebConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusWebConfig{
		ReadBufferSize: spec.ReadBufferSize,
	}
}

type GatusUiSpec struct {
	Title               *string              `json:"title,omitempty"`
	Description         *string              `json:"description,omitempty"`
	DashboardHeading    *string              `json:"dashboard-heading,omitempty"`
	DashboardSubheading *string              `json:"dashboard-subheading,omitempty"`
	Header              *string              `json:"header,omitempty"`
	Logo                *string              `json:"logo,omitempty"`
	Link                *string              `json:"link,omitempty"`
	Favicon             *GatusUiFaviconSpec  `json:"favicon,omitempty"`
	Buttons             []*GatusUiButtonSpec `json:"buttons,omitempty"`
	CustomCSS           *string              `json:"custom-css,omitempty"`
	Darkmode            *bool                `json:"dark-mode,omitempty"`
	DefaultSortBy       *string              `json:"default-sort-by,omitempty"`
	DefaultFilterBy     *string              `json:"default-filter-by,omitempty"`
	LoginSubtitle       *string              `json:"login-subtitle,omitempty"`
}

func (spec *GatusUiSpec) ToGatusConfig() *gatusconfig.GatusUiConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusUiConfig{
		Title:               spec.Title,
		Description:         spec.Description,
		DashboardHeading:    spec.DashboardHeading,
		DashboardSubheading: spec.DashboardSubheading,
		Header:              spec.Header,
		Logo:                spec.Logo,
		Link:                spec.Link,
		Favicon:             spec.Favicon.ToGatusConfig(),
		Buttons:             ToGatusConfigList(spec.Buttons),
		CustomCSS:           spec.CustomCSS,
		Darkmode:            spec.Darkmode,
		DefaultSortBy:       spec.DefaultSortBy,
		DefaultFilterBy:     spec.DefaultFilterBy,
		LoginSubtitle:       spec.LoginSubtitle,
	}
}

type GatusUiFaviconSpec struct {
	Default   *string `json:"default,omitempty"`
	Size16x16 *string `json:"size16x16,omitempty"`
	Size32x32 *string `json:"size32x32,omitempty"`
}

func (spec *GatusUiFaviconSpec) ToGatusConfig() *gatusconfig.GatusUiFaviconConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusUiFaviconConfig{
		Default:   spec.Default,
		Size16x16: spec.Size16x16,
		Size32x32: spec.Size32x32,
	}
}

type GatusUiButtonSpec struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

func (spec *GatusUiButtonSpec) ToGatusConfig() *gatusconfig.GatusUiButtonConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusUiButtonConfig{
		Name: spec.Name,
		Link: spec.Link,
	}
}

type GatusMaintenanceSpec struct {
	Enabled  *bool    `json:"enabled,omitempty"`
	Start    string   `json:"start"`
	Duration string   `json:"duration"`
	Timezone *string  `json:"timezone,omitempty"`
	Every    []string `json:"every,omitempty"`
}

func (spec *GatusMaintenanceSpec) ToGatusConfig() *gatusconfig.GatusMaintenanceConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusMaintenanceConfig{
		Enabled:  spec.Enabled,
		Start:    spec.Start,
		Duration: spec.Duration,
		Timezone: spec.Timezone,
		Every:    spec.Every,
	}
}

type GatusConnectivitySpec struct {
	Checker GatusConnectivityCheckerSpec `json:"checker"`
}

func (spec *GatusConnectivitySpec) ToGatusConfig() *gatusconfig.GatusConnectivityConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusConnectivityConfig{
		Checker: *spec.Checker.ToGatusConfig(),
	}
}

type GatusConnectivityCheckerSpec struct {
	Target   string  `json:"target"`
	Interval *string `json:"interval,omitempty"`
}

func (spec *GatusConnectivityCheckerSpec) ToGatusConfig() *gatusconfig.GatusConnectivityCheckerConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusConnectivityCheckerConfig{
		Target:   spec.Target,
		Interval: spec.Interval,
	}
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status",description="The current status"
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".status.replicas",description="Number of currently available Replicas"

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
