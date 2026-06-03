package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallToolRejectsDuplicateAndUpdateToolReplacesIt(t *testing.T) {
	toolsDir := t.TempDir()
	currentVersion := "0.1.0"

	original := writeTestPackage(t, map[string]string{
		"manifest.json": `{
			"id": "update_me",
			"name": "Update Me",
			"version": "1.0.0",
			"entry": "main.tc",
			"runtime": "tc"
		}`,
		"main.tc": `export default async function main(runtime) { await runtime.logs.write("old"); }`,
	})
	if result, err := InstallTool(original, toolsDir, currentVersion, ToolRegistry{}); err != nil || !result.Valid {
		t.Fatalf("initial install failed: result=%#v err=%v", result, err)
	}

	registry, err := ScanInstalledTools(toolsDir, currentVersion, nil)
	if err != nil {
		t.Fatal(err)
	}

	updated := writeTestPackage(t, map[string]string{
		"manifest.json": `{
			"id": "update_me",
			"name": "Update Me",
			"version": "1.1.0",
			"entry": "main.tc",
			"runtime": "tc"
		}`,
		"main.tc": `export default async function main(runtime) { await runtime.logs.write("new"); }`,
	})

	if result, err := InstallTool(updated, toolsDir, currentVersion, registry); err == nil || result.Valid {
		t.Fatalf("duplicate install should be rejected: result=%#v err=%v", result, err)
	}

	if result, err := UpdateTool(updated, toolsDir, currentVersion, registry); err != nil || !result.Valid {
		t.Fatalf("update failed: result=%#v err=%v", result, err)
	}

	destination := filepath.Join(toolsDir, "update_me.tl")
	if _, err := os.Stat(destination); err != nil {
		t.Fatal(err)
	}
	pkg, result := ValidatePackage(destination, currentVersion)
	if !result.Valid {
		t.Fatalf("updated package is invalid: %#v", result.Errors)
	}
	if pkg.Version != "1.1.0" {
		t.Fatalf("expected updated version 1.1.0, got %q", pkg.Version)
	}
}
