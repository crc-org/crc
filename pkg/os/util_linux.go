package os

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/code-ready/crc/pkg/crc/logging"
)

func WriteToFileAsRoot(reason, content, filepath string) error {
	logging.Infof("Will use root access: %s", reason)
	cmd := exec.Command("sudo", "tee", filepath) // #nosec G204
	cmd.Stdin = strings.NewReader(content)
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed writing to file as root: %s: %s: %v", filepath, buf.String(), err)
	}
	return nil
}

func RemoveFileAsRoot(reason, filepath string) error {
	_, _, err := RunWithPrivilege(reason, "rm", "-fr", filepath)
	return err
}
