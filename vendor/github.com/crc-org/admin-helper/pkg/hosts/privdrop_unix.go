//go:build darwin || linux

package hosts

import (
	"fmt"
	"syscall"
)

func DropPrivileges() error {
	return dropPrivileges()
}

var (
	sysSetgroups = syscall.Setgroups
	sysSetgid    = syscall.Setgid
	sysSetuid    = syscall.Setuid
	sysGetuid    = syscall.Getuid
	sysGetgid    = syscall.Getgid
	sysGeteuid   = syscall.Geteuid
	sysGetegid   = syscall.Getegid
)

// applyPrivilegeDrop lowers the process from setuid-root to targetUID/targetGID.
// Order matters: supplementary groups are cleared first, then the effective GID,
// then the effective UID. A final setuid(0) probe verifies root cannot be
// regained after the drop.
func applyPrivilegeDrop(targetUID, targetGID int) error {
	if err := sysSetgroups([]int{}); err != nil {
		return fmt.Errorf("failed to drop supplementary groups: %w", err)
	}

	if err := sysSetgid(targetGID); err != nil {
		return fmt.Errorf("failed to setgid(%d): %w", targetGID, err)
	}

	if err := sysSetuid(targetUID); err != nil {
		return fmt.Errorf("failed to setuid(%d): %w", targetUID, err)
	}

	if err := sysSetuid(0); err == nil {
		return fmt.Errorf("able to regain root after dropping privileges")
	}

	return nil
}

func dropPrivileges() error {
	targetUID := sysGetuid()
	targetGID := sysGetgid()

	if sysGeteuid() == targetUID && sysGetegid() == targetGID {
		return nil
	}

	return applyPrivilegeDrop(targetUID, targetGID)
}
