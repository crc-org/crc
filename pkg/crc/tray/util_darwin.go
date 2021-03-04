package tray

import (
	"github.com/code-ready/crc/pkg/crc/config"
)

// ValidateTrayAutostart checks tray-auto-start is used in macOS and its a bool
func ValidateTrayAutostart(value interface{}) (bool, string) {
	return config.ValidateBool(value)
}
