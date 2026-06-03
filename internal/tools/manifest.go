package tools

type ToolManifest struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Version              string   `json:"version"`
	Author               string   `json:"author,omitempty"`
	Description          string   `json:"description,omitempty"`
	Entry                string   `json:"entry"`
	Runtime              string   `json:"runtime"`
	Icon                 string   `json:"icon,omitempty"`
	RequiresAdmin        bool     `json:"requires_admin,omitempty"`
	Permissions          []string `json:"permissions,omitempty"`
	TargetPlatforms      []string `json:"target_platforms,omitempty"`
	MinGRTBoxVersion     string   `json:"min_grtbox_version,omitempty"`
	PackageFormatVersion string   `json:"package_format_version,omitempty"`
}

type ToolMetadata struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Version              string   `json:"version"`
	Author               string   `json:"author,omitempty"`
	Description          string   `json:"description,omitempty"`
	Entry                string   `json:"entry"`
	Runtime              string   `json:"runtime"`
	RequiresAdmin        bool     `json:"requires_admin,omitempty"`
	Permissions          []string `json:"permissions,omitempty"`
	TargetPlatforms      []string `json:"target_platforms,omitempty"`
	MinGRTBoxVersion     string   `json:"min_grtbox_version,omitempty"`
	PackageFormatVersion string   `json:"package_format_version,omitempty"`
	PackageLocation      string   `json:"package_location"`
	ValidationStatus     string   `json:"validation_status"`
	UsesDefaultToolIcon  bool     `json:"uses_default_tool_icon"`
}
