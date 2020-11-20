package os

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/code-ready/crc/pkg/crc/logging"
)

func WriteToFileAsRoot(reason, content, filepath string, mode os.FileMode) error {
	logging.Infof("Will use root access: %s", reason)
	cmd := exec.Command("sudo", "tee", filepath) // #nosec G204
	cmd.Stdin = strings.NewReader(content)
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed writing to file as root: %s: %s: %v", filepath, buf.String(), err)
	}
	if _, _, err := RunWithPrivilege(fmt.Sprintf("Changing permission for %s to %o ", filepath, mode),
		"chmod", strconv.FormatUint(uint64(mode), 8), filepath); err != nil {
		return err
	}
	return nil
}

func RemoveFileAsRoot(reason, filepath string) error {
	_, _, err := RunWithPrivilege(reason, "rm", "-fr", filepath)
	return err
}
