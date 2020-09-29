package config

import (
	"fmt"

	"github.com/spf13/cast"
)

func RequiresRestartMsg(key string, _ interface{}) string {
	return fmt.Sprintf("Changes to configuration property '%s' are only applied when a new CRC instance is created.\n"+
		"If you already have a CRC instance, then for this configuration change to take effect, "+
		"delete the CRC instance with 'crc delete' and start a new one with 'crc start'.", key)
}

func SuccessfullyApplied(key string, value interface{}) string {
	return fmt.Sprintf("Successfully configured %s to %s", key, cast.ToString(value))
}
