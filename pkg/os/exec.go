package os

import (
	"bytes"
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
	err = cmd.Run()
	return stdOut.String(), stdErr.String(), err
}
