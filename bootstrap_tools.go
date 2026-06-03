package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"grtbox/internal/tools"
)

const bootstrapStateVersion = "1.0.0"

var firstRunToolIDs = []string{
	"dtb_general",
	"easy_firmware",
	"internet_bridge",
}

type bootstrapState struct {
	Version      string   `json:"version"`
	ToolStoreURL string   `json:"tool_store_url"`
	AttemptedAt  string   `json:"attempted_at"`
	InstalledIDs []string `json:"installed_ids"`
	SkippedIDs   []string `json:"skipped_ids"`
	FailedIDs    []string `json:"failed_ids"`
	Errors       []string `json:"errors"`
}

func (a *App) bootstrapDefaultToolsOnce() {
	if a.bootstrapStatePath == "" {
		a.logger.Warn("First-run Tool Store bootstrap skipped: state path is not configured")
		return
	}

	if _, err := os.Stat(a.bootstrapStatePath); err == nil {
		a.logger.Info("First-run Tool Store bootstrap already completed")
		return
	} else if !errors.Is(err, os.ErrNotExist) {
		a.logger.Warn(fmt.Sprintf("First-run Tool Store bootstrap state could not be inspected: %s", err))
	}

	state := bootstrapState{
		Version:      bootstrapStateVersion,
		ToolStoreURL: ToolStoreIndexURL,
		AttemptedAt:  time.Now().Format(time.RFC3339),
		InstalledIDs: []string{},
		SkippedIDs:   []string{},
		FailedIDs:    []string{},
		Errors:       []string{},
	}

	a.logger.Info("First-run Tool Store bootstrap starting")
	missing := a.missingBootstrapToolIDs(&state)
	if len(missing) == 0 {
		a.logger.Info("First-run Tool Store bootstrap found all default tools already installed")
		a.writeBootstrapState(state)
		return
	}

	urls, err := fetchToolStoreURLs(ToolStoreIndexURL)
	if err != nil {
		message := fmt.Sprintf("Failed to read Tool Store index for first-run bootstrap: %s", err)
		a.logger.Error(message)
		state.Errors = append(state.Errors, message)
		state.FailedIDs = appendRemainingIDs(state.FailedIDs, missing)
		a.writeBootstrapState(state)
		return
	}

	for _, packageURL := range urls {
		if len(missing) == 0 {
			break
		}
		a.installBootstrapCandidate(packageURL, missing, &state)
	}

	for id := range missing {
		message := fmt.Sprintf("Default tool id was not found in Tool Store: %s", id)
		a.logger.Error(message)
		state.Errors = append(state.Errors, message)
		state.FailedIDs = appendUnique(state.FailedIDs, id)
	}

	a.writeBootstrapState(state)
	a.RefreshTools()
	a.logger.Info("First-run Tool Store bootstrap finished")
}

func (a *App) missingBootstrapToolIDs(state *bootstrapState) map[string]bool {
	a.mu.RLock()
	registry := a.registry.Clone()
	a.mu.RUnlock()

	missing := map[string]bool{}
	for _, id := range firstRunToolIDs {
		if _, ok := registry.Find(id); ok {
			state.SkippedIDs = appendUnique(state.SkippedIDs, id)
			a.logger.Info(fmt.Sprintf("First-run bootstrap skipped already installed tool: %s", id))
			continue
		}
		missing[id] = true
	}
	return missing
}

func (a *App) installBootstrapCandidate(packageURL string, missing map[string]bool, state *bootstrapState) {
	tempPath, cleanup, err := downloadToolStorePackage(packageURL)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		a.logger.Warn(fmt.Sprintf("First-run bootstrap skipped Tool Store package %s: %s", packageURL, err))
		return
	}

	pkg, validation := tools.ValidatePackage(tempPath, CurrentGRTBoxVersion)
	if !validation.Valid {
		a.logger.Warn(fmt.Sprintf("First-run bootstrap skipped invalid Tool Store package %s: %s", packageURL, validation.Message))
		return
	}
	if !missing[pkg.ID] {
		return
	}

	a.mu.RLock()
	registry := a.registry.Clone()
	a.mu.RUnlock()
	if _, ok := registry.Find(pkg.ID); ok {
		state.SkippedIDs = appendUnique(state.SkippedIDs, pkg.ID)
		delete(missing, pkg.ID)
		return
	}

	result, err := tools.InstallTool(tempPath, a.toolsDir, CurrentGRTBoxVersion, registry)
	if err != nil {
		message := fmt.Sprintf("Failed to install first-run tool %s from Tool Store: %s", pkg.ID, err)
		a.logger.Error(message)
		state.Errors = append(state.Errors, message)
		state.FailedIDs = appendUnique(state.FailedIDs, pkg.ID)
		delete(missing, pkg.ID)
		return
	}
	if !result.Valid {
		message := fmt.Sprintf("First-run tool %s was rejected: %s", pkg.ID, result.Message)
		a.logger.Error(message)
		state.Errors = append(state.Errors, message)
		state.FailedIDs = appendUnique(state.FailedIDs, pkg.ID)
		delete(missing, pkg.ID)
		return
	}

	state.InstalledIDs = appendUnique(state.InstalledIDs, pkg.ID)
	delete(missing, pkg.ID)
	a.logger.Info(fmt.Sprintf("First-run tool installed from Tool Store: %s", pkg.ID))
	a.RefreshTools()
}

func (a *App) writeBootstrapState(state bootstrapState) {
	if err := os.MkdirAll(filepath.Dir(a.bootstrapStatePath), 0o755); err != nil {
		a.logger.Error(fmt.Sprintf("Failed to create first-run bootstrap state directory: %s", err))
		return
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		a.logger.Error(fmt.Sprintf("Failed to encode first-run bootstrap state: %s", err))
		return
	}

	if err := os.WriteFile(a.bootstrapStatePath, data, 0o644); err != nil {
		a.logger.Error(fmt.Sprintf("Failed to write first-run bootstrap state: %s", err))
	}
}

func appendRemainingIDs(values []string, missing map[string]bool) []string {
	for id := range missing {
		values = appendUnique(values, id)
	}
	return values
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
