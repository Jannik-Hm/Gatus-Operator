package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GatusConfig struct {
	Metrics bool `json:"metrics"`

	Storage *GatusStorageConfig `json:"storage,omitempty"`

	Alerting *GatusAlertingConfig `json:"alerting,omitempty"`

	Announcements []GatusAnnouncementConfig `json:"announcements,omitempty"`

	Endpoints []GatusEndpointConfig `json:"endpoints,omitempty"`

	ExternalEndpoints []GatusExternalEndpointConfig `json:"external-endpoints,omitempty"`

	Security *GatusSecurityConfig `json:"security,omitempty"`

	Concurrency *int32 `json:"concurrency,omitempty"`

	DisableMonitoringLock *bool `json:"disable-monitoring-lock,omitempty"`

	SkipInvalidConfigUpdate *bool `json:"skip-invalid-config-update,omitempty"`

	Web *GatusWebConfig `json:"web,omitempty"`

	Ui *GatusUiConfig `json:"ui,omitempty"`

	Maintenance string `json:"maintenance,omitempty"`
}

// TODO: add kubebuilder markers

type GatusStorageConfig struct {
	Type              *GatusStorageType `json:"type,omitempty"`
	Path              *string           `json:"path,omitempty"`
	Caching           *bool             `json:"caching,omitempty"`
	MaximumResultsNum *int32            `json:"maximum-number-of-results,omitempty"`
	MaximumEventsNum  *int32            `json:"maximum-number-of-events,omitempty"`
}

type GatusStorageType string

const (
	GatusStorageTypeMemory   GatusStorageType = "memory"
	GatusStorageTypeSQLite   GatusStorageType = "sqlite"
	GatusStorageTypePostgres GatusStorageType = "postgres"
)

type GatusAlertingConfig struct {
}

type GatusAnnouncementConfig struct {
	Timestamp metav1.Time      `json:"timestamp"`
	Type      AnnouncementType `json:"type,omitempty"`
	Message   string           `json:"message"`
	Archived  *bool            `json:"archived,omitempty"`
}

type AnnouncementType string

const (
	AnnouncementTypeOutage      AnnouncementType = "outage"
	AnnouncementTypeWarning     AnnouncementType = "warning"
	AnnouncementTypeInformation AnnouncementType = "information"
	AnnouncementTypeOperational AnnouncementType = "operational"
	AnnouncementTypeNone        AnnouncementType = "none"
)

type GatusEndpointAlertConfig struct {
	Type                    string         `json:"type"`
	Enabled                 *bool          `json:"enabled,omitempty"`
	FailureThreshold        *int32         `json:"failure-threshold,omitempty"`
	SuccessThreshold        *int32         `json:"success-threshold,omitempty"`
	MinimumReminderInterval *string        `json:"minimum-reminder-interval,omitempty"`
	SendOnResolved          *bool          `json:"send-on-resolved,omitempty"`
	Description             *string        `json:"description,omitempty"`
	ProviderOverride        map[string]any `json:"provider-override,omitempty"` // TODO: test if this is accepted, otherwise may need to implement all providers...
}

type GatusEndpointConfig struct {
}

type GatusExternalEndpointConfig struct {
}

type GatusSecurityConfig struct {
}

type GatusWebConfig struct {
}

type GatusUiConfig struct {
}
