package gatusconfig

type GatusSecurityConfig struct {
	Basic *GatusSecurityBasicConfig `json:"basic,omitempty"`
	OIDC  *GatusSecurityOIDCConfig  `json:"oidc,omitempty"`
}

type GatusSecurityBasicConfig struct {
	Username string `json:"username"`
	PassHash string `json:"password-bcrypt-base64"`
}

type GatusSecurityOIDCConfig struct {
	IssuerURL       string   `json:"issuer-url"`
	RedirectURL     string   `json:"redirect-url"`
	ClientID        string   `json:"client-id"`
	ClientSecret    string   `json:"client-secret"`
	Scopes          []string `json:"scopes,omitempty"`
	AllowedSubjects []string `json:"allowed-subjects,omitempty"`
	SessionTTL      *string  `json:"session-ttl,omitempty"`
}
