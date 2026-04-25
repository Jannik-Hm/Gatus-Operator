package config

import (
	"fmt"
	"os"
	"strconv"
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

	// defaults
	result.WatchIngresses = true
	result.WatchIngressClasses = true
	result.WatchHTTPRoutes = true
	result.WatchGateways = true

	// get env settings
	watch_ingress, exists := os.LookupEnv("WATCH_INGRESS")
	if exists {
		val, err := strconv.ParseBool(watch_ingress)
		if err != nil {
			return nil, fmt.Errorf("error parsing WATCH_INGRESS: %w", err)
		}
		result.WatchIngresses = val
	}

	watch_ingress_class, exists := os.LookupEnv("WATCH_INGRESSCLASS")
	if exists {
		val, err := strconv.ParseBool(watch_ingress_class)
		if err != nil {
			return nil, fmt.Errorf("error parsing WATCH_INGRESSCLASS: %w", err)
		}
		result.WatchIngressClasses = val
	}

	watch_http_route, exists := os.LookupEnv("WATCH_HTTPROUTE")
	if exists {
		val, err := strconv.ParseBool(watch_http_route)
		if err != nil {
			return nil, fmt.Errorf("error parsing WATCH_HTTPROUTE: %w", err)
		}
		result.WatchHTTPRoutes = val
	}

	watch_gateway, exists := os.LookupEnv("WATCH_GATEWAY")
	if exists {
		val, err := strconv.ParseBool(watch_gateway)
		if err != nil {
			return nil, fmt.Errorf("error parsing WATCH_GATEWAY: %w", err)
		}
		result.WatchGateways = val
	}

	return &result, nil
}
