package config

import (
	"fmt"
	"path/filepath"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/os"
	"github.com/spf13/cast"
)

func RequiresRestartMsg(key string, _ interface{}) string {
	return fmt.Sprintf("Changes to configuration property '%s' are only applied when the CRC instance is started.\n"+
		"If you already have a running CRC instance, then for this configuration change to take effect, "+
		"stop the CRC instance with 'crc stop' and restart it with 'crc start'.", key)
}

func RequiresDeleteMsg(key string, _ interface{}) string {
	return fmt.Sprintf("Changes to configuration property '%s' are only applied when the CRC instance is created.\n"+
		"If you already have a running CRC instance, then for this configuration change to take effect, "+
		"delete the CRC instance with 'crc delete' and start it with 'crc start'.", key)
}

func RequiresDeleteAndSetupMsg(key string, value interface{}) string {

	if key == Preset && value == string(preset.Podman) {
		logging.Warn(preset.PodmanDeprecatedWarning)
	}
	// since we cannot easily import the machine package here to check for existence of the CRC vm
	// we rely on the existence of the machine config file to determine if a VM exists
	if os.FileExists(filepath.Join(constants.MachineInstanceDir, "crc", "config.json")) {
		return fmt.Sprintf("Changes to configuration property '%s' are only applied when the CRC instance is created.\n"+
			"If you already have a running CRC instance with different %s, then for this configuration change to take effect, "+
			"first delete the CRC instance with 'crc delete'. Then to confirm your system is ready, and you have the needed system bundle, "+
			"please run 'crc setup' before 'crc start'.", key, key)
	}
	return "To confirm your system is ready, and you have the needed system bundle, please run 'crc setup' before 'crc start'."
}

func SuccessfullyApplied(key string, value interface{}) string {
	return fmt.Sprintf("Successfully configured %s to %s", key, cast.ToString(value))
}

func RequiresCRCSetup(key string, _ interface{}) string {
	return fmt.Sprintf("Changes to configuration property '%s' are only applied during 'crc setup'.\n"+
		"Please run 'crc setup' for this configuration to take effect.", key)
}

func RequiresCleanupAndSetupMsg(key string, _ interface{}) string {
	return fmt.Sprintf("Changes to configuration property '%s' are only applied during 'crc setup'.\n"+
		"Please run 'crc cleanup' followed by 'crc setup' for this configuration to take effect.", key)
}

func RequiresHTTPPortChangeWarning(key string, value interface{}) string {
	return fmt.Sprintf("Changes to configuration property '%s' will break OpenShift HTTP routes.\n"+
		"In order to access OpenShift applications through HTTP URLs "+
		"the %d port must be manually specified, such as http://myapp.apps-crc.testing:%d", key, value, value)
}

func RequiresHTTPSPortChangeWarning(key string, value interface{}) string {
	return fmt.Sprintf("Changes to configuration property '%s' will break OpenShift HTTPS routes.\n"+
		"In order to access OpenShift applications through HTTPS URLs "+
		"the %d port must be manually specified, such as https://myapp.apps-crc.testing:%d\n"+
		"After this change, the OpenShift console will be non-functional because of OpenShift limitations", key, value, value)
}
