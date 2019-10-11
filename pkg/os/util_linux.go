package os

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/code-ready/crc/pkg/crc/logging"
)

func WriteToFileAsRoot(reason, content, filepath string) error {
	return writeToFileAsRoot(reason, content, filepath, false)
}

func AppendToFileAsRoot(reason, content, filepath string) error {
	return writeToFileAsRoot(reason, content, filepath, true)
}

func writeToFileAsRoot(reason, content, filepath string, append bool) error {
	logging.Infof("Will use root access: %s", reason)
	append_option := ""
	if append {
		append_option = "-a"
	}
	cmd := exec.Command("sudo", "tee", append_option, filepath) // #nosec G204
	cmd.Stdin = strings.NewReader(content)
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed writing to file as root: %s: %s: %v", filepath, buf.String(), err)
	}
	return nil
}
