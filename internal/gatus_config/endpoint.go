package gatusconfig

import "k8s.io/apimachinery/pkg/runtime"

type GatusEndpointAlertConfig struct {
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

type GatusEndpointConfig struct {
	Enabled            *bool                       `json:"enabled,omitempty"`
	Name               string                      `json:"name"`
	Group              *string                     `json:"group,omitempty"`
	URL                string                      `json:"url"`
	Method             *string                     `json:"method,omitempty"`
	Conditions         []string                    `json:"conditions,omitempty"`
	Interval           *string                     `json:"interval,omitempty"`
	GraphQL            *bool                       `json:"graphql,omitempty"`
	Body               *string                     `json:"body,omitempty"`
	Headers            map[string]string           `json:"headers,omitempty"`
	DNS                *GatusEndpointDNSConfig     `json:"dns,omitempty"`
	SSH                *GatusEndpointSSHConfig     `json:"ssh,omitempty"`
	Alerts             []*GatusEndpointAlertConfig `json:"alerts,omitempty"`
	MaintenanceWindows []*GatusMaintenanceConfig   `json:"maintenance-windows,omitempty"`
	Client             *GatusClientConfig          `json:"client,omitempty"`
	Ui                 *GatusEndpointUiConfig      `json:"ui,omitempty"`
	ExtraLabels        map[string]string           `json:"extra-labels,omitempty"`

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
	Alerts    []*GatusEndpointAlertConfig          `json:"alerts,omitempty"`
	Heartbeat GatusExternalEndpointHeartbeatConfig `json:"heartbeat"`
}

type GatusExternalEndpointHeartbeatConfig struct {
	Interval *string `json:"interval,omitempty"`
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
	Tunnel             *string                              `json:"tunnel,omitempty"` // omit for now, as this is likely not required in a kubernetes setup. If requested, add implementation
}

type GatusClientOauth2Config struct {
	TokenURL     string   `json:"token-url"`     //TODO: secret ref/mounting
	ClientID     string   `json:"client-id"`     //TODO: secret ref/mounting
	ClientSecret string   `json:"client-secret"` //TODO: secret ref/mounting
	Scopes       []string `json:"scopes"`        //TODO: secret ref/mounting
}

type GatusClientIdentityAwareProxyConfig struct {
	Audience string `json:"audience"`
}

type GatusClientmTLSConfig struct {
	CertificateFile *string `json:"certificate-file,omitempty"` //TODO: secret ref/mounting
	PrivateKeyFile  *string `json:"private-key-file,omitempty"` //TODO: secret ref/mounting
	Renegotiation   *string `json:"renegotiation,omitempty"`
}
