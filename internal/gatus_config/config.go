package gatusconfig

import (
	"k8s.io/apimachinery/pkg/runtime"
)

type Config struct {
	Metrics *bool `json:"metrics,omitempty"`

	Storage *GatusStorageConfig `json:"storage,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields
	Alerting *runtime.RawExtension `json:"alerting,omitempty"`

	Announcements []GatusAnnouncementConfig `json:"announcements,omitempty"` // TODO: custom CRD

	Endpoints []GatusEndpointConfig `json:"endpoints,omitempty"` // TODO: custom CRD

	ExternalEndpoints []GatusExternalEndpointConfig `json:"external-endpoints,omitempty"` // TODO: custom CRD

	Security *GatusSecurityConfig `json:"security,omitempty"`

	Concurrency *int32 `json:"concurrency,omitempty"`

	DisableMonitoringLock *bool `json:"disable-monitoring-lock,omitempty"` // this one is deprecated, rather use `Concurrency = 0`

	SkipInvalidConfigUpdate *bool `json:"skip-invalid-config-update,omitempty"` // irrelevant as updates will be handled outside of the configmap

	Web *GatusWebConfig `json:"web,omitempty"` // add read-buffer-size config to "Instance"

	Ui *GatusUiConfig `json:"ui,omitempty"`

	Maintenance *GatusMaintenanceConfig `json:"maintenance,omitempty"`

	Suites []GatusSuiteConfig `json:"suites,omitempty"` // TODO: custom CRD

	Tunneling map[string]GatusTunnelingConfig `json:"tunnel,omitempty"` // omit for now, as this is likely not required in a kubernetes setup. If requested, add implementation

	// Remote // omitted for now, as it is still in alpha and subject to change

	Connectivity *GatusConnectivityConfig `json:"connectivity,omitempty"`
}
