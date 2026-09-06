package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type ProjectPermissions struct {
	Version           string          `json:"version"`
	AllowedCategories map[string]bool `json:"allowed_categories"`
}

var (
	projectPermMu sync.RWMutex
	projectCache  = make(map[string]map[string]bool)
)

func permissionsFilePath(projectDir string) string {
	return filepath.Join(projectDir, ".letsGo", "permissions.json")
}

// GrantProject persists an allowed category for the given project directory.
func GrantProject(projectDir, category string) error {
	projectPermMu.Lock()
	defer projectPermMu.Unlock()

	dir := filepath.Clean(projectDir)
	permsDir := filepath.Join(dir, ".letsGo")
	_ = os.MkdirAll(permsDir, 0755)

	filePath := permissionsFilePath(dir)
	var perms ProjectPermissions
	if data, err := os.ReadFile(filePath); err == nil {
		_ = json.Unmarshal(data, &perms)
	}
	if perms.AllowedCategories == nil {
		perms.AllowedCategories = make(map[string]bool)
	}
	perms.Version = "1.0"
	perms.AllowedCategories[category] = true

	// Update cache
	if _, ok := projectCache[dir]; !ok {
		projectCache[dir] = make(map[string]bool)
	}
	projectCache[dir][category] = true

	data, err := json.MarshalIndent(perms, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// IsProjectGranted checks if the given category has been permanently granted for the project.
func IsProjectGranted(projectDir, category string) bool {
	projectPermMu.RLock()
	dir := filepath.Clean(projectDir)
	if m, ok := projectCache[dir]; ok {
		if allowed, ok := m[category]; ok {
			projectPermMu.RUnlock()
			return allowed
		}
	}
	projectPermMu.RUnlock()

	filePath := permissionsFilePath(dir)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	var perms ProjectPermissions
	if err := json.Unmarshal(data, &perms); err != nil {
		return false
	}

	allowed := perms.AllowedCategories[category]
	projectPermMu.Lock()
	if _, ok := projectCache[dir]; !ok {
		projectCache[dir] = make(map[string]bool)
	}
	projectCache[dir][category] = allowed
	projectPermMu.Unlock()

	return allowed
}
