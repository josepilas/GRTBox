package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"grtbox/internal/logs"
)

func ScanInstalledTools(dir string, currentVersion string, logger *logs.Logger) (ToolRegistry, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ToolRegistry{}, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return ToolRegistry{}, err
	}

	packages := []ToolPackage{}
	idCounts := map[string]int{}

	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".tl" {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())
		pkg, validation := ValidatePackage(fullPath, currentVersion)
		pkg.Validation = validation
		pkg.Metadata.ValidationStatus = validation.Message
		packages = append(packages, pkg)

		if pkg.ID != "" {
			idCounts[pkg.ID]++
		}
	}

	for i := range packages {
		if packages[i].ID != "" && idCounts[packages[i].ID] > 1 {
			packages[i].Validation.AddError("Duplicate tool id: " + packages[i].ID + ".")
			packages[i].Metadata.ValidationStatus = packages[i].Validation.Message
		}
	}

	sort.SliceStable(packages, func(i, j int) bool {
		return packages[i].DisplayName() < packages[j].DisplayName()
	})

	if logger != nil {
		logger.Info(fmt.Sprintf("Scanned %d tool package(s)", len(packages)))
	}

	return ToolRegistry{Tools: packages}, nil
}
