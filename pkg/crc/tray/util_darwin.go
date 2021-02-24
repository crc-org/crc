package tray

import (
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/os/launchd"
)

const (
	trayAgentLabel   = "crc.tray"
	daemonAgentLabel = "crc.daemon"
)

func DisableTrayAutostart() error {
	return launchd.RemovePlist(trayAgentLabel)
}

func DisableDaemonAutostart() error {
	return launchd.RemovePlist(daemonAgentLabel)
}

// ValidateTrayAutostart checks tray-auto-start is used in macOS and its a bool
func ValidateTrayAutostart(value interface{}) (bool, string) {
	return config.ValidateBool(value)
}
