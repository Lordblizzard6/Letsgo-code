//go:build !windows

package config

import "os"

// restrictConfigPermissions limits the config file to the owner (API keys).
// No-op on platforms without POSIX semantics handled via build tags.
func restrictConfigPermissions(path string) error {
	return os.Chmod(path, 0o600)
}
