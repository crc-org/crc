//go:build !linux
// +build !linux

package preflight

import (
	"runtime"
	"strings"

	"github.com/code-ready/crc/pkg/crc/daemonclient"
	"github.com/code-ready/crc/pkg/crc/logging"
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

func olderDaemonVersionRunning() error {
	// Here daemonclient.GetVersionFromDaemonAPI() can return the error
	// if the daemon is not running or daemon version API is not responding
	// in both situation we can't check if daemon is running with an older
	// version of crc or not, so we are just ignoring the error from it.
	v, err := daemonclient.GetVersionFromDaemonAPI()
	if err != nil {
		return nil
	}
	return daemonclient.CheckIfOlderVersion(v)
}
