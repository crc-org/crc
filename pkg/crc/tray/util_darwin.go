package tray

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/os/launchd"
	"github.com/spf13/cast"
)

var agentLabels = []string{"crc.tray", "crc.daemon"}

func DisableEnableTrayAutostart(key string, value interface{}) string {
	// Enable
	if cast.ToBool(value) {
		return config.RequiresCRCSetup(key, value)
	}
	// Disable
	for _, agentLabel := range agentLabels {
		if err := launchd.RemovePlist(agentLabel); err != nil {
			return fmt.Sprintf("Error trying to disable auto-start of tray: %s", err.Error())
		}
	}
	return "Successfully disabled auto-start of tray at login."
}

// ValidateTrayAutostart checks tray-auto-start is used in macOS and its a bool
func ValidateTrayAutostart(value interface{}) (bool, string) {
	return config.ValidateBool(value)
}
