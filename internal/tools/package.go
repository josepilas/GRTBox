package tools

import (
	"encoding/base64"
	"path/filepath"
	"strings"
)

const DefaultToolIconName = "Default Tool Icon"

type ToolValidationResult struct {
	Valid    bool     `json:"valid"`
	Message  string   `json:"message"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

type ToolPackage struct {
	RegistryKey   string               `json:"registry_key"`
	ID            string               `json:"id"`
	Name          string               `json:"name"`
	Version       string               `json:"version"`
	Author        string               `json:"author,omitempty"`
	Description   string               `json:"description,omitempty"`
	Entry         string               `json:"entry,omitempty"`
	Runtime       string               `json:"runtime,omitempty"`
	Location      string               `json:"location"`
	ExtractedPath string               `json:"extracted_path,omitempty"`
	IconData      string               `json:"icon_data"`
	IconName      string               `json:"icon_name"`
	Manifest      *ToolManifest        `json:"manifest,omitempty"`
	Metadata      ToolMetadata         `json:"metadata"`
	Validation    ToolValidationResult `json:"validation"`
}

type ToolRegistry struct {
	Tools []ToolPackage `json:"tools"`
}

func NewValidationResult() ToolValidationResult {
	return ToolValidationResult{
		Valid:    true,
		Message:  "Package Valid",
		Errors:   []string{},
		Warnings: []string{},
	}
}

func (r *ToolValidationResult) AddError(message string) {
	r.Valid = false
	r.Message = "Package Invalid"
	r.Errors = append(r.Errors, message)
}

func (r *ToolValidationResult) AddWarning(message string) {
	r.Warnings = append(r.Warnings, message)
}

func (p ToolPackage) DisplayName() string {
	if p.Name != "" {
		return p.Name
	}
	if p.ID != "" {
		return p.ID
	}
	return strings.TrimSuffix(filepath.Base(p.Location), filepath.Ext(p.Location))
}

func (r ToolRegistry) Clone() ToolRegistry {
	out := ToolRegistry{Tools: make([]ToolPackage, len(r.Tools))}
	copy(out.Tools, r.Tools)
	return out
}

func (r ToolRegistry) Find(toolID string) (ToolPackage, bool) {
	for _, tool := range r.Tools {
		if tool.RegistryKey == toolID || tool.ID == toolID {
			return tool, true
		}
	}
	return ToolPackage{}, false
}

func DefaultToolIconDataURI() string {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="96" height="96" viewBox="0 0 96 96" role="img" aria-label="Default Tool Icon"><rect width="96" height="96" rx="18" fill="#252525"/><rect x="30" y="34" width="36" height="28" rx="3" fill="#e8e8e8"/><path d="M35 40h26v4H35zm0 9h18v4H35z" fill="#2d2d2d"/><path d="M24 46h6v8h-6zm42 0h6v8h-6z" fill="#8d8d8d"/><path d="M42 25h12v9H42z" fill="#8d8d8d"/></svg>`
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}
