//go:build linux

package preflight

import (
	"fmt"
	"os"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
	crcos "github.com/crc-org/crc/v2/pkg/os"
)

const adminHelperCapability = "cap_dac_override"

func checkAdminHelperPrivileges(path string) error {
	getcap, _, err := crcos.RunWithDefaultLocale("getcap", path)
	if err != nil {
		return err
	}
	if !strings.Contains(getcap, adminHelperCapability+"=ep") {
		return fmt.Errorf("%s does not have the %s capability set (%s)", path, adminHelperCapability, strings.TrimSpace(getcap))
	}
	return nil
}

func configureAdminHelperPrivileges(path string) error {
	logging.Debugf("Setting %s capability for %s", adminHelperCapability, path)

	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSetuid != 0 {
		_, _, err = crcos.RunPrivileged(fmt.Sprintf("Removing legacy SUID bit from %s", path), "chmod", "u-s", path)
		if err != nil {
			return fmt.Errorf("unable to remove SUID bit from %s: %w", path, err)
		}
	}

	_, _, err = crcos.RunPrivileged(
		fmt.Sprintf("Setting %s capability for %s", adminHelperCapability, path),
		"setcap", adminHelperCapability+"=ep", path)
	if err != nil {
		return fmt.Errorf("unable to set %s capability on %s: %w", adminHelperCapability, path, err)
	}
	return nil
}
