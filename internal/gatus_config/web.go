package gatusconfig

type GatusWebConfig struct {
	Address        *string `json:"address,omitempty"`
	Port           *int32  `json:"port,omitempty"`
	ReadBufferSize *int32  `json:"read-buffer-size,omitempty"`
	// TLS            *GatusWebTLSConfig `json:"tls,omitempty"` // no sense in allowing a user to change tls
}
