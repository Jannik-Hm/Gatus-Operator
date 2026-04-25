package config

import (
	"fmt"
	"os"
)

type Config struct {
	// default oci image of gatus to use
	DefaultGatusImage string

	// preference order of protocols, defaults to "https">"http">"tcp">"udp"
	ProtocolPreference []string

	WatchIngresses      bool
	WatchIngressClasses bool
	WatchHTTPRoutes     bool
	WatchGateways       bool
}

func Load() (*Config, error) {
	result := Config{}

	default_image, exists := os.LookupEnv("DEFAULT_GATUS_IMAGE")
	if !exists {
		return nil, fmt.Errorf("Missing env var for default gatus image (key DEFAULT_GATUS_IMAGE)")
	}
	result.DefaultGatusImage = default_image

	// TODO: make this user editable
	result.ProtocolPreference = []string{"https", "http", "tcp", "udp"}

	// TODO: make this user editable
	result.WatchIngresses = true
	result.WatchIngressClasses = true
	result.WatchHTTPRoutes = true
	result.WatchGateways = true

	return &result, nil
}
