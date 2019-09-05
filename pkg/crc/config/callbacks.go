package config

import (
	"fmt"
)

func RequiresRestartMsg(key string, value string) string {
	return fmt.Sprintf("Changes to the '%s' setting with value '%s' is only applied when a new crc instance is created.\n"+
		"If you already have an existing crc instance, then to let the configuration changes take effect, "+
		"you must delete the current instance with 'crc delete' "+
		"and then start a new one with 'crc start'.", key, value)
}

func SuccessfullyApplied(key string, value string) string {
	return fmt.Sprintf("Successfully configured %s to %s", key, value)
}
