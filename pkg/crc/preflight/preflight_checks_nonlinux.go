//go:build !linux
// +build !linux

package preflight

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/crc-org/crc/pkg/crc/daemonclient"
	crcerrors "github.com/crc-org/crc/pkg/crc/errors"
	"github.com/crc-org/crc/pkg/crc/logging"
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

// isDaemonProcess return true if any of the following conditions met
// - crc daemon <options>
// - crc --log-level=<level> daemon <options>
// - crc --log-level <level> daemon <options>
func isDaemonProcess(cmdLine []string) bool {
	if len(cmdLine) >= 2 && cmdLine[1] == "daemon" {
		return true
	}
	if len(cmdLine) >= 3 && strings.HasPrefix(cmdLine[1], "--log-level") && cmdLine[2] == "daemon" {
		return true
	}
	if len(cmdLine) >= 4 && cmdLine[1] == "--log-level" && cmdLine[3] == "daemon" {
		return true
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
