package dns

import (
	goos "os"
	"os/exec"
	"syscall"

	"github.com/code-ready/crc/pkg/crc/logging"
	crcos "github.com/code-ready/crc/pkg/os"

	"github.com/code-ready/crc/pkg/crc/state"
)

func EnsureDNSDaemonRunning() error {
	if isRunning() {
		logging.InfoF("DNS daemon running with pid %d", state.GlobalState.DnsPID)
		return nil
	}

	proxyCmd, err := createDNSCommand()
	if err != nil {
		return err
	}

	err = proxyCmd.Start()
	if err != nil {
		return err
	}

	state.GlobalState.DnsPID = proxyCmd.Process.Pid
	state.GlobalState.Write()
	return nil
}

func createDNSCommand() (*exec.Cmd, error) {
	cmd, err := crcos.CurrentExecutable()
	if err != nil {
		return nil, err
	}

	args := []string{
		"daemon",
		"dns"}
	daemonCmd := exec.Command(cmd, args...)
	// don't inherit any file handles
	daemonCmd.Stderr = nil
	daemonCmd.Stdin = nil
	daemonCmd.Stdout = nil
	//exportCmd.SysProcAttr = process.SysProcForBackgroundProcess()
	//exportCmd.Env = process.EnvForBackgroundProcess()

	return daemonCmd, nil
}

func GetPID() int {
	if isRunning() {
		return state.GlobalState.DnsPID
	}
	return 0
}

func isRunning() bool {
	if state.GlobalState.DnsPID <= 0 {
		return false
	}

	process, err := goos.FindProcess(state.GlobalState.DnsPID)
	// TODO: make sure name matches
	if err != nil {
		return false
	}

	// for Windows FindProcess is enough
	if crcos.CurrentOS() == crcos.WINDOWS {
		return true
	}

	// for non Windows we need to send a signal to get more information
	err = process.Signal(syscall.Signal(0))
	if err == nil {
		return true
	} else {
		return false
	}
}
