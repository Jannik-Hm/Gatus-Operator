package gatusconfig

import (
	"maps"
	"slices"

	"k8s.io/apimachinery/pkg/runtime"
)

type GatusEndpointAlertConfig struct {
	Type                    string  `json:"type"`
	Enabled                 *bool   `json:"enabled,omitempty"`
	FailureThreshold        *int32  `json:"failure-threshold,omitempty"`
	SuccessThreshold        *int32  `json:"success-threshold,omitempty"`
	MinimumReminderInterval *string `json:"minimum-reminder-interval,omitempty"`
	SendOnResolved          *bool   `json:"send-on-resolved,omitempty"`
	Description             *string `json:"description,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields
	ProviderOverride map[string]runtime.RawExtension `json:"provider-override,omitempty"`
}

func (obj *GatusEndpointAlertConfig) Clone() *GatusEndpointAlertConfig {
	if obj == nil {
		return nil
	}
	c := GatusEndpointAlertConfig{
		Type:                    obj.Type,
		Enabled:                 clonePtr(obj.Enabled),
		FailureThreshold:        clonePtr(obj.FailureThreshold),
		SuccessThreshold:        clonePtr(obj.SuccessThreshold),
		MinimumReminderInterval: clonePtr(obj.MinimumReminderInterval),
		SendOnResolved:          clonePtr(obj.SendOnResolved),
		Description:             clonePtr(obj.Description),
	}

	if obj.ProviderOverride != nil {
		c.ProviderOverride = make(map[string]runtime.RawExtension, len(obj.ProviderOverride))
		for k, v := range obj.ProviderOverride {
			c.ProviderOverride[k] = *v.DeepCopy()
		}
	}
	return &c
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

func (obj *GatusEndpointConfig) Clone() *GatusEndpointConfig {
	if obj == nil {
		return nil
	}
	return &GatusEndpointConfig{
		Enabled:            clonePtr(obj.Enabled),
		Name:               obj.Name,
		Group:              clonePtr(obj.Group),
		URL:                obj.URL,
		Method:             clonePtr(obj.Method),
		GraphQL:            clonePtr(obj.GraphQL),
		Conditions:         slices.Clone(obj.Conditions),
		Interval:           clonePtr(obj.Interval),
		Body:               clonePtr(obj.Body),
		Headers:            maps.Clone(obj.Headers),
		ExtraLabels:        maps.Clone(obj.ExtraLabels),
		AlwaysRun:          clonePtr(obj.AlwaysRun),
		Store:              maps.Clone(obj.Store),
		Alerts:             listClone(obj.Alerts),
		MaintenanceWindows: listClone(listClone(obj.MaintenanceWindows)),
		DNS:                obj.DNS.Clone(),
		SSH:                obj.SSH.Clone(),
		Client:             obj.Client.Clone(),
		Ui:                 obj.Ui.Clone(),
	}
}

func (obj *GatusEndpointConfig) Merge(configs ...*GatusEndpointConfig) *GatusEndpointConfig {
	merged := obj.Clone()

	dns_configs := make([]*GatusEndpointDNSConfig, 0)

	for _, cfg := range configs {
		merged.Enabled = FillIfNotValue(merged.Enabled, cfg.Enabled, nil)

		merged.Name = FillIfNotValue(merged.Name, cfg.Name, "")

		merged.Group = FillIfNotValue(merged.Group, cfg.Group, nil)

		merged.URL = FillIfNotValue(merged.URL, cfg.URL, "")

		merged.Method = FillIfNotValue(merged.Method, cfg.Method, nil)

		merged.GraphQL = FillIfNotValue(merged.GraphQL, cfg.GraphQL, nil)

		// merged.Conditions = FillIfNotValue(merged.GraphQL, cfg.GraphQL, nil)

		merged.Interval = FillIfNotValue(merged.Interval, cfg.Interval, nil)

		merged.Body = FillIfNotValue(merged.Body, cfg.Body, nil)

		// merged.Headers = FillIfNotValue(merged.GraphQL, cfg.GraphQL, nil)

		// merged.ExtraLabels = FillIfNotValue(merged.GraphQL, cfg.GraphQL, nil)

		merged.AlwaysRun = FillIfNotValue(merged.AlwaysRun, cfg.AlwaysRun, nil)

		// merged.Store = FillIfNotValue(merged.GraphQL, cfg.GraphQL, nil)

		// merged.Alerts = FillIfNotValue(merged.GraphQL, cfg.GraphQL, nil)

		// merged.MaintenanceWindows = FillIfNotValue(merged.GraphQL, cfg.GraphQL, nil)

		dns_configs = append(dns_configs, cfg.DNS)
		merged.DNS = FillIfNotValue(merged.DNS, cfg.DNS, nil)

		merged.SSH = FillIfNotValue(merged.SSH, cfg.SSH, nil)

		merged.Client = FillIfNotValue(merged.Client, cfg.Client, nil)

		merged.Ui = FillIfNotValue(merged.Ui, cfg.Ui, nil)
	}

	merged.DNS = obj.DNS.Merge(dns_configs...)

	return merged
}

type GatusEndpointDNSConfig struct {
	QueryType *string `json:"query-type,omitempty"`
	QueryName *string `json:"query-name,omitempty"`
}

func (obj *GatusEndpointDNSConfig) Clone() *GatusEndpointDNSConfig {
	if obj == nil {
		return nil
	}
	return &GatusEndpointDNSConfig{
		QueryType: clonePtr(obj.QueryType),
		QueryName: clonePtr(obj.QueryName),
	}
}

func (obj *GatusEndpointDNSConfig) Merge(dns_configs ...*GatusEndpointDNSConfig) *GatusEndpointDNSConfig {
	panic("unimplemented")
}

type GatusEndpointSSHConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (obj *GatusEndpointSSHConfig) Clone() *GatusEndpointSSHConfig {
	if obj == nil {
		return nil
	}
	return &GatusEndpointSSHConfig{
		Username: obj.Username,
		Password: obj.Password,
	}
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

func (obj *GatusEndpointUiConfig) Clone() *GatusEndpointUiConfig {
	if obj == nil {
		return nil
	}
	return &GatusEndpointUiConfig{
		HideConditions:              clonePtr(obj.HideConditions),
		HideHostname:                clonePtr(obj.HideHostname),
		HidePort:                    clonePtr(obj.HidePort),
		HideUrl:                     clonePtr(obj.HideUrl),
		HideErrors:                  clonePtr(obj.HideErrors),
		DontResolveFailedConditions: clonePtr(obj.DontResolveFailedConditions),
		ResolveSuccessfulConditions: clonePtr(obj.ResolveSuccessfulConditions),
		Badge:                       obj.Badge.Clone(),
	}
}

type GatusEndpointUiBadgeConfig struct {
	ResponseTime *GatusEndpointUiBadgeResponseTimeConfig `json:"response-time,omitempty"`
}

func (obj *GatusEndpointUiBadgeConfig) Clone() *GatusEndpointUiBadgeConfig {
	if obj == nil {
		return nil
	}
	return &GatusEndpointUiBadgeConfig{
		ResponseTime: obj.ResponseTime.Clone(),
	}
}

type GatusEndpointUiBadgeResponseTimeConfig struct {
	Thresholds []int32 `json:"thresholds,omitempty"`
}

func (obj *GatusEndpointUiBadgeResponseTimeConfig) Clone() *GatusEndpointUiBadgeResponseTimeConfig {
	if obj == nil {
		return nil
	}
	return &GatusEndpointUiBadgeResponseTimeConfig{
		Thresholds: slices.Clone(obj.Thresholds),
	}
}

type GatusExternalEndpointConfig struct {
	Enabled   *bool                                `json:"enabled,omitempty"`
	Name      string                               `json:"name"`
	Group     *string                              `json:"group,omitempty"`
	Token     string                               `json:"token"`
	Alerts    []*GatusEndpointAlertConfig          `json:"alerts,omitempty"`
	Heartbeat GatusExternalEndpointHeartbeatConfig `json:"heartbeat"`
}

func (obj *GatusExternalEndpointConfig) Clone() *GatusExternalEndpointConfig {
	if obj == nil {
		return nil
	}
	return &GatusExternalEndpointConfig{
		Enabled:   clonePtr(obj.Enabled),
		Name:      obj.Name,
		Group:     clonePtr(obj.Group),
		Token:     obj.Token,
		Heartbeat: obj.Heartbeat.Clone(),
		Alerts:    listClone(obj.Alerts),
	}
}

type GatusExternalEndpointHeartbeatConfig struct {
	Interval *string `json:"interval,omitempty"`
}

func (obj GatusExternalEndpointHeartbeatConfig) Clone() GatusExternalEndpointHeartbeatConfig {
	return GatusExternalEndpointHeartbeatConfig{
		Interval: clonePtr(obj.Interval),
	}
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

func (obj *GatusClientConfig) Clone() *GatusClientConfig {
	if obj == nil {
		return nil
	}
	return &GatusClientConfig{
		Insecure:           clonePtr(obj.Insecure),
		IgnoreRedirect:     clonePtr(obj.IgnoreRedirect),
		Timeout:            clonePtr(obj.Timeout),
		DNSResolver:        clonePtr(obj.DNSResolver),
		ProxyURL:           clonePtr(obj.ProxyURL),
		Network:            clonePtr(obj.Network),
		Tunnel:             clonePtr(obj.Tunnel),
		Oauth2:             obj.Oauth2.Clone(),
		IdentityAwareProxy: obj.IdentityAwareProxy.Clone(),
		MTLS:               obj.MTLS.Clone(),
	}
}

type GatusClientOauth2Config struct {
	TokenURL     string   `json:"token-url"`     //TODO: secret ref/mounting
	ClientID     string   `json:"client-id"`     //TODO: secret ref/mounting
	ClientSecret string   `json:"client-secret"` //TODO: secret ref/mounting
	Scopes       []string `json:"scopes"`        //TODO: secret ref/mounting
}

func (obj *GatusClientOauth2Config) Clone() *GatusClientOauth2Config {
	if obj == nil {
		return nil
	}
	return &GatusClientOauth2Config{
		TokenURL:     obj.TokenURL,
		ClientID:     obj.ClientID,
		ClientSecret: obj.ClientSecret,
		Scopes:       slices.Clone(obj.Scopes),
	}
}

type GatusClientIdentityAwareProxyConfig struct {
	Audience string `json:"audience"`
}

func (obj *GatusClientIdentityAwareProxyConfig) Clone() *GatusClientIdentityAwareProxyConfig {
	if obj == nil {
		return nil
	}
	return &GatusClientIdentityAwareProxyConfig{
		Audience: obj.Audience,
	}
}

type GatusClientmTLSConfig struct {
	CertificateFile *string `json:"certificate-file,omitempty"` //TODO: secret ref/mounting
	PrivateKeyFile  *string `json:"private-key-file,omitempty"` //TODO: secret ref/mounting
	Renegotiation   *string `json:"renegotiation,omitempty"`
}

func (obj *GatusClientmTLSConfig) Clone() *GatusClientmTLSConfig {
	if obj == nil {
		return nil
	}
	return &GatusClientmTLSConfig{
		CertificateFile: clonePtr(obj.CertificateFile),
		PrivateKeyFile:  clonePtr(obj.PrivateKeyFile),
		Renegotiation:   clonePtr(obj.Renegotiation),
	}
}
