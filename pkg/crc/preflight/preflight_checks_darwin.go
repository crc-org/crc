package preflight

import (
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	neturl "net/url"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	dl "github.com/code-ready/crc/pkg/download"
	crcos "github.com/code-ready/crc/pkg/os"
)

const (
	virtualBoxDownloadURL   = "https://download.virtualbox.org/virtualbox/6.0.4/VirtualBox-6.0.4-128413-OSX.dmg"
	virtualBoxMountLocation = "/Volumes/VirtualBox"

	hyperkitDriverCommand = "crc-driver-hyperkit"
	hyperkitDriverVersion = "0.12.5"

	resolverDir  = "/etc/resolver"
	resolverFile = "/etc/resolver/testing"
	hostFile     = "/etc/hosts"
)

var (
	virtualBoxPkgLocation     = fmt.Sprintf("%s/VirtualBox.pkg", virtualBoxMountLocation)
	hyperkitDownloadURL       = fmt.Sprintf("https://github.com/code-ready/machine-driver-hyperkit/releases/download/v%s/hyperkit", hyperkitDriverVersion)
	hyperkitDriverDownloadURL = fmt.Sprintf("https://github.com/code-ready/machine-driver-hyperkit/releases/download/v%s/crc-driver-hyperkit", hyperkitDriverVersion)
)

// Add darwin specific checks
func checkVirtualBoxInstalled() (bool, error) {
	logging.Debug("Checking if VirtualBox is installed")
	_, err := exec.LookPath("VBoxManage")
	if err != nil {
		return false, errors.New("VirtualBox cli VBoxManage is not found in the path")
	}
	logging.Debug("VirtualBox was found")
	return true, nil
}

func fixVirtualBoxInstallation() (bool, error) {
	logging.Debug("Downloading VirtualBox")
	// Download the virtualbox installer in ~/.crc/cache
	tempFilePath := filepath.Join(constants.MachineCacheDir, "virtualbox.dmg")
	_, err := dl.Download(virtualBoxDownloadURL, tempFilePath, 0600)
	if err != nil {
		return false, err
	}
	defer os.Remove(tempFilePath)
	logging.Debug("Installing VirtualBox")
	stdOut, stdErr, err := crcos.RunWithPrivilege("mount VirtualBox disk image", "hdiutil", "attach", tempFilePath)
	if err != nil {
		return false, fmt.Errorf("Could not mount the virtualbox.dmg file: %s %v: %s", stdOut, err, stdErr)
	}
	stdOut, stdErr, err = crcos.RunWithPrivilege("run VirtualBox installation", "installer", "-package", virtualBoxPkgLocation, "-target", "/")
	if err != nil {
		return false, fmt.Errorf("Could not install VirtualBox.pkg: %s %v: %s", stdOut, err, stdErr)
	}
	stdOut, stdErr, err = crcos.RunWithPrivilege("unmount VirtualBox disk image", "hdiutil", "detach", virtualBoxMountLocation)
	if err != nil {
		return false, fmt.Errorf("Could not install VirtualBox.pkg: %s %v: %s", stdOut, err, stdErr)
	}
	logging.Debug("VirtualBox installed")
	return true, nil
}

