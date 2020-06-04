package cmd

import (
	"os"

	"github.com/code-ready/crc/pkg/crc/api"
	"github.com/code-ready/crc/pkg/crc/constants"
)

func runDaemon() {
	// Remove if an old socket is present
	os.Remove(constants.DaemonSocketPath)
	api.RunCrcDaemonService("CodeReady Containers", false)
}
