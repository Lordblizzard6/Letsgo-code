package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	versionCheckURL = "https://api.github.com/repos/user/claude-code-go/releases/latest"
	currentVersion  = "0.1.0"
)

// ReleaseInfo contains information about a release
type ReleaseInfo struct {
	Version     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Assets      []Asset   `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name       string `json:"name"`
	BrowserURL string `json:"browser_download_url"`
	Size       int64  `json:"size"`
}

// UpdateChecker handles version checking
type UpdateChecker struct {
	LastCheck time.Time
	LastInfo  *ReleaseInfo
}

// NewUpdateChecker creates a new update checker
func NewUpdateChecker() *UpdateChecker {
	return &UpdateChecker{}
}

// CheckForUpdates checks if a new version is available
func (uc *UpdateChecker) CheckForUpdates() (*ReleaseInfo, error) {
	resp, err := http.Get(versionCheckURL)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	uc.LastCheck = time.Now()
	uc.LastInfo = &release

	// Check if newer
	if isNewer(release.Version, currentVersion) {
		return &release, nil
	}

	return nil, nil // No update available
}

// isNewer checks if v1 is newer than v2
func isNewer(v1, v2 string) bool {
	// Simple version comparison (assumes semver format)
	return v1 != v2 && v1 > v2
}

// GetCurrentVersion returns the current version
func GetCurrentVersion() string {
	return currentVersion
}

// DownloadUpdate downloads the update for the current platform
func DownloadUpdate(release *ReleaseInfo) (string, error) {
	// Find appropriate asset
	asset := findAssetForPlatform(release.Assets)
	if asset == nil {
		return "", fmt.Errorf("no asset found for current platform (%s/%s)", runtime.GOOS, runtime.GOARCH)
	}

	// Download to temp location
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, asset.Name)

	resp, err := http.Get(asset.BrowserURL)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download error: %s", resp.Status)
	}

	file, err := os.Create(tempFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	return tempFile, nil
}

// findAssetForPlatform finds the appropriate asset for the current platform
func findAssetForPlatform(assets []Asset) *Asset {
	osSuffix := runtime.GOOS
	archSuffix := runtime.GOARCH

	switch runtime.GOOS {
	case "windows":
		osSuffix = "windows"
	case "darwin":
		osSuffix = "darwin"
	case "linux":
		osSuffix = "linux"
	}

	switch runtime.GOARCH {
	case "amd64":
		archSuffix = "amd64"
	case "arm64":
		archSuffix = "arm64"
	}

	for _, asset := range assets {
		name := asset.Name
		if contains(name, osSuffix) && contains(name, archSuffix) {
			return &asset
		}
	}

	// Fallback: just OS match
	for _, asset := range assets {
		if contains(asset.Name, osSuffix) {
			return &asset
		}
	}

	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ShouldAutoCheck returns true if auto-check should run
func ShouldAutoCheck() bool {
	lastCheck := getLastCheckTime()
	// Check once per day
	return time.Since(lastCheck) > 24*time.Hour
}

// getLastCheckTime gets the last update check time
func getLastCheckTime() time.Time {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".letsGo", "last_update_check")

	info, err := os.Stat(path)
	if err != nil {
		return time.Time{} // Never checked
	}

	return info.ModTime()
}

// SaveCheckTime saves the current time as last check
func SaveCheckTime() error {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".letsGo", "last_update_check")

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	file.Close()
	return nil
}
