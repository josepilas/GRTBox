package tools

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func InstallTool(filePath string, toolsDir string, currentVersion string, registry ToolRegistry) (ToolValidationResult, error) {
	return installToolPackage(filePath, toolsDir, currentVersion, registry, false)
}

func UpdateTool(filePath string, toolsDir string, currentVersion string, registry ToolRegistry) (ToolValidationResult, error) {
	return installToolPackage(filePath, toolsDir, currentVersion, registry, true)
}

func installToolPackage(filePath string, toolsDir string, currentVersion string, registry ToolRegistry, updateExisting bool) (ToolValidationResult, error) {
	pkg, result := ValidatePackage(filePath, currentVersion)
	if !result.Valid {
		return result, errors.New(result.Message)
	}

	var existing *ToolPackage
	for _, installed := range registry.Tools {
		if installed.ID == pkg.ID {
			copy := installed
			existing = &copy
			break
		}
	}

	if existing != nil && !updateExisting {
		result.AddError("This tool is already in your library: " + pkg.Name + ".")
		return result, errors.New(result.Message)
	}
	if existing == nil && updateExisting {
		result.AddError("This tool is not installed yet: " + pkg.Name + ".")
		return result, errors.New(result.Message)
	}

	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		result.AddError("Failed to create tools directory.")
		return result, err
	}

	destination := filepath.Join(toolsDir, sanitizeFileName(pkg.ID)+".tl")
	if existing != nil && existing.Location != "" {
		destination = existing.Location
	}

	if _, err := os.Stat(destination); err == nil {
		if !updateExisting {
			result.AddError("Destination package already exists.")
			return result, errors.New(result.Message)
		}
	} else if !os.IsNotExist(err) {
		result.AddError("Failed to inspect destination package.")
		return result, err
	}

	if err := copyFile(filePath, destination); err != nil {
		result.AddError("Failed to copy package into the tools directory.")
		return result, err
	}

	if updateExisting {
		result.Message = "Tool Updated Successfully"
	} else {
		result.Message = "Tool Installed Successfully"
	}
	return result, nil
}

func copyFile(source string, destination string) error {
	sourceInfo, sourceErr := os.Stat(source)
	destinationInfo, destinationErr := os.Stat(destination)
	if sourceErr == nil && destinationErr == nil && os.SameFile(sourceInfo, destinationInfo) {
		return nil
	}

	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	temp, err := os.CreateTemp(filepath.Dir(destination), ".grtbox-install-*.tl")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)

	if _, err := io.Copy(temp, in); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tempName, destination); err != nil {
		if removeErr := os.Remove(destination); removeErr != nil && !os.IsNotExist(removeErr) {
			return err
		}
		if renameErr := os.Rename(tempName, destination); renameErr != nil {
			return renameErr
		}
	}
	return nil
}

func registryKeyFromPath(filePath string) string {
	name := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	if name == "" {
		return "tool"
	}
	return sanitizeFileName(name)
}

func sanitizeFileName(value string) string {
	var builder strings.Builder
	for _, r := range value {
		if r >= 'a' && r <= 'z' {
			builder.WriteRune(r)
			continue
		}
		if r >= 'A' && r <= 'Z' {
			builder.WriteRune(r)
			continue
		}
		if r >= '0' && r <= '9' {
			builder.WriteRune(r)
			continue
		}
		if r == '.' || r == '_' || r == '-' {
			builder.WriteRune(r)
			continue
		}
		builder.WriteRune('_')
	}

	out := strings.Trim(builder.String(), "._-")
	if out == "" {
		return "tool"
	}
	return out
}
