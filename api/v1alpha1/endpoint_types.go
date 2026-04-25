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
	"slices"

	gatusconfig "github.com/Jannik-Hm/Gatus-Operator/internal/gatus_config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EndpointSpec defines the desired state of Endpoint
type EndpointSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	InstanceReference `json:",inline"`

	// +required
	Config EndpointConfigSpec `json:"config"`
}

type EndpointConfigSpec struct {
	//+optional
	Enabled *bool `json:"enabled,omitempty"`

	//+required
	Name string `json:"name"`

	//+optional
	Group *string `json:"group,omitempty"`

	//+required
	URL string `json:"url"`

	//+optional
	Method *string `json:"method,omitempty"`

	//+required
	Conditions []string `json:"conditions"`

	//+optional
	Interval *string `json:"interval,omitempty"`

	//+optional
	GraphQL *bool `json:"graphql,omitempty"`

	//+optional
	Body *string `json:"body,omitempty"`

	//+optional
	Headers map[string]string `json:"headers,omitempty"`

	//+optional
	DNS *EndpointDNSSpec `json:"dns,omitempty"`

	//+optional
	SSH *EndpointSSHSpec `json:"ssh,omitempty"`

	//+optional
	Alerts []*EndpointAlertSpec `json:"alerts,omitempty"`

	//+optional
	MaintenanceWindows []*GatusMaintenanceSpec `json:"maintenance-windows,omitempty"`

	//+optional
	Client *EndpointClientSpec `json:"client,omitempty"`

	//+optional
	Ui *EndpointUiSpec `json:"ui,omitempty"`

	//+optional
	ExtraLabels map[string]string `json:"extra-labels,omitempty"`
}

func (spec *EndpointConfigSpec) ToGatusConfig() *gatusconfig.GatusEndpointConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusEndpointConfig{
		Enabled:            spec.Enabled,
		Name:               spec.Name,
		Group:              spec.Group,
		URL:                spec.URL,
		Method:             spec.Method,
		Conditions:         spec.Conditions,
		Interval:           spec.Interval,
		GraphQL:            spec.GraphQL,
		Body:               spec.Body,
		Headers:            maps.Clone(spec.Headers),
		DNS:                spec.DNS.ToGatusConfig(),
		SSH:                spec.SSH.ToGatusConfig(),
		Alerts:             ToGatusConfigList(spec.Alerts),
		MaintenanceWindows: ToGatusConfigList(spec.MaintenanceWindows),
		Client:             spec.Client.ToGatusConfig(),
		Ui:                 spec.Ui.ToGatusConfig(),
		ExtraLabels:        spec.ExtraLabels,
	}
}

// EndpointStatus defines the observed state of Endpoint.
type EndpointStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the Endpoint resource.
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

type EndpointDNSSpec struct {
	QueryType *string `json:"query-type,omitempty"`
	QueryName *string `json:"query-name,omitempty"`
}

func (spec *EndpointDNSSpec) ToGatusConfig() *gatusconfig.GatusEndpointDNSConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusEndpointDNSConfig{
		QueryType: spec.QueryType,
		QueryName: spec.QueryName,
	}
}

type EndpointSSHSpec struct {
	Username string `json:"username"` //TODO: secret ref/mounting
	Password string `json:"password"` //TODO: secret ref/mounting
}

func (spec *EndpointSSHSpec) ToGatusConfig() *gatusconfig.GatusEndpointSSHConfig {
	if spec == nil {
		return nil
	}
	// TODO: env substitution
	return &gatusconfig.GatusEndpointSSHConfig{
		Username: spec.Username,
		Password: spec.Password,
	}
}

