package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"grtbox/internal/tools"
)

const (
	ToolStoreIndexURL       = "https://grtbox.unaux.com/tools/org/tools.json"
	maxToolStoreIndexBytes  = 1024 * 1024
	maxToolStorePackageSize = 250 * 1024 * 1024
)

var infinityFreeChallengeRE = regexp.MustCompile(`var a=toNumbers\("([0-9a-fA-F]+)"\),b=toNumbers\("([0-9a-fA-F]+)"\),c=toNumbers\("([0-9a-fA-F]+)"\).*?location\.href="([^"]+)"`)

type ToolStoreIndex struct {
	Tools []string `json:"tools"`
}

type ToolStorePackage struct {
	URL              string                     `json:"url"`
	RegistryKey      string                     `json:"registry_key"`
	ID               string                     `json:"id"`
	Name             string                     `json:"name"`
	Version          string                     `json:"version"`
	Author           string                     `json:"author,omitempty"`
	Description      string                     `json:"description,omitempty"`
	Entry            string                     `json:"entry,omitempty"`
	Runtime          string                     `json:"runtime,omitempty"`
	IconData         string                     `json:"icon_data"`
	IconName         string                     `json:"icon_name"`
	Manifest         *tools.ToolManifest        `json:"manifest,omitempty"`
	Validation       tools.ToolValidationResult `json:"validation"`
	Installed        bool                       `json:"installed"`
	InstalledVersion string                     `json:"installed_version,omitempty"`
	UpdateAvailable  bool                       `json:"update_available"`
	InstalledToolID  string                     `json:"installed_tool_id,omitempty"`
}

func (a *App) ListToolStore() ([]ToolStorePackage, error) {
	urls, err := fetchToolStoreURLs(ToolStoreIndexURL)
	if err != nil {
		a.logger.Error(fmt.Sprintf("Tool Store index failed: %s", err))
		return nil, err
	}

	a.mu.RLock()
	registry := a.registry.Clone()
	a.mu.RUnlock()

	items := make([]ToolStorePackage, 0, len(urls))
	for _, packageURL := range urls {
		item := ToolStorePackage{
			URL:         packageURL,
			RegistryKey: storeRegistryKey(packageURL),
			Name:        "Unavailable Store Package",
			IconData:    tools.DefaultToolIconDataURI(),
			IconName:    tools.DefaultToolIconName,
			Validation:  tools.NewValidationResult(),
		}

		tempPath, cleanup, err := downloadToolStorePackage(packageURL)
		if err != nil {
			item.Validation.AddError(err.Error())
			items = append(items, item)
			continue
		}

		pkg, validation := tools.ValidatePackage(tempPath, CurrentGRTBoxVersion)
		if cleanup != nil {
			cleanup()
		}
		item.ID = pkg.ID
		item.Name = pkg.Name
		item.Version = pkg.Version
		item.Author = pkg.Author
		item.Description = pkg.Description
		item.Entry = pkg.Entry
		item.Runtime = pkg.Runtime
		item.IconData = pkg.IconData
		item.IconName = pkg.IconName
		item.Manifest = pkg.Manifest
		item.Validation = validation
		item.applyInstalledState(registry)

		items = append(items, item)
	}

	return items, nil
}

func (a *App) InstallStoreTool(packageURL string) (tools.ToolValidationResult, error) {
	tempPath, cleanup, err := downloadToolStorePackage(packageURL)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		result := tools.NewValidationResult()
		result.AddError(err.Error())
		return result, err
	}

	a.mu.RLock()
	currentRegistry := a.registry.Clone()
	a.mu.RUnlock()

	result, err := tools.InstallTool(tempPath, a.toolsDir, CurrentGRTBoxVersion, currentRegistry)
	if err != nil {
		a.logger.Error(fmt.Sprintf("Tool Store install failed for %s: %s", packageURL, err))
		return result, err
	}

	a.logger.Info(fmt.Sprintf("Tool installed from Tool Store: %s", packageURL))
	a.RefreshTools()
	return result, nil
}

