package machine

import (
	"fmt"
	"path/filepath"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	crcPreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	crcos "github.com/crc-org/crc/v2/pkg/os"
)

func copyDiskImage(destDir string, preset crcPreset.Preset) (string, string, error) {
	const destFormat = "qcow2"

	imageName := fmt.Sprintf("%s.qcow2", constants.InstanceName(preset))

	srcPath := filepath.Join(constants.MachineInstanceDir, constants.InstanceDirName(preset), imageName)
	destPath := filepath.Join(destDir, imageName)

	_, _, err := crcos.RunWithDefaultLocale("qemu-img", "convert", "-f", "qcow2", "-O", destFormat, srcPath, destPath)
	if err != nil {
		return "", "", err
	}

	return destPath, destFormat, nil
}
