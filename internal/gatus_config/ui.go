package gatusconfig

type GatusUiConfig struct {
	Title               *string                `json:"title,omitempty"`
	Description         *string                `json:"description,omitempty"`
	DashboardHeading    *string                `json:"dashboard-heading,omitempty"`
	DashboardSubheading *string                `json:"dashboard-subheading,omitempty"`
	Header              *string                `json:"header,omitempty"`
	Logo                *string                `json:"logo,omitempty"`
	Link                *string                `json:"link,omitempty"`
	Favicon             *GatusUiFaviconConfig  `json:"favicon,omitempty"`
	Buttons             []*GatusUiButtonConfig `json:"buttons,omitempty"`
	CustomCSS           *string                `json:"custom-css,omitempty"`
	Darkmode            *bool                  `json:"dark-mode,omitempty"`
	DefaultSortBy       *string                `json:"default-sort-by,omitempty"`
	DefaultFilterBy     *string                `json:"default-filter-by,omitempty"`
	LoginSubtitle       *string                `json:"login-subtitle,omitempty"`
}

type GatusUiFaviconConfig struct {
	Default   *string `json:"default,omitempty"`
	Size16x16 *string `json:"size16x16,omitempty"`
	Size32x32 *string `json:"size32x32,omitempty"`
}

type GatusUiButtonConfig struct {
	Name string `json:"name"`
	Link string `json:"link"`
}