func (a *App) UpdateStoreTool(packageURL string) (tools.ToolValidationResult, error) {
	tempPath, cleanup, err := downloadToolStorePackage(packageURL)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		result := tools.NewValidationResult()
		result.AddError(err.Error())
		return result, err
	}

	pkg, validation := tools.ValidatePackage(tempPath, CurrentGRTBoxVersion)
	if !validation.Valid {
		return validation, errors.New(validation.Message)
	}

	a.mu.RLock()
	currentRegistry := a.registry.Clone()
	a.mu.RUnlock()

	result, err := tools.UpdateTool(tempPath, a.toolsDir, CurrentGRTBoxVersion, currentRegistry)
	if err != nil {
		a.logger.Error(fmt.Sprintf("Tool Store update failed for %s: %s", packageURL, err))
		return result, err
	}

	if err := tools.RemoveExtractedPackage(a.extractedDir, pkg.ID); err != nil {
		a.logger.Warn(fmt.Sprintf("Failed to remove extracted files for %s after Tool Store update: %s", pkg.DisplayName(), err))
	}

	a.logger.Info(fmt.Sprintf("Tool updated from Tool Store: %s", packageURL))
	a.RefreshTools()
	return result, nil
}

func (item *ToolStorePackage) applyInstalledState(registry tools.ToolRegistry) {
	if item.ID == "" {
		return
	}

	for _, installed := range registry.Tools {
		if installed.ID != item.ID {
			continue
		}
		item.Installed = true
		item.InstalledVersion = installed.Version
		item.InstalledToolID = installed.RegistryKey
		item.UpdateAvailable = isStoreVersionNewer(item.Version, installed.Version)
		return
	}
}

func fetchToolStoreURLs(indexURL string) ([]string, error) {
	if err := validateStoreURL(indexURL); err != nil {
		return nil, err
	}

	client, err := newToolStoreHTTPClient(30 * time.Second)
	if err != nil {
		return nil, err
	}
	data, contentType, err := downloadToolStoreBytes(client, indexURL, maxToolStoreIndexBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to read Tool Store index: %w", err)
	}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || (trimmed[0] != '{' && trimmed[0] != '[') {
		return nil, fmt.Errorf("Tool Store did not return valid JSON. The remote URL returned %s instead of raw tools.json", contentTypeDescription(contentType))
	}

	var index ToolStoreIndex
	if trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &index.Tools); err != nil {
			return nil, fmt.Errorf("Tool Store JSON is invalid: %w", err)
		}
	} else if err := json.Unmarshal(trimmed, &index); err != nil {
		return nil, fmt.Errorf("Tool Store JSON is invalid: %w", err)
	}

	seen := map[string]bool{}
	urls := make([]string, 0, len(index.Tools))
	for _, raw := range index.Tools {
		raw = strings.TrimSpace(raw)
		if raw == "" || seen[raw] {
			continue
		}
		if err := validateStoreURL(raw); err != nil {
			return nil, err
		}
		seen[raw] = true
		urls = append(urls, raw)
	}
	return urls, nil
}

func downloadToolStorePackage(packageURL string) (string, func(), error) {
	if err := validateStoreURL(packageURL); err != nil {
		return "", nil, err
	}

	client, err := newToolStoreHTTPClient(90 * time.Second)
	if err != nil {
		return "", nil, err
	}
	data, _, err := downloadToolStoreBytes(client, packageURL, maxToolStorePackageSize)
	if err != nil {
		return "", nil, fmt.Errorf("failed to download .tl package: %w", err)
	}
	if len(data) > maxToolStorePackageSize {
		return "", nil, errors.New(".tl package is too large")
	}

	temp, err := os.CreateTemp("", "grtbox-store-*.tl")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temporary package file: %w", err)
	}
	tempPath := temp.Name()
	cleanup := func() {
		_ = os.Remove(tempPath)
	}

	_, err = temp.Write(data)
	if closeErr := temp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to save .tl package: %w", err)
	}
	return tempPath, cleanup, nil
}

func newToolStoreHTTPClient(timeout time.Duration) (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Jar:     jar,
		Timeout: timeout,
	}, nil
}

func downloadToolStoreBytes(client *http.Client, rawURL string, limit int64) ([]byte, string, error) {
	data, contentType, err := downloadToolStoreBytesOnce(client, rawURL, limit)
	if err != nil {
		return nil, contentType, err
	}
	if !looksLikeInfinityFreeChallenge(data) {
		return data, contentType, nil
	}

	retryURL, err := acceptInfinityFreeChallenge(client, rawURL, data)
	if err != nil {
		return nil, contentType, err
	}

	data, contentType, err = downloadToolStoreBytesOnce(client, retryURL, limit)
	if err != nil {
		return nil, contentType, err
	}
	if looksLikeInfinityFreeChallenge(data) {
		return nil, contentType, errors.New("remote host kept returning the JavaScript challenge after the Tool Store accepted it")
	}
	return data, contentType, nil
}

