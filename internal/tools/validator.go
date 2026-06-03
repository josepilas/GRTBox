package tools

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func ValidatePackage(filePath string, currentVersion string) (ToolPackage, ToolValidationResult) {
	result := NewValidationResult()
	pkg := fallbackPackage(filePath, result)

	if strings.ToLower(filepath.Ext(filePath)) != ".tl" {
		result.AddError("Package file must use the .tl extension.")
		pkg.Validation = result
		return pkg, result
	}

	reader, err := zip.OpenReader(filePath)
	if err != nil {
		result.AddError("Package is not a valid ZIP-based .tl file.")
		pkg.Validation = result
		return pkg, result
	}
	defer reader.Close()

	validateZipLayout(reader.File, &result)

	manifestBytes, ok, err := readZipEntry(reader.File, "manifest.json")
	if err != nil {
		result.AddError("Failed to read manifest.json: " + err.Error())
		pkg.Validation = result
		return pkg, result
	}
	if !ok {
		result.AddError("manifest.json is required.")
		pkg.Validation = result
		return pkg, result
	}

	var manifest ToolManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		result.AddError("manifest.json is not valid JSON.")
		pkg.Validation = result
		return pkg, result
	}

	pkg = packageFromManifest(filePath, manifest, result)

	validateManifest(manifest, currentVersion, &result)

	_, ok, err = readZipEntry(reader.File, "main.tc")
	if err != nil {
		result.AddError("Failed to read main.tc: " + err.Error())
	}
	if !ok {
		result.AddError("main.tc is required.")
	}

	entryName := normalizeZipName(manifest.Entry)
	if entryName == "" {
		entryName = "main.tc"
	}
	_, entryOK, err := readZipEntry(reader.File, entryName)
	if err != nil {
		result.AddError("Failed to read entry file: " + err.Error())
	}
	if !entryOK {
		result.AddError("Entry file does not exist: " + entryName + ".")
	}

	iconName := normalizeZipName(manifest.Icon)
	if iconName == "" {
		iconName = "icon.png"
	}
	if iconBytes, ok, err := readZipEntry(reader.File, iconName); err == nil && ok {
		if strings.EqualFold(iconName, "icon.png") {
			if err := validatePNG(iconBytes); err != nil {
				result.AddError("icon.png is not a valid PNG image.")
			}
		}
		pkg.IconData = imageDataURI(iconName, iconBytes)
		pkg.IconName = iconName
		pkg.Metadata.UsesDefaultToolIcon = false
	} else {
		pkg.IconData = DefaultToolIconDataURI()
		pkg.IconName = DefaultToolIconName
		pkg.Metadata.UsesDefaultToolIcon = true
		if manifest.Icon != "" {
			result.AddWarning("Icon file was declared but not found; the default icon is being used.")
		}
	}

	pkg.Validation = result
	pkg.Metadata.ValidationStatus = result.Message
	return pkg, result
}

func validateManifest(manifest ToolManifest, currentVersion string, result *ToolValidationResult) {
	if manifest.ID == "" {
		result.AddError("manifest.json must include id.")
	}
	if manifest.Name == "" {
		result.AddError("manifest.json must include name.")
	}
	if manifest.Version == "" {
		result.AddError("manifest.json must include version.")
	}
	if manifest.Entry == "" {
		result.AddError("manifest.json must include entry.")
	} else if err := validatePortablePackagePath(manifest.Entry); err != nil {
		result.AddError("Manifest entry path is not portable: " + err.Error())
	} else if strings.ToLower(filepath.Ext(manifest.Entry)) != ".tc" {
		result.AddError("Manifest entry must point to a .tc file.")
	}
	if manifest.Runtime == "" {
		result.AddError("manifest.json must include runtime.")
	} else if !IsTCRuntime(manifest.Runtime) {
		result.AddError("manifest.json runtime must be \"tc\".")
	}
	if manifest.Icon != "" {
		if err := validatePortablePackagePath(manifest.Icon); err != nil {
			result.AddError("Manifest icon path is not portable: " + err.Error())
		}
	}
	if manifest.ID != "" && !isValidToolID(manifest.ID) {
		result.AddError("Tool id may only contain letters, numbers, dots, underscores, and hyphens.")
	}
	if manifest.MinGRTBoxVersion != "" {
		comparison, err := compareSemver(currentVersion, manifest.MinGRTBoxVersion)
		if err != nil {
			result.AddError("min_grtbox_version must be a semantic version.")
		} else if comparison < 0 {
			result.AddError(fmt.Sprintf("Package requires GRTBox %s or newer.", manifest.MinGRTBoxVersion))
		}
	}
	if manifest.PackageFormatVersion != "" {
		comparison, err := compareSemver("1.0.0", manifest.PackageFormatVersion)
		if err != nil {
			result.AddError("package_format_version must be a semantic version.")
		} else if comparison < 0 {
			result.AddError(fmt.Sprintf("Package format %s is newer than this GRTBox package loader.", manifest.PackageFormatVersion))
		}
	}
	if len(manifest.TargetPlatforms) > 0 && !containsPlatform(manifest.TargetPlatforms, runtime.GOOS) {
		result.AddWarning("Current platform is not listed in target_platforms.")
	}
}

func fallbackPackage(filePath string, result ToolValidationResult) ToolPackage {
	name := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	return ToolPackage{
		RegistryKey: registryKeyFromPath(filePath),
		ID:          name,
		Name:        name,
		Location:    filePath,
		IconData:    DefaultToolIconDataURI(),
		IconName:    DefaultToolIconName,
		Validation:  result,
		Metadata: ToolMetadata{
			ID:                  name,
			Name:                name,
			PackageLocation:     filePath,
			ValidationStatus:    result.Message,
			UsesDefaultToolIcon: true,
		},
	}
}

