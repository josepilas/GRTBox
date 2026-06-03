package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync"

	"grtbox/internal/logs"
	"grtbox/internal/storage"
	"grtbox/internal/tools"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const CurrentGRTBoxVersion = "0.1.0"

type App struct {
	ctx                context.Context
	logger             *logs.Logger
	registry           tools.ToolRegistry
	toolsDir           string
	extractedDir       string
	bootstrapStatePath string
	mu                 sync.RWMutex
}

func NewApp() *App {
	toolsDir, err := storage.ToolsDir()
	if err != nil {
		toolsDir = "tools"
	}
	extractedDir, err := storage.ExtractedToolsDir()
	if err != nil {
		extractedDir = "extracted_tools"
	}
	bootstrapStatePath, err := storage.BootstrapStatePath()
	if err != nil {
		bootstrapStatePath = "first_run_bootstrap.json"
	}

	return &App{
		logger:             logs.NewLogger(),
		registry:           tools.ToolRegistry{},
		toolsDir:           toolsDir,
		extractedDir:       extractedDir,
		bootstrapStatePath: bootstrapStatePath,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.logger.Info("GRTBox starting")
	if err := os.MkdirAll(a.toolsDir, 0o755); err != nil {
		a.logger.Error(fmt.Sprintf("Failed to create tools directory: %s", err))
		return
	}
	if err := os.MkdirAll(a.extractedDir, 0o755); err != nil {
		a.logger.Error(fmt.Sprintf("Failed to create extracted tools directory: %s", err))
		return
	}
	a.RefreshTools()
	a.bootstrapDefaultToolsOnce()
}

func (a *App) ListTools() []tools.ToolPackage {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.registry.Clone().Tools
}

func (a *App) RefreshTools() []tools.ToolPackage {
	registry, err := tools.ScanInstalledTools(a.toolsDir, CurrentGRTBoxVersion, a.logger)
	if err != nil {
		a.logger.Error(fmt.Sprintf("Refresh failed: %s", err))
	}

	a.mu.Lock()
	a.registry = registry
	a.mu.Unlock()

	return registry.Clone().Tools
}

func (a *App) InstallTool(filePath string) (tools.ToolValidationResult, error) {
	if filePath == "" {
		result := tools.NewValidationResult()
		result.AddError("No package file selected.")
		return result, errors.New("no package file selected")
	}

	a.mu.RLock()
	currentRegistry := a.registry.Clone()
	a.mu.RUnlock()

	result, err := tools.InstallTool(filePath, a.toolsDir, CurrentGRTBoxVersion, currentRegistry)
	if err != nil {
		a.logger.Error(fmt.Sprintf("Install failed for %s: %s", filePath, err))
		return result, err
	}

	a.logger.Info(fmt.Sprintf("Tool installed from %s", filePath))
	a.RefreshTools()
	return result, nil
}

func (a *App) UpdateTool(filePath string) (tools.ToolValidationResult, error) {
	if filePath == "" {
		result := tools.NewValidationResult()
		result.AddError("No package file selected.")
		return result, errors.New("no package file selected")
	}

	pkg, validation := tools.ValidatePackage(filePath, CurrentGRTBoxVersion)
	if !validation.Valid {
		return validation, errors.New(validation.Message)
	}

	a.mu.RLock()
	currentRegistry := a.registry.Clone()
	a.mu.RUnlock()

	result, err := tools.UpdateTool(filePath, a.toolsDir, CurrentGRTBoxVersion, currentRegistry)
	if err != nil {
		a.logger.Error(fmt.Sprintf("Update failed for %s: %s", filePath, err))
		return result, err
	}

	if err := tools.RemoveExtractedPackage(a.extractedDir, pkg.ID); err != nil {
		a.logger.Warn(fmt.Sprintf("Failed to remove extracted files for %s after update: %s", pkg.DisplayName(), err))
	}

	a.logger.Info(fmt.Sprintf("Tool updated from %s", filePath))
	a.RefreshTools()
	return result, nil
}

func (a *App) PreviewToolPackage(filePath string) (tools.ToolPackage, error) {
	if filePath == "" {
		return tools.ToolPackage{}, errors.New("no package file selected")
	}
	pkg, result := tools.ValidatePackage(filePath, CurrentGRTBoxVersion)
	if !result.Valid {
		return pkg, errors.New(result.Message)
	}
	return pkg, nil
}

func (a *App) RemoveTool(toolID string) error {
	pkg, ok := a.findTool(toolID)
	if !ok {
		return fmt.Errorf("tool not found: %s", toolID)
	}
	if pkg.Location == "" {
		return fmt.Errorf("tool has no package location: %s", toolID)
	}

	if err := os.Remove(pkg.Location); err != nil {
		a.logger.Error(fmt.Sprintf("Remove failed for %s: %s", pkg.Name, err))
		return err
	}
	if err := tools.RemoveExtractedPackage(a.extractedDir, pkg.ID); err != nil {
		a.logger.Warn(fmt.Sprintf("Failed to remove extracted files for %s: %s", pkg.Name, err))
	}

	a.logger.Info(fmt.Sprintf("Tool removed: %s", pkg.DisplayName()))
	a.RefreshTools()
	return nil
}

func (a *App) GetToolDetails(toolID string) (tools.ToolPackage, error) {
	pkg, ok := a.findTool(toolID)
	if !ok {
		return tools.ToolPackage{}, fmt.Errorf("tool not found: %s", toolID)
	}
	return pkg, nil
}

func (a *App) OpenTool(toolID string) (tools.ToolPackage, error) {
	pkg, ok := a.findTool(toolID)
	if !ok {
		return tools.ToolPackage{}, fmt.Errorf("tool not found: %s", toolID)
	}
	if !pkg.Validation.Valid {
		return pkg, fmt.Errorf("package invalid: %s", pkg.Validation.Message)
	}
	extracted, err := tools.ExtractPackage(pkg, a.extractedDir)
	if err != nil {
		a.logger.Error(fmt.Sprintf("Failed to extract %s: %s", pkg.DisplayName(), err))
		return pkg, err
	}
	a.logger.Info(fmt.Sprintf("Tool opened: %s", extracted.DisplayName()))
	return extracted, nil
}

func (a *App) ValidateTool(filePath string) (tools.ToolValidationResult, error) {
	_, result := tools.ValidatePackage(filePath, CurrentGRTBoxVersion)
	if !result.Valid {
		return result, errors.New(result.Message)
	}
	return result, nil
}

func (a *App) GetLogs() []logs.Entry {
	return a.logger.Entries()
}

func (a *App) GetToolsDirectory() string {
	return a.toolsDir
}

func (a *App) GetExtractedToolsDirectory() string {
	return a.extractedDir
}

func (a *App) SelectToolPackage() (string, error) {
	if a.ctx == nil {
		return "", errors.New("application context is not ready")
	}

	return wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Install Tool",
		Filters: []wailsruntime.FileFilter{
			{
				DisplayName: "GRTBox Tool Package (*.tl)",
				Pattern:     "*.tl",
			},
		},
	})
}

func (a *App) OpenExternalURL(rawURL string) error {
	if a.ctx == nil {
		return errors.New("application context is not ready")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return fmt.Errorf("invalid external URL: %s", rawURL)
	}
	wailsruntime.BrowserOpenURL(a.ctx, rawURL)
	return nil
}

func (a *App) RunToolAction(toolID string, action string) (logs.Entry, error) {
	pkg, ok := a.findTool(toolID)
	if !ok {
		return logs.Entry{}, fmt.Errorf("tool not found: %s", toolID)
	}
	if action == "" {
		return logs.Entry{}, errors.New("action is empty")
	}

	if tools.IsLogAction(action) {
		entry := a.logger.Info(fmt.Sprintf("%s executed action %q", pkg.DisplayName(), action))
		return entry, nil
	}

	entry := a.logger.Warn(fmt.Sprintf("%s requested unsupported action %q", pkg.DisplayName(), action))
	return entry, nil
}

func (a *App) findTool(toolID string) (tools.ToolPackage, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.registry.Find(toolID)
}
