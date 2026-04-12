package gatusconfig

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