func tryRemoveDestFile(url string, destDir string) error {
	u, err := neturl.Parse(url)
	if err != nil {
		return fmt.Errorf("Cannot parse URL %s", url)
	}

	urlPath, err := neturl.PathUnescape(u.EscapedPath())
	if err != nil {
		return fmt.Errorf("Cannot unescape URL path %s", urlPath)
	}
	destFilename := path.Base(urlPath)
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

	filename, err := dl.Download(url, destDir, 0644)
	if err != nil {
		return "", err
	}

	err = os.Chmod(filename, mode)
	if err != nil {
		os.Remove(filename)
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

func checkSuid(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if fi.Mode()&os.ModeSetuid == 0 {
		return false, fmt.Errorf("%s does not have the SUID bit set (%s)", path, fi.Mode().String())
	}
	if fi.Sys().(*syscall.Stat_t).Uid != 0 {
		return false, fmt.Errorf("%s is not owned by root", path)
	}

	return true, nil
}

func checkHyperKitInstalled() (bool, error) {
	logging.Debugf("Checking if hyperkit is installed")
	hyperkitPath := filepath.Join(constants.CrcBinDir, "hyperkit")
	err := unix.Access(hyperkitPath, unix.X_OK)
	if err != nil {
		logging.Debugf("%s not executable", hyperkitPath)
		return false, err
	}

	return checkSuid(hyperkitPath)
}

func fixHyperKitInstallation() (bool, error) {
	hyperkitFile, err := download(hyperkitDownloadURL, constants.CrcBinDir, 0755)
	if err != nil {
		return false, err
	}

	err = setSuid(hyperkitFile)
	if err != nil {
		return false, err
	}

	return true, nil
}

func checkMachineDriverHyperKitInstalled() (bool, error) {
	logging.Debugf("Checking if %s is installed", hyperkitDriverCommand)
	hyperkitPath := filepath.Join(constants.CrcBinDir, hyperkitDriverCommand)
	err := unix.Access(hyperkitPath, unix.X_OK)
	if err != nil {
		return false, err
	}

	// Check the version of driver if it matches to supported one
	stdOut, stdErr, err := crcos.RunWithDefaultLocale(hyperkitPath, "version")
	if err != nil {
		return false, err
	}
	if !strings.Contains(stdOut, hyperkitDriverVersion) {
		return false, fmt.Errorf("%s does not have right version \n Required: %s \n Got: %s use 'crc setup' command.\n %v\n", hyperkitDriverCommand, hyperkitDriverVersion, stdOut, stdErr)
	}
	logging.Debugf("%s is already installed in %s", hyperkitDriverCommand, hyperkitPath)

	return checkSuid(hyperkitPath)
}

func fixMachineDriverHyperKitInstalled() (bool, error) {
	hyperkitDriverPath, err := download(hyperkitDriverDownloadURL, constants.CrcBinDir, 0755)
	if err != nil {
		return false, err
	}

	err = setSuid(hyperkitDriverPath)
	if err != nil {
		return false, err
	}

	return true, nil
}

func checkResolverFilePermissions() (bool, error) {
	return isUserHaveFileWritePermission(resolverFile)
}

func fixResolverFilePermissions() (bool, error) {
	// Check if resolver directory available or not
	if _, err := os.Stat(resolverDir); os.IsNotExist(err) {
		logging.Debugf("Creating %s directory", resolverDir)
		stdOut, stdErr, err := crcos.RunWithPrivilege(fmt.Sprintf("create dir %s", resolverDir), "mkdir", resolverDir)
		if err != nil {
			return false, fmt.Errorf("Unable to create the resolver Dir: %s %v: %s", stdOut, err, stdErr)
		}
	}
	logging.Debugf("Making %s readable/writable by the current user", resolverFile)
	stdOut, stdErr, err := crcos.RunWithPrivilege(fmt.Sprintf("create file %s", resolverFile), "touch", resolverFile)
	if err != nil {
		return false, fmt.Errorf("Unable to create the resolver file: %s %v: %s", stdOut, err, stdErr)
	}

	return addFileWritePermissionToUser(resolverFile)
}

func checkHostsFilePermissions() (bool, error) {
	return isUserHaveFileWritePermission(hostFile)
}

func fixHostsFilePermissions() (bool, error) {
	return addFileWritePermissionToUser(hostFile)
}

func isUserHaveFileWritePermission(filename string) (bool, error) {
	err := unix.Access(filename, unix.R_OK|unix.W_OK)
	if err != nil {
		return false, fmt.Errorf("%s is not readable/writable by the current user", filename)
	}
	return true, nil
}

func addFileWritePermissionToUser(filename string) (bool, error) {
	logging.Debugf("Making %s readable/writable by the current user", filename)
	currentUser, err := user.Current()
	if err != nil {
		return false, err
	}

	stdOut, stdErr, err := crcos.RunWithPrivilege(fmt.Sprintf("change ownership of %s", filename), "chown", currentUser.Username, filename)
	if err != nil {
		return false, fmt.Errorf("Unable to change ownership of the filename: %s %v: %s", stdOut, err, stdErr)
	}

	err = os.Chmod(filename, 0500)
	if err != nil {
		return false, fmt.Errorf("Unable to change permissions of the filename: %s %v: %s", stdOut, err, stdErr)
	}
	logging.Debugf("%s is readable/writable by current user", filename)

	return true, nil
}
