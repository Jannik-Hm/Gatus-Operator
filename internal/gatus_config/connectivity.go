package gatusconfig

type GatusConnectivityConfig struct {
	Checker GatusConnectivityCheckerConfig `json:"checker"`
}

type GatusConnectivityCheckerConfig struct {
	Target   string  `json:"target"`
	Interval *string `json:"interval,omitempty"`
}
