package os

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
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

func GetFirstExistentPath(paths []string) (string, error) {
	readablePath := ""
	for _, path := range paths {
		logging.Debug(fmt.Sprintf("Trying %s..", path))
		_, err := os.Stat(path)
		if err != nil {
			logging.Debug(fmt.Sprintf("Failed to open %s: %s", path, err))
		} else {
			readablePath = path

			break
		}
	}
	if readablePath == "" {
		return "", fmt.Errorf("Failed to find a readable file on any of these paths: %s", strings.Join(paths, ", "))
	}

	return readablePath, nil
}

func ChownAsRoot(user *user.User, filepath string) error {
	logging.Infof("Will use root access to change owner & group of file %s to %s", filepath, user.Username)
	cmd := exec.Command("sudo", "chown", fmt.Sprintf("%s.%s", user.Uid, user.Gid), filepath) // #nosec G204
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to change owner & group of %s to %s: %s: %s: %v", filepath, user.Username, buf.String(), err)
	}
	return nil
}