func downloadToolStoreBytesOnce(client *http.Client, rawURL string, limit int64) ([]byte, string, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "GRTBox/"+CurrentGRTBoxVersion)

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, contentType, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	data, err := readLimited(resp.Body, limit)
	if err != nil {
		return nil, contentType, err
	}
	return data, contentType, nil
}

func looksLikeInfinityFreeChallenge(data []byte) bool {
	return bytes.Contains(data, []byte(`document.cookie="__test="`)) &&
		bytes.Contains(data, []byte("slowAES.decrypt")) &&
		bytes.Contains(data, []byte("location.href="))
}

func acceptInfinityFreeChallenge(client *http.Client, rawURL string, data []byte) (string, error) {
	match := infinityFreeChallengeRE.FindSubmatch(data)
	if len(match) != 5 {
		return "", errors.New("Tool Store remote host returned a JavaScript challenge that could not be parsed")
	}

	cookieValue, err := decryptInfinityFreeCookie(string(match[1]), string(match[2]), string(match[3]))
	if err != nil {
		return "", err
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if client.Jar != nil {
		client.Jar.SetCookies(&url.URL{Scheme: parsed.Scheme, Host: parsed.Host}, []*http.Cookie{
			{
				Name:  "__test",
				Value: cookieValue,
				Path:  "/",
			},
		})
	}

	retryURL, err := url.Parse(string(match[4]))
	if err != nil {
		return "", err
	}
	if !retryURL.IsAbs() {
		retryURL = parsed.ResolveReference(retryURL)
	}
	return retryURL.String(), nil
}

func decryptInfinityFreeCookie(keyHex string, ivHex string, cipherHex string) (string, error) {
	keyBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", err
	}
	ivBytes, err := hex.DecodeString(ivHex)
	if err != nil {
		return "", err
	}
	cipherBytes, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", err
	}
	if len(cipherBytes) == 0 || len(cipherBytes)%aes.BlockSize != 0 {
		return "", errors.New("invalid Tool Store challenge cipher block size")
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}
	if len(ivBytes) != block.BlockSize() {
		return "", errors.New("invalid Tool Store challenge IV size")
	}

	plain := make([]byte, len(cipherBytes))
	cipher.NewCBCDecrypter(block, ivBytes).CryptBlocks(plain, cipherBytes)
	if len(plain) > aes.BlockSize {
		plain = trimPKCSLikePadding(plain)
	}
	return hex.EncodeToString(plain), nil
}

func trimPKCSLikePadding(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	padding := int(data[len(data)-1])
	if padding <= 0 || padding > aes.BlockSize || padding > len(data) {
		return data
	}
	for _, value := range data[len(data)-padding:] {
		if int(value) != padding {
			return data
		}
	}
	return data[:len(data)-padding]
}

func contentTypeDescription(contentType string) string {
	contentType = strings.TrimSpace(contentType)
	if contentType == "" {
		return "a non-JSON response"
	}
	return contentType
}

func readLimited(reader io.Reader, limit int64) ([]byte, error) {
	var buffer bytes.Buffer
	written, err := io.Copy(&buffer, io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if written > limit {
		return nil, errors.New("response is too large")
	}
	return buffer.Bytes(), nil
}

func validateStoreURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed == nil || parsed.Host == "" {
		return fmt.Errorf("invalid Tool Store URL: %s", raw)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("Tool Store URL must use http or https: %s", raw)
	}
	return nil
}

func isStoreVersionNewer(remote string, installed string) bool {
	comparison, err := tools.CompareSemver(remote, installed)
	if err == nil {
		return comparison > 0
	}
	return strings.TrimSpace(remote) != "" && strings.TrimSpace(remote) != strings.TrimSpace(installed)
}

func storeRegistryKey(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	source := rawURL
	if err == nil && parsed != nil {
		source = parsed.Host + "_" + strings.TrimSuffix(strings.TrimPrefix(parsed.EscapedPath(), "/"), filepath.Ext(parsed.Path))
	}

	var builder strings.Builder
	for _, r := range source {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '_' || r == '-' {
			builder.WriteRune(r)
			continue
		}
		builder.WriteRune('_')
	}
	out := strings.Trim(builder.String(), "._-")
	if out == "" {
		return "store_tool"
	}
	return out
}
