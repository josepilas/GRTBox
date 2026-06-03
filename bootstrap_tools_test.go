package main

import (
	"os"
	"path/filepath"
	"testing"

	"grtbox/internal/logs"
	"grtbox/internal/tools"
)

func TestFirstRunToolIDsAreConfigured(t *testing.T) {
	expected := map[string]bool{
		"dtb_general":     false,
		"easy_firmware":   false,
		"internet_bridge": false,
	}

	for _, id := range firstRunToolIDs {
		if _, ok := expected[id]; !ok {
			t.Fatalf("unexpected first-run tool id: %s", id)
		}
		expected[id] = true
	}

	for id, found := range expected {
		if !found {
			t.Fatalf("missing first-run tool id: %s", id)
		}
	}
}

func TestAppendUniqueDoesNotDuplicateValues(t *testing.T) {
	values := appendUnique([]string{"dtb_general"}, "dtb_general")
	values = appendUnique(values, "easy_firmware")

	if len(values) != 2 {
		t.Fatalf("expected 2 values, got %#v", values)
	}
	if values[0] != "dtb_general" || values[1] != "easy_firmware" {
		t.Fatalf("unexpected values: %#v", values)
	}
}

func TestLiveToolStoreContainsFirstRunToolIDs(t *testing.T) {
	if os.Getenv("GRTBOX_LIVE_TOOLSTORE_TEST") != "1" {
		t.Skip("set GRTBOX_LIVE_TOOLSTORE_TEST=1 to check the live Tool Store")
	}

	urls, err := fetchToolStoreURLs(ToolStoreIndexURL)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]bool{}
	for _, id := range firstRunToolIDs {
		expected[id] = false
	}

	for _, packageURL := range urls {
		tempPath, cleanup, err := downloadToolStorePackage(packageURL)
		if cleanup != nil {
			defer cleanup()
		}
		if err != nil {
			t.Fatalf("failed to download %s: %v", packageURL, err)
		}

		pkg, validation := tools.ValidatePackage(tempPath, CurrentGRTBoxVersion)
		if !validation.Valid {
			t.Fatalf("invalid Tool Store package %s: %#v", packageURL, validation.Errors)
		}
		if _, ok := expected[pkg.ID]; ok {
			expected[pkg.ID] = true
		}
	}

	for id, found := range expected {
		if !found {
			t.Fatalf("live Tool Store is missing first-run tool id: %s", id)
		}
	}
}

func TestLiveFirstRunBootstrapInstallsDefaultTools(t *testing.T) {
	if os.Getenv("GRTBOX_LIVE_TOOLSTORE_TEST") != "1" {
		t.Skip("set GRTBOX_LIVE_TOOLSTORE_TEST=1 to check the live Tool Store")
	}

	stateDir := t.TempDir()
	app := &App{
		logger:             logs.NewLogger(),
		registry:           tools.ToolRegistry{},
		toolsDir:           filepath.Join(t.TempDir(), "tools"),
		extractedDir:       filepath.Join(t.TempDir(), "extracted_tools"),
		bootstrapStatePath: filepath.Join(stateDir, "first_run_bootstrap.json"),
	}

	app.RefreshTools()
	app.bootstrapDefaultToolsOnce()

	if _, err := os.Stat(app.bootstrapStatePath); err != nil {
		t.Fatalf("expected bootstrap state file to be created: %v", err)
	}

	app.RefreshTools()
	for _, id := range firstRunToolIDs {
		if _, ok := app.registry.Find(id); !ok {
			t.Fatalf("expected first-run tool to be installed: %s", id)
		}
	}
}
