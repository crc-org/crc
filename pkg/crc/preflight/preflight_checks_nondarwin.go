// +build !darwin

package preflight

import (
	"errors"
	"net"

	"github.com/code-ready/crc/pkg/crc/constants"
)

const daemonNotRunningErrMsg = "Network mode 'vsock' requires 'crc daemon' to be running, run it manually on different terminal/tab"

var daemonRunningChecks = Check{
	configKeySuffix:  "check-if-daemon-running",
	checkDescription: "Checking if crc daemon is running",
	check:            checkIfCRCDaemonRunning,
	fixDescription:   daemonNotRunningErrMsg,
	flags:            NoFix,
}

func checkIfCRCDaemonRunning() error {
	l, err := net.Dial("unix", constants.DaemonSocketPath)
	if err != nil {
		// Log or report the error here
		return errors.New(daemonNotRunningErrMsg)
	}

	defer func() {
		if l != nil {
			l.Close()
		}
	}()

	return nil
}
