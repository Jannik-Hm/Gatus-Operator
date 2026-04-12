package gatusconfig

import "k8s.io/apimachinery/pkg/runtime"

type GatusSuiteConfig struct {
	Enabled  *bool   `json:"enabled,omitempty"`
	Name     string  `json:"name"`
	Group    *string `json:"group,omitempty"`
	Interval *string `json:"interval,omitempty"`
	Timeout  *string `json:"timeout,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields
	Context   map[string]runtime.RawExtension `json:"context,omitempty"`
	Endpoints []GatusEndpointConfig           `json:"endpoints,omitempty"`
}