type EndpointAlertSpec struct {
	Type                    string  `json:"type"`
	Enabled                 *bool   `json:"enabled,omitempty"`
	FailureThreshold        *int32  `json:"failure-threshold,omitempty"`
	SuccessThreshold        *int32  `json:"success-threshold,omitempty"`
	MinimumReminderInterval *string `json:"minimum-reminder-interval,omitempty"`
	SendOnResolved          *bool   `json:"send-on-resolved,omitempty"`
	Description             *string `json:"description,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields
	ProviderOverride map[string]runtime.RawExtension `json:"provider-override,omitempty"` // TODO: test if CRD passes these through or specific kubebuilder annotations are needed
}

func (spec *EndpointAlertSpec) ToGatusConfig() *gatusconfig.GatusEndpointAlertConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusEndpointAlertConfig{
		Type:                    spec.Type,
		Enabled:                 spec.Enabled,
		FailureThreshold:        spec.FailureThreshold,
		SuccessThreshold:        spec.SuccessThreshold,
		MinimumReminderInterval: spec.MinimumReminderInterval,
		SendOnResolved:          spec.SendOnResolved,
		Description:             spec.Description,
		ProviderOverride:        maps.Clone(spec.ProviderOverride),
	}
}

type EndpointClientSpec struct {
	Insecure           *bool                                 `json:"insecure,omitempty"`
	IgnoreRedirect     *bool                                 `json:"ignore-redirect,omitempty"`
	Timeout            *string                               `json:"timeout,omitempty"`
	DNSResolver        *string                               `json:"dns-resolver,omitempty"`
	Oauth2             *EndpointClientOauth2Spec             `json:"oauth2,omitempty"`
	ProxyURL           *string                               `json:"proxy-url,omitempty"`
	IdentityAwareProxy *EndpointClientIdentityAwareProxySpec `json:"identity-aware-proxy,omitempty"`
	MTLS               *EndpointClientmTLSSpec               `json:"tls,omitempty"`
	Network            *string                               `json:"network,omitempty"`
	// Tunnel             *string                              `json:"tunnel,omitempty"` // omit for now, as this is likely not required in a kubernetes setup. If requested, add implementation
}

func (spec *EndpointClientSpec) ToGatusConfig() *gatusconfig.GatusClientConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusClientConfig{
		Insecure:           spec.Insecure,
		IgnoreRedirect:     spec.IgnoreRedirect,
		Timeout:            spec.Timeout,
		DNSResolver:        spec.DNSResolver,
		Oauth2:             spec.Oauth2.ToGatusConfig(),
		ProxyURL:           spec.ProxyURL,
		IdentityAwareProxy: spec.IdentityAwareProxy.ToGatusConfig(),
		MTLS:               spec.MTLS.ToGatusConfig(),
		Network:            spec.Network,
		// Tunnel: , // omit for now, as this is likely not required in a kubernetes setup. If requested, add implementation
	}
}

type EndpointClientOauth2Spec struct {
	TokenURL     string   `json:"token-url"`     //TODO: secret ref/mounting
	ClientID     string   `json:"client-id"`     //TODO: secret ref/mounting
	ClientSecret string   `json:"client-secret"` //TODO: secret ref/mounting
	Scopes       []string `json:"scopes"`        //TODO: secret ref/mounting
}

func (spec *EndpointClientOauth2Spec) ToGatusConfig() *gatusconfig.GatusClientOauth2Config {
	if spec == nil {
		return nil
	}
	// TODO: set these fields via env substitution
	return &gatusconfig.GatusClientOauth2Config{
		TokenURL:     spec.TokenURL,
		ClientID:     spec.ClientID,
		ClientSecret: spec.ClientSecret,
		Scopes:       slices.Clone(spec.Scopes),
	}
}

type EndpointClientIdentityAwareProxySpec struct {
	Audience string `json:"audience"`
}

func (spec *EndpointClientIdentityAwareProxySpec) ToGatusConfig() *gatusconfig.GatusClientIdentityAwareProxyConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusClientIdentityAwareProxyConfig{
		Audience: spec.Audience,
	}
}

type EndpointClientmTLSSpec struct {
	CertificateFile *string `json:"certificate-file,omitempty"` //TODO: secret ref/mounting
	PrivateKeyFile  *string `json:"private-key-file,omitempty"` //TODO: secret ref/mounting
	Renegotiation   *string `json:"renegotiation,omitempty"`
}

func (spec *EndpointClientmTLSSpec) ToGatusConfig() *gatusconfig.GatusClientmTLSConfig {
	if spec == nil {
		return nil
	}
	// TODO: mount secrets into deployment and generate path
	return &gatusconfig.GatusClientmTLSConfig{
		CertificateFile: spec.CertificateFile,
		PrivateKeyFile:  spec.PrivateKeyFile,
		Renegotiation:   spec.Renegotiation,
	}
}

type EndpointUiSpec struct {
	HideConditions              *bool                `json:"hide-conditions,omitempty"`
	HideHostname                *bool                `json:"hide-hostname,omitempty"`
	HidePort                    *bool                `json:"hide-port,omitempty"`
	HideUrl                     *bool                `json:"hide-url,omitempty"`
	HideErrors                  *bool                `json:"hide-errors,omitempty"`
	DontResolveFailedConditions *bool                `json:"dont-resolve-failed-conditions,omitempty"`
	ResolveSuccessfulConditions *bool                `json:"resolve-successful-conditions,omitempty"`
	Badge                       *EndpointUiBadgeSpec `json:"badge,omitempty"`
}

func (spec *EndpointUiSpec) ToGatusConfig() *gatusconfig.GatusEndpointUiConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusEndpointUiConfig{
		HideConditions:              spec.HideConditions,
		HideHostname:                spec.HideHostname,
		HidePort:                    spec.HidePort,
		HideUrl:                     spec.HideUrl,
		HideErrors:                  spec.HideErrors,
		DontResolveFailedConditions: spec.DontResolveFailedConditions,
		ResolveSuccessfulConditions: spec.ResolveSuccessfulConditions,
		Badge:                       spec.Badge.ToGatusConfig(),
	}
}

type EndpointUiBadgeSpec struct {
	ResponseTime *EndpointUiBadgeResponseTimeSpec `json:"response-time,omitempty"`
}

func (spec *EndpointUiBadgeSpec) ToGatusConfig() *gatusconfig.GatusEndpointUiBadgeConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusEndpointUiBadgeConfig{
		ResponseTime: spec.ResponseTime.ToGatusConfig(),
	}
}

type EndpointUiBadgeResponseTimeSpec struct {
	Thresholds []int32 `json:"thresholds,omitempty"`
}

func (spec *EndpointUiBadgeResponseTimeSpec) ToGatusConfig() *gatusconfig.GatusEndpointUiBadgeResponseTimeConfig {
	if spec == nil {
		return nil
	}
	return &gatusconfig.GatusEndpointUiBadgeResponseTimeConfig{
		Thresholds: slices.Clone(spec.Thresholds),
	}
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Endpoint is the Schema for the endpoints API
type Endpoint struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Endpoint
	// +required
	Spec EndpointSpec `json:"spec"`

	// status defines the observed state of Endpoint
	// +optional
	Status EndpointStatus `json:"status,omitzero"`
}

func (crd *Endpoint) GetInstances() []client.ObjectKey {
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

// EndpointList contains a list of Endpoint
type EndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Endpoint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Endpoint{}, &EndpointList{})
}