func packageFromManifest(filePath string, manifest ToolManifest, result ToolValidationResult) ToolPackage {
	return ToolPackage{
		RegistryKey: registryKeyFromPath(filePath),
		ID:          manifest.ID,
		Name:        manifest.Name,
		Version:     manifest.Version,
		Author:      manifest.Author,
		Description: manifest.Description,
		Entry:       manifest.Entry,
		Runtime:     manifest.Runtime,
		Location:    filePath,
		IconData:    DefaultToolIconDataURI(),
		IconName:    DefaultToolIconName,
		Manifest:    &manifest,
		Validation:  result,
		Metadata: ToolMetadata{
			ID:                   manifest.ID,
			Name:                 manifest.Name,
			Version:              manifest.Version,
			Author:               manifest.Author,
			Description:          manifest.Description,
			Entry:                manifest.Entry,
			Runtime:              manifest.Runtime,
			RequiresAdmin:        manifest.RequiresAdmin,
			Permissions:          manifest.Permissions,
			TargetPlatforms:      manifest.TargetPlatforms,
			MinGRTBoxVersion:     manifest.MinGRTBoxVersion,
			PackageFormatVersion: manifest.PackageFormatVersion,
			PackageLocation:      filePath,
			ValidationStatus:     result.Message,
			UsesDefaultToolIcon:  true,
		},
	}
}

func readZipEntry(files []*zip.File, entryName string) ([]byte, bool, error) {
	entryName = normalizeZipName(entryName)
	for _, file := range files {
		if normalizeZipName(file.Name) != entryName {
			continue
		}

		reader, err := file.Open()
		if err != nil {
			return nil, true, err
		}
		defer reader.Close()

		bytes, err := io.ReadAll(reader)
		if err != nil {
			return nil, true, err
		}
		return bytes, true, nil
	}
	return nil, false, nil
}

func validateZipLayout(files []*zip.File, result *ToolValidationResult) {
	seen := map[string]string{}
	for _, file := range files {
		if strings.Contains(file.Name, "\\") {
			result.AddError("ZIP entry path is not portable: " + file.Name + " (use forward slashes inside .tl packages)")
			continue
		}

		normalized := normalizeZipName(file.Name)
		if normalized == "" {
			continue
		}

		if err := validatePortablePackagePath(normalized); err != nil {
			result.AddError("ZIP entry path is not portable: " + file.Name + " (" + err.Error() + ")")
			continue
		}

		key := strings.ToLower(normalized)
		if existing, ok := seen[key]; ok && existing != normalized {
			result.AddError("ZIP contains case-insensitive duplicate entries: " + existing + " and " + normalized + ".")
			continue
		}
		seen[key] = normalized
	}
}

func normalizeZipName(name string) string {
	name = strings.TrimPrefix(path.Clean(strings.ReplaceAll(name, "\\", "/")), "./")
	if name == "." {
		return ""
	}
	return name
}

func validatePortablePackagePath(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("path is empty")
	}
	if strings.Contains(name, "\\") {
		return fmt.Errorf("use forward slashes inside .tl packages")
	}
	normalized := normalizeZipName(name)
	if normalized == "" {
		return fmt.Errorf("path is empty")
	}
	if normalized != name {
		return fmt.Errorf("path must already be normalized")
	}
	if strings.HasPrefix(normalized, "/") {
		return fmt.Errorf("absolute paths are not allowed")
	}
	if len(normalized) >= 2 && normalized[1] == ':' {
		return fmt.Errorf("Windows drive paths are not allowed")
	}
	for _, part := range strings.Split(normalized, "/") {
		if part == "" || part == "." || part == ".." {
			return fmt.Errorf("path cannot contain empty, current, or parent directory segments")
		}
		if strings.ContainsAny(part, `<>:"|?*`) {
			return fmt.Errorf("path contains characters that are not portable across supported systems")
		}
	}
	return nil
}

func ValidatePortablePackagePath(name string) error {
	return validatePortablePackagePath(name)
}

func validatePNG(data []byte) error {
	_, err := png.DecodeConfig(bytes.NewReader(data))
	return err
}

func imageDataURI(fileName string, data []byte) string {
	mimeType := "image/png"
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".webp":
		mimeType = "image/webp"
	case ".svg":
		mimeType = "image/svg+xml"
	}
	return "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data)
}

func isValidToolID(id string) bool {
	for _, r := range id {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '.' || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}

func containsPlatform(platforms []string, platform string) bool {
	for _, item := range platforms {
		normalized := strings.ToLower(strings.TrimSpace(item))
		if normalized == platform {
			return true
		}
		if normalized == "macos" && platform == "darwin" {
			return true
		}
	}
	return false
}

func compareSemver(left string, right string) (int, error) {
	leftParts, err := parseSemver(left)
	if err != nil {
		return 0, err
	}
	rightParts, err := parseSemver(right)
	if err != nil {
		return 0, err
	}

	for i := 0; i < 3; i++ {
		if leftParts[i] > rightParts[i] {
			return 1, nil
		}
		if leftParts[i] < rightParts[i] {
			return -1, nil
		}
	}
	return 0, nil
}

func CompareSemver(left string, right string) (int, error) {
	return compareSemver(left, right)
}

func parseSemver(version string) ([3]int, error) {
	var out [3]int
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	version = strings.Split(version, "-")[0]
	version = strings.Split(version, "+")[0]
	if version == "" {
		return out, fmt.Errorf("empty version")
	}

	parts := strings.Split(version, ".")
	if len(parts) > 3 {
		return out, fmt.Errorf("too many version parts")
	}

	for i, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil {
			return out, err
		}
		out[i] = value
	}

	return out, nil
}
