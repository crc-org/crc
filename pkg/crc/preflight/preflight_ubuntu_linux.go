// +build linux

package preflight

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	crcos "github.com/code-ready/crc/pkg/os"
)

var ubuntuPreflightChecks = []Check{
	{
		configKeySuffix:    "check-apparmor-profile-setup",
		checkDescription:   "Checking if AppArmor is configured",
		check:              checkAppArmorExceptionIsPresent(ioutil.ReadFile),
		fixDescription:     "Updating AppArmor configuration",
		fix:                addAppArmorExceptionForQcowDisks(ioutil.ReadFile, crcos.WriteToFileAsRoot),
		cleanupDescription: "Cleaning up AppArmor configuration",
		cleanup:            removeAppArmorExceptionForQcowDisks(ioutil.ReadFile, crcos.WriteToFileAsRoot),
	},
}

const (
	appArmorTemplate = "/etc/apparmor.d/libvirt/TEMPLATE.qemu"
	appArmorHeader   = "profile LIBVIRT_TEMPLATE flags=(attach_disconnected) {"
)

type reader func(filename string) ([]byte, error)
type writer func(reason, content, filepath string, mode os.FileMode) error

// Verify the line is present in AppArmor template
func checkAppArmorExceptionIsPresent(reader reader) func() error {
	return func() error {
		template, err := reader(appArmorTemplate)
		if err != nil {
			return err
		}
		if !strings.Contains(string(template), expectedLines()) {
			return errors.New("AppArmor profile not configured")
		}
		return nil
	}
}

// Add the exception `cacheDir/*/crc.qcow2 rk` in AppArmor template
func addAppArmorExceptionForQcowDisks(reader reader, writer writer) func() error {
	return replaceInAppArmorTemplate(reader, writer, appArmorHeader, expectedLines())
}

// Eventually remove the exception in AppArmor template
func removeAppArmorExceptionForQcowDisks(reader reader, writer writer) func() error {
	return replaceInAppArmorTemplate(reader, writer, expectedLines(), appArmorHeader)
}

// Search for the string `before` in the AppArmor template and replace it with `after` string.
func replaceInAppArmorTemplate(reader reader, writer writer, before string, after string) func() error {
	return func() error {
		template, err := reader(appArmorTemplate)
		if err != nil {
			return err
		}
		if !strings.Contains(string(template), before) {
			return fmt.Errorf("unexpected AppArmor template file %s, cannot configure it automatically", appArmorTemplate)
		}
		content := strings.Replace(string(template), before, after, 1)
		return writer("Updating AppArmor configuration", content, appArmorTemplate, 0644)
	}
}

func expectedLines() string {
	line := fmt.Sprintf("  %s rk,", filepath.Join(constants.MachineCacheDir, "*", "crc.qcow2"))
	return fmt.Sprintf("%s\n%s\n", appArmorHeader, line)
}
