package gatusconfig

type GatusMaintenanceConfig struct {
	Enabled  *bool    `json:"enabled,omitempty"`
	Start    string   `json:"start"`
	Duration string   `json:"duration"`
	Timezone *string  `json:"timezone,omitempty"`
	Every    []string `json:"every,omitempty"`
}
