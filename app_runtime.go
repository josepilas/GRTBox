package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"grtbox/internal/logs"
	"grtbox/internal/tools"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type RuntimeExecOptions struct {
	WorkingDirectory string            `json:"workingDirectory,omitempty"`
	TimeoutSeconds   int               `json:"timeoutSeconds,omitempty"`
	Env              map[string]string `json:"env,omitempty"`
	Input            string            `json:"input,omitempty"`
}

type RuntimeExecResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}

type RuntimeOSInfo struct {
	Platform      string `json:"platform"`
	Arch          string `json:"arch"`
	Family        string `json:"family"`
	PathSeparator string `json:"pathSeparator"`
	ListSeparator string `json:"listSeparator"`
	LineSeparator string `json:"lineSeparator"`
	HomeDir       string `json:"homeDir,omitempty"`
	ConfigDir     string `json:"configDir,omitempty"`
	CacheDir      string `json:"cacheDir,omitempty"`
	TempDir       string `json:"tempDir"`
}

type RuntimeStoragePaths struct {
	ToolID    string `json:"toolID,omitempty"`
	ConfigDir string `json:"configDir"`
	DataDir   string `json:"dataDir"`
	CacheDir  string `json:"cacheDir"`
	TempDir   string `json:"tempDir"`
}

type RuntimeFileInfo struct {
	Path         string `json:"path"`
	Exists       bool   `json:"exists"`
	IsDir        bool   `json:"isDir"`
	Size         int64  `json:"size"`
	ModifiedTime string `json:"modifiedTime,omitempty"`
}

type RuntimeDirectoryEntry struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	IsDir        bool   `json:"isDir"`
	Size         int64  `json:"size"`
	ModifiedTime string `json:"modifiedTime,omitempty"`
}

type RuntimeFileFilter struct {
	DisplayName string `json:"displayName"`
	Pattern     string `json:"pattern"`
}

type RuntimeOpenFileDialogOptions struct {
	Title   string              `json:"title,omitempty"`
	Filters []RuntimeFileFilter `json:"filters,omitempty"`
}

type RuntimeSaveFileDialogOptions struct {
	Title           string              `json:"title,omitempty"`
	DefaultFilename string              `json:"defaultFilename,omitempty"`
	Filters         []RuntimeFileFilter `json:"filters,omitempty"`
}

