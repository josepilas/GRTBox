package tools

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePackageAcceptsPortableToolWithoutIcon(t *testing.T) {
	packagePath := writeTestPackage(t, map[string]string{
		"manifest.json": `{
			"id": "portable_tool",
			"name": "Portable Tool",
			"version": "1.0.0",
			"entry": "main.tc",
			"runtime": "tc",
			"target_platforms": ["windows", "linux", "macos"],
			"min_grtbox_version": "0.1.0",
			"package_format_version": "1.0.0"
		}`,
		"main.tc": `export default async function main(runtime) { await runtime.logs.write("Portable"); }`,
	})

	pkg, result := ValidatePackage(packagePath, "0.1.0")
	if !result.Valid {
		t.Fatalf("expected package to be valid: %#v", result.Errors)
	}
	if pkg.IconName != DefaultToolIconName {
		t.Fatalf("expected default icon, got %q", pkg.IconName)
	}
}

func TestValidatePackageRejectsNonPortableEntryPath(t *testing.T) {
	packagePath := writeTestPackage(t, map[string]string{
		"manifest.json": `{
			"id": "bad_path_tool",
			"name": "Bad Path Tool",
			"version": "1.0.0",
			"entry": "../main.tc",
			"runtime": "tc"
		}`,
		"main.tc": `export default async function main(runtime) {}`,
	})

	_, result := ValidatePackage(packagePath, "0.1.0")
	if result.Valid {
		t.Fatalf("expected invalid package")
	}
	assertValidationContains(t, result.Errors, "Manifest entry path is not portable")
}

func TestValidatePackageRejectsCaseInsensitiveDuplicateZipEntries(t *testing.T) {
	packagePath := writeTestPackage(t, map[string]string{
		"manifest.json": `{
			"id": "duplicate_tool",
			"name": "Duplicate Tool",
			"version": "1.0.0",
			"entry": "main.tc",
			"runtime": "tc"
		}`,
		"main.tc": `export default async function main(runtime) {}`,
		"MAIN.TC": `export default async function main(runtime) {}`,
	})

	_, result := ValidatePackage(packagePath, "0.1.0")
	if result.Valid {
		t.Fatalf("expected invalid package")
	}
	assertValidationContains(t, result.Errors, "case-insensitive duplicate")
}

func TestValidatePackageRejectsNonTCRuntime(t *testing.T) {
	packagePath := writeTestPackage(t, map[string]string{
		"manifest.json": `{
			"id": "bad_runtime_tool",
			"name": "Bad Runtime Tool",
			"version": "1.0.0",
			"entry": "main.tc",
			"runtime": "native"
		}`,
		"main.tc": `export default async function main(runtime) {}`,
	})

	_, result := ValidatePackage(packagePath, "0.1.0")
	if result.Valid {
		t.Fatalf("expected invalid package")
	}
	assertValidationContains(t, result.Errors, "runtime must be")
}

func TestExamplePackagesValidate(t *testing.T) {
	examplesDir := filepath.Join("..", "..", "examples")
	if _, err := os.Stat(examplesDir); err != nil {
		t.Skip("example .tl packages are not included in the open source app snapshot")
	}

	for _, packageName := range []string{"sample_tool.tl", "internet_bridge.tl", "dtb_general.tl", "easy_firmware.tl", "portmaster_auto_installer.tl"} {
		t.Run(packageName, func(t *testing.T) {
			packagePath := filepath.Join(examplesDir, packageName)
			if _, err := os.Stat(packagePath); err != nil {
				t.Skipf("example package is not available in this checkout: %s", packageName)
			}
			_, result := ValidatePackage(packagePath, "0.1.0")
			if !result.Valid {
				t.Fatalf("expected %s to be valid: %#v", packageName, result.Errors)
			}
		})
	}
}

func writeTestPackage(t *testing.T, entries map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	packagePath := filepath.Join(dir, "test_tool.tl")
	file, err := os.Create(packagePath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	defer writer.Close()

	for name, contents := range entries {
		header := &zip.FileHeader{Name: name, Method: zip.Deflate}
		entry, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Ext(name) == ".json" && !json.Valid([]byte(contents)) {
			t.Fatalf("test entry %s is not valid JSON", name)
		}
		if _, err := entry.Write([]byte(contents)); err != nil {
			t.Fatal(err)
		}
	}

	return packagePath
}

func assertValidationContains(t *testing.T, messages []string, want string) {
	t.Helper()
	for _, message := range messages {
		if strings.Contains(message, want) {
			return
		}
	}
	t.Fatalf("expected validation message containing %q, got %#v", want, messages)
}
