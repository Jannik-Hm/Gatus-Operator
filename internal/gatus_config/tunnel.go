package gatusconfig

type GatusTunnelingConfig struct {
	Type       string  `json:"type"`
	Host       string  `json:"host"`
	Port       *int32  `json:"port,omitempty"`
	Username   string  `json:"username"`
	Password   *string `json:"password,omitempty"`
	PrivateKey *string `json:"private-key,omitempty"`
}
