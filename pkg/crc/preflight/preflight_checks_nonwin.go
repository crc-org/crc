// +build !windows

package preflight

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/code-ready/crc/pkg/crc/adminhelper"
	"github.com/code-ready/crc/pkg/crc/logging"
	crcos "github.com/code-ready/crc/pkg/os"
)

var nonWinPreflightChecks = [...]Check{
	{
		configKeySuffix:  "check-root-user",
		checkDescription: "Checking if running as non-root",
		check:            checkIfRunningAsNormalUser,
		fixDescription:   "crc should not be run as root",
		flags:            NoFix,
	},
	{
		cleanupDescription: "Removing hosts file records added by CRC",
		cleanup:            removeHostsFileEntry,
		flags:              CleanUpOnly,
	},
}

func checkIfRunningAsNormalUser() error {
	if os.Geteuid() != 0 {
		return nil
	}
	logging.Debug("Ran as root")
	return errors.New("crc should not be run as root")
}

func setSuid(path string) error {
	logging.Debugf("Making %s suid", path)

	stdOut, stdErr, err := crcos.RunPrivileged(fmt.Sprintf("Changing ownership of %s", path), "chown", "root", path)
	if err != nil {
		return fmt.Errorf("Unable to set ownership of %s to root: %s %v: %s",
			path, stdOut, err, stdErr)
	}

	/* Can't do this before the chown as the chown will reset the suid bit */
	stdOut, stdErr, err = crcos.RunPrivileged(fmt.Sprintf("Setting suid for %s", path), "chmod", "u+s,g+x", path)
	if err != nil {
		return fmt.Errorf("Unable to set suid bit on %s: %s %v: %s", path, stdOut, err, stdErr)
	}
	return nil
}

func checkSuid(path string) error {
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

func removeHostsFileEntry() error {
	err := adminhelper.CleanHostsFile()
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
