package config

import (
	"fmt"
)

func RequiresRestartMsg(key, _ string) string {
	return fmt.Sprintf("Changes to configuration property '%s' are only applied when a new CRC instance is created.\n"+
		"If you already have a CRC instance, then for this configuration change to take effect, "+
		"delete the CRC instance with 'crc delete' and start a new one with 'crc start'.", key)
}

func SuccessfullyApplied(key, value string) string {
	return fmt.Sprintf("Successfully configured %s to %s", key, value)
}
