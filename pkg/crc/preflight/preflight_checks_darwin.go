package preflight

import (
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
	crcos "github.com/code-ready/crc/pkg/os"
)

const (
	virtualBoxDownloadURL   = "https://download.virtualbox.org/virtualbox/6.0.4/VirtualBox-6.0.4-128413-OSX.dmg"
	virtualBoxMountLocation = "/Volumes/VirtualBox"

	resolverDir  = "/etc/resolver"
	resolverFile = "/etc/resolver/testing"
	resolvFile   = "/etc/resolv.conf"
)

var (
	virtualBoxPkgLocation = fmt.Sprintf("%s/VirtualBox.pkg", virtualBoxMountLocation)
)

// Add darwin specific checks
func checkVirtualBoxInstalled() (bool, error) {
	logging.Debug("Checking if VirtualBox is installed")
	path, err := exec.LookPath("VBoxManage")
	if err != nil {
		return false, errors.New("VirtualBox cli VBoxManage is not found in the path")
	}
	fi, _ := os.Stat(path)
	if fi.Mode()&os.ModeSymlink != 0 {
		path, err = os.Readlink(path)
		if err != nil {
			return false, errors.New("VirtualBox cli VBoxManage is not found in the path")
		}
	}
	logging.Debug("VirtualBox was found")
	return true, nil
}

func fixVirtualBoxInstallation() (bool, error) {
	logging.Debug("Downloading VirtualBox")
	// Download the driver binary in /tmp
	tempFilePath := filepath.Join(os.TempDir(), "virtualbox.dmg")
	out, err := os.Create(tempFilePath)
	if err != nil {
		return false, err
	}
	logging.Debug("Downloading from ", virtualBoxDownloadURL, " to ", tempFilePath)
	defer out.Close()
	resp, err := http.Get(virtualBoxDownloadURL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return false, err
	}

	logging.Debug("Installing VirtualBox")
	stdOut, stdErr, err := crcos.RunWithPrivilege("hdiutil", "attach", tempFilePath)
	if err != nil {
		return false, fmt.Errorf("Could not mount the virtualbox.dmg file: %s %v: %s", stdOut, err, stdErr)
	}
	stdOut, stdErr, err = crcos.RunWithPrivilege("installer", "-package", virtualBoxPkgLocation, "-target", "/")
	if err != nil {
		return false, fmt.Errorf("Could not install VirtualBox.pkg: %s %v: %s", stdOut, err, stdErr)
	}
	stdOut, stdErr, err = crcos.RunWithPrivilege("hdiutil", "detach", virtualBoxMountLocation)
	if err != nil {
		return false, fmt.Errorf("Could not install VirtualBox.pkg: %s %v: %s", stdOut, err, stdErr)
	}
	logging.Debug("VirtualBox installed")
	return true, nil
}

func checkResolverFilePermissions() (bool, error) {
	return isUserHaveFileWritePermission(resolverFile)
}

func fixResolverFilePermissions() (bool, error) {
	// Check if resolver directory available or not
	if _, err := os.Stat(resolverDir); os.IsNotExist(err) {
		logging.DebugF("Creating %s directory", resolverDir)
		stdOut, stdErr, err := crcos.RunWithPrivilege("mkdir", resolverDir)
		if err != nil {
			return false, fmt.Errorf("Unable to create the resolver Dir: %s %v: %s", stdOut, err, stdErr)
		}
	}
	logging.DebugF("Making %s readable/writable by the current user", resolverFile)
	stdOut, stdErr, err := crcos.RunWithPrivilege("touch", resolverFile)
	if err != nil {
		return false, fmt.Errorf("Unable to create the resolver file: %s %v: %s", stdOut, err, stdErr)
	}

	return addFileWritePermissionToUser(resolverFile)
}

// Check if oc binary is cached or not
func checkOcBinaryCached() (bool, error) {
	oc := oc.OcCached{}
	if !oc.IsCached() {
		return false, errors.New("oc binary is not cached.")
	}
	logging.Debug("oc binary already cached")
	return true, nil
}

func fixOcBinaryCached() (bool, error) {
	oc := oc.OcCached{}
	if err := oc.EnsureIsCached(); err != nil {
		return false, fmt.Errorf("Not able to download oc %v", err)
	}
	logging.Debug("oc binary cached")
	return true, nil
}

func checkResolvConfFilePermissions() (bool, error) {
	return isUserHaveFileWritePermission(resolvFile)
}

func fixResolvConfFilePermissions() (bool, error) {
	return addFileWritePermissionToUser(resolvFile)
}

func isUserHaveFileWritePermission(filename string) (bool, error) {
	err := unix.Access(filename, unix.R_OK|unix.W_OK)
	if err != nil {
		return false, fmt.Errorf("%s is not readable/writable by the current user", filename)
	}
	return true, nil
}

func addFileWritePermissionToUser(filename string) (bool, error) {
	logging.DebugF("Making %s readable/writable by the current user", filename)
	currentUser, err := user.Current()
	if err != nil {
		return false, err
	}

	stdOut, stdErr, err := crcos.RunWithPrivilege("chown", currentUser.Username, filename)
	if err != nil {
		return false, fmt.Errorf("Unable to change ownership of the filename: %s %v: %s", stdOut, err, stdErr)
	}

	err = os.Chmod(filename, 0644)
	if err != nil {
		return false, fmt.Errorf("Unable to change permissions of the filename: %s %v: %s", stdOut, err, stdErr)
	}
	logging.DebugF("%s is readable/writable by current user", filename)

	return true, nil
}
