// +build !windows

package cmd

import (
	"os"

	"github.com/code-ready/crc/pkg/crc/api"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
)

func runDaemon(config crcConfig.Storage) {
	// Remove if an old socket is present
	os.Remove(constants.DaemonSocketPath)
	crcAPIServer, err := api.CreateAPIServer(constants.DaemonSocketPath, config)
	if err != nil {
		logging.Fatal("Failed to launch daemon", err)
	}
	crcAPIServer.Serve()
}