func (a *App) ReadToolModule(toolID string, relativePath string) (string, error) {
	pkg, ok := a.findTool(toolID)
	if !ok {
		return "", fmt.Errorf("tool not found: %s", toolID)
	}
	if pkg.Manifest == nil {
		return "", errors.New("tool manifest is not loaded")
	}
	if strings.ToLower(relativePath) == "entry" {
		relativePath = pkg.Entry
	}
	if strings.ToLower(relativePath) == "main" {
		relativePath = "main.tc"
	}
	if len(relativePath) < 3 || strings.ToLower(relativePath[len(relativePath)-3:]) != ".tc" {
		return "", errors.New("only .tc modules can be loaded")
	}

	path, err := tools.ResolveExtractedPackagePath(a.extractedDir, pkg.ID, pkg.Version, relativePath)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		if _, err := tools.ExtractPackage(pkg, a.extractedDir); err != nil {
			return "", err
		}
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (a *App) RuntimeLogsWrite(toolID string, message string) logs.Entry {
	if toolID == "" {
		return a.logger.Info(message)
	}
	return a.logger.Info(fmt.Sprintf("%s: %s", toolID, message))
}

func (a *App) RuntimeOSPlatform() string {
	return runtime.GOOS
}

func (a *App) RuntimeOSInfo() RuntimeOSInfo {
	info := RuntimeOSInfo{
		Platform:      runtime.GOOS,
		Arch:          runtime.GOARCH,
		Family:        platformFamily(runtime.GOOS),
		PathSeparator: string(os.PathSeparator),
		ListSeparator: string(os.PathListSeparator),
		LineSeparator: lineSeparator(),
		TempDir:       os.TempDir(),
	}
	if home, err := os.UserHomeDir(); err == nil {
		info.HomeDir = home
	}
	if config, err := os.UserConfigDir(); err == nil {
		info.ConfigDir = config
	}
	if cache, err := os.UserCacheDir(); err == nil {
		info.CacheDir = cache
	}
	return info
}

func (a *App) RuntimeOSIsAdmin() (bool, error) {
	if runtime.GOOS == "windows" {
		result, err := a.RuntimePowerShellExec("[Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent() | ForEach-Object { $_.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator) }", RuntimeExecOptions{TimeoutSeconds: 10})
		if err != nil {
			return false, err
		}
		return strings.EqualFold(strings.TrimSpace(result.Stdout), "true"), nil
	}
	result, err := runCommand("id", []string{"-u"}, RuntimeExecOptions{TimeoutSeconds: 5})
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(result.Stdout) == "0", nil
}

func (a *App) RuntimeProcessExec(command string, args []string, options RuntimeExecOptions) (RuntimeExecResult, error) {
	if command == "" {
		return RuntimeExecResult{ExitCode: -1}, errors.New("command is empty")
	}
	return runCommand(command, args, options)
}

func (a *App) RuntimeProcessWhich(command string) (string, error) {
	if strings.TrimSpace(command) == "" {
		return "", errors.New("command is empty")
	}
	path, err := exec.LookPath(command)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) RuntimeShellExec(command string, options RuntimeExecOptions) (RuntimeExecResult, error) {
	if command == "" {
		return RuntimeExecResult{ExitCode: -1}, errors.New("command is empty")
	}
	if runtime.GOOS == "windows" {
		return runCommand("cmd.exe", []string{"/C", command}, options)
	}
	return runCommand("sh", []string{"-c", command}, options)
}

func (a *App) RuntimePowerShellExec(script string, options RuntimeExecOptions) (RuntimeExecResult, error) {
	if script == "" {
		return RuntimeExecResult{ExitCode: -1}, errors.New("PowerShell script is empty")
	}
	powershell := "powershell.exe"
	if runtime.GOOS != "windows" {
		powershell = "pwsh"
	}
	return runCommand(powershell, []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", script}, options)
}

func (a *App) RuntimeFilesystemReadFile(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (a *App) RuntimeFilesystemReadFileBase64(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func (a *App) RuntimeFilesystemWriteFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

func (a *App) RuntimeFilesystemStat(path string) RuntimeFileInfo {
	info := RuntimeFileInfo{Path: path}
	stat, err := os.Stat(path)
	if err != nil {
		return info
	}
	info.Exists = true
	info.IsDir = stat.IsDir()
	info.Size = stat.Size()
	info.ModifiedTime = stat.ModTime().Format(time.RFC3339)
	return info
}

func (a *App) RuntimeFilesystemListDir(path string) ([]RuntimeDirectoryEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	out := make([]RuntimeDirectoryEntry, 0, len(entries))
	for _, entry := range entries {
		item := RuntimeDirectoryEntry{
			Name:  entry.Name(),
			Path:  filepath.Join(path, entry.Name()),
			IsDir: entry.IsDir(),
		}
		if info, err := entry.Info(); err == nil {
			item.Size = info.Size()
			item.ModifiedTime = info.ModTime().Format(time.RFC3339)
		}
		out = append(out, item)
	}
	return out, nil
}

func (a *App) RuntimeFilesystemMkdirAll(path string) error {
	return os.MkdirAll(path, 0o755)
}

func (a *App) RuntimeFilesystemRemoveFile(path string) error {
	return os.Remove(path)
}

func (a *App) RuntimeFilesystemRemoveDir(path string) error {
	return os.RemoveAll(path)
}

func (a *App) RuntimeFilesystemExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (a *App) RuntimeEnvGet(name string) string {
	return os.Getenv(name)
}

func (a *App) RuntimeStoragePaths(toolID string) (RuntimeStoragePaths, error) {
	toolID = sanitizeStorageSegment(toolID)
	configBase, err := os.UserConfigDir()
	if err != nil {
		return RuntimeStoragePaths{}, err
	}
	cacheBase, err := os.UserCacheDir()
	if err != nil {
		cacheBase = filepath.Join(configBase, "cache")
	}
	base := RuntimeStoragePaths{
		ToolID:    toolID,
		ConfigDir: filepath.Join(configBase, "GRTBox"),
		DataDir:   filepath.Join(configBase, "GRTBox", "data"),
		CacheDir:  filepath.Join(cacheBase, "GRTBox"),
		TempDir:   filepath.Join(os.TempDir(), "GRTBox"),
	}
	if toolID != "" {
		base.ConfigDir = filepath.Join(base.ConfigDir, "tools", toolID)
		base.DataDir = filepath.Join(base.DataDir, "tools", toolID)
		base.CacheDir = filepath.Join(base.CacheDir, "tools", toolID)
		base.TempDir = filepath.Join(base.TempDir, "tools", toolID)
	}
	return base, nil
}

func (a *App) RuntimeStorageEnsure(toolID string) (RuntimeStoragePaths, error) {
	paths, err := a.RuntimeStoragePaths(toolID)
	if err != nil {
		return paths, err
	}
	for _, path := range []string{paths.ConfigDir, paths.DataDir, paths.CacheDir, paths.TempDir} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return paths, err
		}
	}
	return paths, nil
}

func (a *App) RuntimePathJoin(parts []string) string {
	clean := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		clean = append(clean, filepath.FromSlash(part))
	}
	if len(clean) == 0 {
		return ""
	}
	return filepath.Clean(filepath.Join(clean...))
}

func (a *App) RuntimePathNormalize(path string) string {
	if path == "" {
		return ""
	}
	return filepath.Clean(filepath.FromSlash(path))
}

func (a *App) RuntimePathToSlash(path string) string {
	return filepath.ToSlash(path)
}

func (a *App) RuntimePathFromSlash(path string) string {
	return filepath.FromSlash(path)
}

func (a *App) RuntimePathBaseName(path string) string {
	return filepath.Base(path)
}

func (a *App) RuntimePathDirName(path string) string {
	return filepath.Dir(path)
}

func (a *App) RuntimePathExtName(path string) string {
	return filepath.Ext(path)
}

func (a *App) RuntimePathIsAbs(path string) bool {
	return filepath.IsAbs(path)
}

func (a *App) RuntimeCryptoHashFile(path string, algorithm string) (string, error) {
	hasher, err := newHash(algorithm)
	if err != nil {
		return "", err
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (a *App) RuntimeDialogsOpenFile(options RuntimeOpenFileDialogOptions) (string, error) {
	if a.ctx == nil {
		return "", errors.New("application context is not ready")
	}
	title := options.Title
	if title == "" {
		title = "Open File"
	}
	return wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title:   title,
		Filters: runtimeFileFilters(options.Filters),
	})
}

func (a *App) RuntimeDialogsSaveFile(options RuntimeSaveFileDialogOptions) (string, error) {
	if a.ctx == nil {
		return "", errors.New("application context is not ready")
	}
	title := options.Title
	if title == "" {
		title = "Save File"
	}
	return wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: options.DefaultFilename,
		Filters:         runtimeFileFilters(options.Filters),
	})
}

