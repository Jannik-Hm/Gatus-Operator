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

	Maintenance *GatusMaintenanceConfig `json:"maintenance,omitempty"`

	Suites []GatusSuiteConfig `json:"suites,omitempty"`

	Tunneling map[string]GatusTunnelingConfig `json:"tunnel,omitempty"`
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

type GatusAlertingConfig any // TODO: test if CRD passes these through or specific kubebuilder annotations are needed or replace by actual structs

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
	ProviderOverride        map[string]any `json:"provider-override,omitempty"` // TODO: test if CRD passes these through or specific kubebuilder annotations are needed
}

type GatusEndpointConfig struct {
	Enabled            *bool                      `json:"enabled,omitempty"`
	Name               string                     `json:"name"`
	Group              *string                    `json:"group,omitempty"`
	URL                string                     `json:"url"`
	Method             *string                    `json:"method,omitempty"`
	Conditions         []string                   `json:"conditions,omitempty"`
	Interval           *string                    `json:"interval,omitempty"`
	GraphQL            *bool                      `json:"graphql,omitempty"`
	Body               *string                    `json:"body,omitempty"`
	Headers            map[string]string          `json:"headers,omitempty"`
	DNS                *GatusEndpointDNSConfig    `json:"dns,omitempty"`
	SSH                *GatusEndpointSSHConfig    `json:"ssh,omitempty"`
	Alerts             []GatusEndpointAlertConfig `json:"alerts,omitempty"`
	MaintenanceWindows []GatusMaintenanceConfig   `json:"maintenance-windows,omitempty"`
	Client             *GatusClientConfig         `json:"client,omitempty"`
	Ui                 *GatusEndpointUiConfig     `json:"ui,omitempty"`
	ExtraLabels        map[string]string          `json:"extra-labels,omitempty"`

	// Suites only
	AlwaysRun *bool             `json:"always-run,omitempty"`
	Store     map[string]string `json:"store,omitempty"`
}

type GatusEndpointDNSConfig struct {
	QueryType *string `json:"query-type,omitempty"`
	QueryName *string `json:"query-name,omitempty"`
}

type GatusEndpointSSHConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type GatusEndpointUiConfig struct {
	HideConditions              *bool                       `json:"hide-conditions,omitempty"`
	HideHostname                *bool                       `json:"hide-hostname,omitempty"`
	HidePort                    *bool                       `json:"hide-port,omitempty"`
	HideUrl                     *bool                       `json:"hide-url,omitempty"`
	HideErrors                  *bool                       `json:"hide-errors,omitempty"`
	DontResolveFailedConditions *bool                       `json:"dont-resolve-failed-conditions,omitempty"`
	ResolveSuccessfulConditions *bool                       `json:"resolve-successful-conditions,omitempty"`
	Badge                       *GatusEndpointUiBadgeConfig `json:"badge,omitempty"`
}

type GatusEndpointUiBadgeConfig struct {
	ResponseTime *GatusEndpointUiBadgeResponseTimeConfig `json:"response-time,omitempty"`
}

type GatusEndpointUiBadgeResponseTimeConfig struct {
	Thresholds []int32 `json:"thresholds,omitempty"`
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
	Title               *string               `json:"title,omitempty"`
	Description         *string               `json:"description,omitempty"`
	DashboardHeading    *string               `json:"dashboard-heading,omitempty"`
	DashboardSubheading *string               `json:"dashboard-subheading,omitempty"`
	Header              *string               `json:"header,omitempty"`
	Logo                *string               `json:"logo,omitempty"`
	Link                *string               `json:"link,omitempty"`
	Buttons             []GatusUiButtonConfig `json:"buttons,omitempty"`
	CustomCSS           *string               `json:"custom-css,omitempty"`
	Darkmode            *bool                 `json:"dark-mode,omitempty"`
	DefaultSortBy       *string               `json:"default-sort-by,omitempty"`
	DefaultFilterBy     *string               `json:"default-filter-by,omitempty"`
	LoginSubtitle       *string               `json:"login-subtitle,omitempty"`
}

type GatusUiFaviconConfig struct {
	Default   *string `json:"default,omitempty"`
	Size16x16 *string `json:"size16x16,omitempty"`
	Size32x32 *string `json:"size32x32,omitempty"`
}

type GatusUiButtonConfig struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

type GatusMaintenanceConfig struct {
	Enabled  *bool    `json:"enabled,omitempty"`
	Start    string   `json:"start"`
	Duration string   `json:"duration"`
	Timezone *string  `json:"timezone,omitempty"`
	Every    []string `json:"every,omitempty"`
}

type GatusSuiteConfig struct {
	Enabled   *bool                 `json:"enabled,omitempty"`
	Name      string                `json:"name"`
	Group     *string               `json:"group,omitempty"`
	Interval  *string               `json:"interval,omitempty"`
	Timeout   *string               `json:"timeout,omitempty"`
	Context   map[string]any        `json:"context,omitempty"`
	Endpoints []GatusEndpointConfig `json:"endpoints,omitempty"`
}

type GatusClientConfig struct {
	Insecure           *bool                                `json:"insecure,omitempty"`
	IgnoreRedirect     *bool                                `json:"ignore-redirect,omitempty"`
	Timeout            *string                              `json:"timeout,omitempty"`
	DNSResolver        *string                              `json:"dns-resolver,omitempty"`
	Oauth2             *GatusClientOauth2Config             `json:"oauth2,omitempty"`
	ProxyURL           *string                              `json:"proxy-url,omitempty"`
	IdentityAwareProxy *GatusClientIdentityAwareProxyConfig `json:"identity-aware-proxy,omitempty"`
	MTLS               *GatusClientmTLSConfig               `json:"tls,omitempty"`
	Network            *string                              `json:"network,omitempty"`
	Tunnel             *string                              `json:"tunnel,omitempty"`
}

type GatusClientOauth2Config struct {
	TokenURL     string   `json:"token-url"`
	ClientID     string   `json:"client-id"`
	ClientSecret string   `json:"client-secret"`
	Scopes       []string `json:"scopes"`
}

type GatusClientIdentityAwareProxyConfig struct {
	Audience string `json:"audience"`
}

type GatusClientmTLSConfig struct {
	CertificateFile *string `json:"certificate-file,omitempty"`
	PrivateKeyFile  *string `json:"private-key-file,omitempty"`
	Renegotiation   *string `json:"renegotiation,omitempty"`
}

type GatusTunnelingConfig struct {
	Type       string  `json:"type"`
	Host       string  `json:"host"`
	Port       *int32  `json:"port,omitempty"`
	Username   string  `json:"username"`
	Password   *string `json:"password,omitempty"`
	PrivateKey *string `json:"private-key,omitempty"`
}
