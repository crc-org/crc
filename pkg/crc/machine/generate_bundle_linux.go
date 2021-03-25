package machine

import (
	"os"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	crcos "github.com/code-ready/crc/pkg/os"
)

func copyDiskImage(destDir string, srcDir string) (string, string, error) {
	tmpDir := filepath.Join(constants.MachineCacheDir, "image-rebase")
	_ = os.RemoveAll(tmpDir) // clean up before using it
	if err := os.Mkdir(tmpDir, 0700); err != nil {
		return "", "", err
	}
	defer func() {
		_ = os.RemoveAll(tmpDir) // clean up after using it
	}()

	// We copy the overlay file because before doing the commit, we need to adjust the backing file.
	baseImageName := filepath.Base(srcDir)
	srcOverlayDiskPath := filepath.Join(constants.MachineInstanceDir, constants.DefaultName, baseImageName)
	destOverlayDiskPath := filepath.Join(tmpDir, "overlayImage.qcow2")
	if err := crcos.CopyFileContents(srcOverlayDiskPath, destOverlayDiskPath, 0644); err != nil {
		return "", "", err
	}

	// We copy the base image because this is where we are going to commit the changes which were made to the VM
	if err := crcos.CopyFileContents(srcDir,
		filepath.Join(tmpDir, baseImageName), 0644); err != nil {
		return "", "", err
	}

	// Use qemu-img commands to rebase and commit it.
	format := "qcow2"
	if _, _, err := crcos.RunWithDefaultLocale("qemu-img", "rebase", "-F", format, "-b",
		filepath.Join(tmpDir, baseImageName), destOverlayDiskPath); err != nil {
		return "", "", err
	}
	if _, _, err := crcos.RunWithDefaultLocale("qemu-img", "commit", destOverlayDiskPath); err != nil {
		return "", "", err
	}

	// Copy the final image to destination dir
	if err := crcos.CopyFileContents(filepath.Join(tmpDir, baseImageName), filepath.Join(destDir, baseImageName), 0644); err != nil {
		return "", "", err
	}
	return filepath.Join(destDir, baseImageName), format, nil
}
