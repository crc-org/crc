//go:build !linux
// +build !linux

package preflight

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"time"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
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

func sshPortCheck() Check {
	return Check{
		configKeySuffix:  "check-ssh-port",
		checkDescription: "Checking SSH port availability",
		check:            checkSSHPortFree(),
		fixDescription:   fmt.Sprintf("crc uses port %d to run SSH", constants.VsockSSHPort),
		flags:            NoFix,
		labels:           None,
	}
}

func checkSSHPortFree() func() error {
	return func() error {

		host := net.JoinHostPort(constants.LocalIP, strconv.Itoa(constants.VsockSSHPort))

		daemonClient := daemonclient.New()
		exposed, err := daemonClient.NetworkClient.List()
		if err == nil {
			// if port already exported by vsock we could proceed
			for _, e := range exposed {
				if e.Local == host {
					return nil
				}
			}
		}

		server, err := net.Listen("tcp", host)
		// if it fails then the port is likely taken
		if err != nil {
			return fmt.Errorf("port %d already in use: %s", constants.VsockSSHPort, err)
		}

		// we successfully used and closed the port
		return server.Close()
	}
}
