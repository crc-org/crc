package machine

import (
	"fmt"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	crcos "github.com/code-ready/crc/pkg/os"
)

func copyDiskImage(destDir string) (string, string, error) {
	const destFormat = "qcow2"

	imageName := fmt.Sprintf("%s.qcow2", constants.DefaultName)

	srcPath := filepath.Join(constants.MachineInstanceDir, constants.DefaultName, imageName)
	destPath := filepath.Join(destDir, imageName)

	_, _, err := crcos.RunWithDefaultLocale("qemu-img", "convert", "-f", "qcow2", "-O", destFormat, srcPath, destPath)
	if err != nil {
		return "", "", err
	}

	return destPath, destFormat, nil
}
