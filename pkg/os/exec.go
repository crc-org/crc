package os

import (
	"bytes"
	"github.com/code-ready/crc/pkg/crc/logging"
	"os"
	"os/exec"
)

// RunWithPrivilege executes a command using sudo
// provide a reason why root is needed as the first argument
func RunWithPrivilege(reason string, cmdAndArgs ...string) (string, string, error) {
	sudo, err := exec.LookPath("sudo")
	if err != nil {
		return "", "", err
	}
	cmd := exec.Command(sudo, cmdAndArgs...) // #nosec G204
	stdOut := new(bytes.Buffer)
	stdErr := new(bytes.Buffer)
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr
	logging.Infof("Will use root access: %s", reason)
	err = cmd.Run()
	return stdOut.String(), stdErr.String(), err
}

func RunWithDefaultLocale(command string, args ...string) (string, string, error) {
	cmd := exec.Command(command, args...) // #nosec G204
	cmd.Env = ReplaceEnv(os.Environ(), "LC_ALL", "C")
	cmd.Env = ReplaceEnv(cmd.Env, "LANG", "C")
	stdOut := new(bytes.Buffer)
	stdErr := new(bytes.Buffer)
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr
	err := cmd.Run()
	return stdOut.String(), stdErr.String(), err
}
