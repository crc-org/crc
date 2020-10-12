// +build !windows

package cmd

import (
	"os"

	"github.com/code-ready/crc/pkg/crc/api"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
)

func runDaemon() {
	// Remove if an old socket is present
	os.Remove(constants.DaemonSocketPath)
	crcAPIServer, err := api.CreateAPIServer(constants.DaemonSocketPath, newConfig, newMachine())
	if err != nil {
		logging.Fatal("Failed to launch daemon", err)
	}
	crcAPIServer.Serve()
}
