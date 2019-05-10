package os

import (
	"bytes"
	"os"
	"os/exec"
)

func RunWithPrivilege(cmdAndArgs ...string) (string, string, error) {
	sudo, err := exec.LookPath("sudo")
	if err != nil {
		return "", "", err
	}
	cmd := exec.Command(sudo, cmdAndArgs...)
	stdOut := new(bytes.Buffer)
	stdErr := new(bytes.Buffer)
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr
	err = cmd.Run()
	return stdOut.String(), stdErr.String(), err
}

func RunWithDefaultLocale(command string, args ...string) (string, string, error) {
	cmd := exec.Command(command, args...)
	cmd.Env = ReplaceEnv(os.Environ(), "LC_ALL", "C")
	cmd.Env = ReplaceEnv(cmd.Env, "LANG", "C")
	stdOut := new(bytes.Buffer)
	stdErr := new(bytes.Buffer)
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr
	err := cmd.Run()
	return stdOut.String(), stdErr.String(), err
}