func runtimeFileFilters(filters []RuntimeFileFilter) []wailsruntime.FileFilter {
	out := make([]wailsruntime.FileFilter, 0, len(filters))
	for _, filter := range filters {
		if filter.DisplayName == "" || filter.Pattern == "" {
			continue
		}
		out = append(out, wailsruntime.FileFilter{
			DisplayName: filter.DisplayName,
			Pattern:     filter.Pattern,
		})
	}
	return out
}

func platformFamily(platform string) string {
	switch platform {
	case "windows":
		return "windows"
	case "darwin":
		return "macos"
	case "linux", "freebsd", "openbsd", "netbsd":
		return "unix"
	default:
		return platform
	}
}

func lineSeparator() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

func sanitizeStorageSegment(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var builder strings.Builder
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '_' || r == '-' {
			builder.WriteRune(r)
			continue
		}
		builder.WriteRune('_')
	}
	return strings.Trim(builder.String(), "._-")
}

func newHash(algorithm string) (hash.Hash, error) {
	switch strings.ToLower(strings.ReplaceAll(strings.TrimSpace(algorithm), "-", "")) {
	case "sha256", "":
		return sha256.New(), nil
	case "sha1":
		return sha1.New(), nil
	case "md5":
		return md5.New(), nil
	default:
		return nil, fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}
}

func runCommand(command string, args []string, options RuntimeExecOptions) (RuntimeExecResult, error) {
	timeout := options.TimeoutSeconds
	if timeout <= 0 {
		timeout = 30
	}
	if timeout > 600 {
		timeout = 600
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	configureBackgroundCommand(cmd)
	if options.WorkingDirectory != "" {
		cmd.Dir = options.WorkingDirectory
	}
	if len(options.Env) > 0 {
		cmd.Env = os.Environ()
		for key, value := range options.Env {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	if options.Input != "" {
		cmd.Stdin = strings.NewReader(options.Input)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := RuntimeExecResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	} else if ctx.Err() == context.DeadlineExceeded {
		result.ExitCode = -1
	} else {
		result.ExitCode = 0
	}
	if ctx.Err() == context.DeadlineExceeded {
		return result, fmt.Errorf("command timed out after %d seconds", timeout)
	}
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return result, nil
		}
		return result, err
	}
	return result, nil
}
