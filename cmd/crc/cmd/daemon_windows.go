package cmd

import (
	"os"

	"github.com/code-ready/crc/pkg/crc/api"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
)

func runDaemon(config crcConfig.Storage) {
	// Remove if an old socket is present
	os.Remove(constants.DaemonSocketPath)
	api.RunCrcDaemonService("CodeReady Containers", false, config)
}
