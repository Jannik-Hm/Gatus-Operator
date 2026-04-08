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
	Enabled   *bool                                `json:"enabled,omitempty"`
	Name      string                               `json:"name"`
	Group     *string                              `json:"group,omitempty"`
	Token     string                               `json:"token"`
	Alerts    []GatusEndpointAlertConfig           `json:"alerts,omitempty"`
	Heartbeat GatusExternalEndpointHeartbeatConfig `json:"heartbeat"`
}

type GatusExternalEndpointHeartbeatConfig struct {
	Interval *string `json:"interval,omitempty"`
}

type GatusSecurityConfig struct {
	Basic *GatusSecurityBasicConfig `json:"basic,omitempty"`
	OIDC  *GatusSecurityOIDCConfig  `json:"oidc,omitempty"`
}

type GatusSecurityBasicConfig struct {
	Username string `json:"username"`
	PassHash string `json:"password-bcrypt-base64"`
}

type GatusSecurityOIDCConfig struct {
	IssuerURL       string   `json:"issuer-url"`
	RedirectURL     string   `json:"redirect-url"`
	ClientID        string   `json:"client-id"`
	ClientSecret    string   `json:"client-secret"`
	Scopes          []string `json:"scopes,omitempty"`
	AllowedSubjects []string `json:"allowed-subjects,omitempty"`
	SessionTTL      *string  `json:"session-ttl,omitempty"`
}

type GatusWebConfig struct {
	Address *string `json:"address,omitempty"`
	// Port           *int32  `json:"port,omitempty"` // no sense in allowing a user to change the port
	ReadBufferSize *int32 `json:"read-buffer-size,omitempty"`
	// TLS            *GatusWebTLSConfig `json:"tls,omitempty"` // no sense in allowing a user to change tls
}

type GatusUiConfig struct {
}
