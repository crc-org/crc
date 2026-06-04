//go:build darwin

package preflight

import (
	"fmt"
	"os"
	"syscall"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
	crcos "github.com/crc-org/crc/v2/pkg/os"
)

func checkAdminHelperPrivileges(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSetuid == 0 {
		return fmt.Errorf("%s does not have the SUID bit set (%s)", path, fi.Mode().String())
	}
	if fi.Sys().(*syscall.Stat_t).Uid != 0 {
		return fmt.Errorf("%s is not owned by root", path)
	}

	return nil
}

func configureAdminHelperPrivileges(path string) error {
	logging.Debugf("Making %s suid", path)

	stdOut, stdErr, err := crcos.RunPrivileged(fmt.Sprintf("Changing ownership of %s", path), "chown", "root", path)
	if err != nil {
		return fmt.Errorf("unable to set ownership of %s to root: %s: %s: %w",
			path, stdOut, stdErr, err)
	}

	/* Can't do this before the chown as the chown will reset the suid bit */
	stdOut, stdErr, err = crcos.RunPrivileged(fmt.Sprintf("Setting suid for %s", path), "chmod", "u+s,g+x", path)
	if err != nil {
		return fmt.Errorf("unable to set suid bit on %s: %s: %s: %w", path, stdOut, stdErr, err)
	}
	return nil
}
