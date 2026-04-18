package config

import (
	"fmt"
	"os"
)

type Config struct {
	DefaultGatusImage  string
	ProtocolPreference []string
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

	return &result, nil
}
