package gatusconfig

import "slices"

type GatusMaintenanceConfig struct {
	Enabled  *bool    `json:"enabled,omitempty"`
	Start    string   `json:"start"`
	Duration string   `json:"duration"`
	Timezone *string  `json:"timezone,omitempty"`
	Every    []string `json:"every,omitempty"`
}

func (obj *GatusMaintenanceConfig) Clone() *GatusMaintenanceConfig {
	if obj == nil {
		return nil
	}
	return &GatusMaintenanceConfig{
		Enabled:  clonePtr(obj.Enabled),
		Start:    obj.Start,
		Duration: obj.Duration,
		Timezone: clonePtr(obj.Timezone),
		Every:    slices.Clone(obj.Every),
	}
}
