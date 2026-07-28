//go:build !darwin && !linux

package hosts

// DropPrivileges is a no-op on platforms that do not use setuid privilege dropping.
func DropPrivileges() error {
	return nil
}
