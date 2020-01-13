package preflight

import (
	"fmt"
	"golang.org/x/sys/unix"
	neturl "net/url"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/hyperkit"
	dl "github.com/code-ready/crc/pkg/download"
	crcos "github.com/code-ready/crc/pkg/os"
)

const (
	resolverDir  = "/etc/resolver"
	resolverFile = "/etc/resolver/testing"
	hostFile     = "/etc/hosts"
)

func basenameFromUrl(url string) (string, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return "", fmt.Errorf("Cannot parse URL %s", url)
	}

	urlPath, err := neturl.PathUnescape(u.EscapedPath())
	if err != nil {
		return "", fmt.Errorf("Cannot unescape URL path %s", urlPath)
	}

	return path.Base(urlPath), nil
}

// Add darwin specific checks
func tryRemoveDestFile(url string, destDir string) error {
	destFilename, err := basenameFromUrl(url)
	if err != nil {
		return err
	}
	destPath := filepath.Join(destDir, destFilename)
	err = os.Remove(destPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Could not remove %s: %v", destPath, err)
	}

	return nil
}

func download(url string, destDir string, mode os.FileMode) (string, error) {
	err := os.MkdirAll(destDir, 0111|mode)
	if err != nil && !os.IsExist(err) {
		return "", fmt.Errorf("Cannot create directory %s", destDir)
	}

	// If the destination file already exists, dl.Download may not be able to
	// overwrite it if we made it suid. We can however delete it beforehand.
	err = tryRemoveDestFile(url, destDir)
	if err != nil {
		return "", err
	}

	filename, err := dl.Download(url, destDir, mode)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func setSuid(path string) error {
	logging.Debugf("Making %s suid", path)

	stdOut, stdErr, err := crcos.RunWithPrivilege(fmt.Sprintf("change ownership of %s", path), "chown", "root:wheel", path)
	if err != nil {
		return fmt.Errorf("Unable to set ownership of %s to root:wheel: %s %v: %s",
			path, stdOut, err, stdErr)
	}

	/* Can't do this before the chown as the chown will reset the suid bit */
	stdOut, stdErr, err = crcos.RunWithPrivilege(fmt.Sprintf("set suid for %s", path), "chmod", "u+s", path)
	if err != nil {
		return fmt.Errorf("Unable to set suid bit on %s: %s %v: %s", path, stdOut, err, stdErr)
	}
	return nil
}

func checkSuid(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSetuid == 0 {
		return fmt.Errorf("%s does not have the SUID bit set (%s)", path, fi.Mode().String())
	}
	if fi.Sys().(*syscall.Stat_t).Uid != 0 {
		return fmt.Errorf("%s is not owned by root", path)
	}

	return nil
}

func checkHyperKitInstalled() error {
	logging.Debugf("Checking if hyperkit is installed")
	hyperkitPath := filepath.Join(constants.CrcBinDir, "hyperkit")
	err := unix.Access(hyperkitPath, unix.X_OK)
	if err != nil {
		logging.Debugf("%s not executable", hyperkitPath)
		return err
	}

	return checkSuid(hyperkitPath)
}

func extractOrDownloadBinary(url string) error {
	binaryName, err := basenameFromUrl(url)
	if err != nil {
		return err
	}
	logging.Debugf("Installing %s", binaryName)
	binaryPath, err := extractBinary(binaryName, 0755)
	if err != nil {
		binaryPath, err = download(url, constants.CrcBinDir, 0755)
		if err != nil {
			return err
		}
	}

	return setSuid(binaryPath)
}

func fixHyperKitInstallation() error {
	return extractOrDownloadBinary(hyperkit.HyperkitDownloadUrl)
}

func checkMachineDriverHyperKitInstalled() error {
	logging.Debugf("Checking if %s is installed", hyperkit.MachineDriverCommand)
	hyperkitPath := filepath.Join(constants.CrcBinDir, hyperkit.MachineDriverCommand)
	err := unix.Access(hyperkitPath, unix.X_OK)
	if err != nil {
		return err
	}

	// Check the version of driver if it matches to supported one
	stdOut, stdErr, err := crcos.RunWithDefaultLocale(hyperkitPath, "version")
	if err != nil {
		return err
	}
	if !strings.Contains(stdOut, hyperkit.MachineDriverVersion) {
		return fmt.Errorf("%s does not have right version \n Required: %s \n Got: %s use 'crc setup' command.\n %v\n", hyperkit.MachineDriverCommand, hyperkit.MachineDriverVersion, stdOut, stdErr)
	}
	logging.Debugf("%s is already installed in %s", hyperkit.MachineDriverCommand, hyperkitPath)

	return checkSuid(hyperkitPath)
}

func fixMachineDriverHyperKitInstalled() error {
	return extractOrDownloadBinary(hyperkit.MachineDriverDownloadUrl)
}

func checkResolverFilePermissions() error {
	return isUserHaveFileWritePermission(resolverFile)
}

func fixResolverFilePermissions() error {
	// Check if resolver directory available or not
	if _, err := os.Stat(resolverDir); os.IsNotExist(err) {
		logging.Debugf("Creating %s directory", resolverDir)
		stdOut, stdErr, err := crcos.RunWithPrivilege(fmt.Sprintf("create dir %s", resolverDir), "mkdir", resolverDir)
		if err != nil {
			return fmt.Errorf("Unable to create the resolver Dir: %s %v: %s", stdOut, err, stdErr)
		}
	}
	logging.Debugf("Making %s readable/writable by the current user", resolverFile)
	stdOut, stdErr, err := crcos.RunWithPrivilege(fmt.Sprintf("create file %s", resolverFile), "touch", resolverFile)
	if err != nil {
		return fmt.Errorf("Unable to create the resolver file: %s %v: %s", stdOut, err, stdErr)
	}

	return addFileWritePermissionToUser(resolverFile)
}

func checkHostsFilePermissions() error {
	return isUserHaveFileWritePermission(hostFile)
}

func fixHostsFilePermissions() error {
	return addFileWritePermissionToUser(hostFile)
}

func isUserHaveFileWritePermission(filename string) error {
	err := unix.Access(filename, unix.R_OK|unix.W_OK)
	if err != nil {
		return fmt.Errorf("%s is not readable/writable by the current user", filename)
	}
	return nil
}

func addFileWritePermissionToUser(filename string) error {
	logging.Debugf("Making %s readable/writable by the current user", filename)
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	stdOut, stdErr, err := crcos.RunWithPrivilege(fmt.Sprintf("change ownership of %s", filename), "chown", currentUser.Username, filename)
	if err != nil {
		return fmt.Errorf("Unable to change ownership of the filename: %s %v: %s", stdOut, err, stdErr)
	}

	err = os.Chmod(filename, 0600)
	if err != nil {
		return fmt.Errorf("Unable to change permissions of the filename: %s %v: %s", stdOut, err, stdErr)
	}
	logging.Debugf("%s is readable/writable by current user", filename)

	return nil
}
