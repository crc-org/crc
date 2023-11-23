//go:build !linux
// +build !linux

package preflight

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/crc-org/crc/v2/pkg/crc/daemonclient"
	crcerrors "github.com/crc-org/crc/v2/pkg/crc/errors"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/shirou/gopsutil/v3/process"
)

func killDaemonProcess() error {
	crcProcessName := "crc"
	if runtime.GOOS == "windows" {
		crcProcessName = "crc.exe"
	}
	processes, err := process.Processes()
	if err != nil {
		return err
	}
	for _, process := range processes {
		// Ignore the error from process.Name() because errors
		// come from processes which are not owned by the current user
		// on Mac - https://github.com/shirou/gopsutil/issues/1288
		name, _ := process.Name()
		if name == crcProcessName {
			cmdLine, err := process.CmdlineSlice()
			if err != nil {
				return err
			}
			if isDaemonProcess(cmdLine) {
				logging.Debugf("Got the pid for %s : %d", name, process.Pid)
				if err := process.Kill(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// isDaemonProcess return true if the cmdline contains the word daemon
// since only one instance of the daemon is allowed to run this should
// be enough to detect the daemon process
func isDaemonProcess(cmdLine []string) bool {
	for _, arg := range cmdLine {
		if arg == "daemon" {
			return true
		}
	}
	return false
}

func daemonRunning() bool {
	if _, err := daemonclient.GetVersionFromDaemonAPI(); err != nil {
		return false
	}
	return true
}

func waitForDaemonRunning() error {
	return crcerrors.Retry(context.Background(), 15*time.Second, func() error {
		if !daemonRunning() {
			return &crcerrors.RetriableError{Err: fmt.Errorf("daemon is not running yet")}
		}
		return nil
	}, 2*time.Second)
}
