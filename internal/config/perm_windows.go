//go:build windows

package config

// restrictConfigPermissions is a no-op on Windows: the OS does not expose
// POSIX-style mode bits; owner access is enforced by the user profile ACL.
func restrictConfigPermissions(path string) error {
	return nil
}
