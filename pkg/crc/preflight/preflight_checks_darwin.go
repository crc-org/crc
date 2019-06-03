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

	"github.com/code-ready/crc/pkg/crc/oc"
	"github.com/code-ready/crc/pkg/crc/logging"
	crcos "github.com/code-ready/crc/pkg/os"
)

const (
	virtualBoxDownloadURL   = "https://download.virtualbox.org/virtualbox/6.0.4/VirtualBox-6.0.4-128413-OSX.dmg"
	virtualBoxMountLocation = "/Volumes/VirtualBox"

	resolverFile = "/etc/resolver/testing"
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
	err := unix.Access(resolverFile, unix.R_OK | unix.W_OK)
	if err != nil {
		return false, fmt.Errorf("%s is not readable/writable by the current user", resolverFile)
	}

	return true, nil
}

func fixResolverFilePermissions() (bool, error) {
	logging.DebugF("Making %s readable/writable by the current user", resolverFile)
	stdOut, stdErr, err := crcos.RunWithPrivilege("touch", resolverFile)
	if err != nil {
		return false, fmt.Errorf("Unable to create the resolver file: %s %v: %s", stdOut, err, stdErr)
	}

	currentUser, err := user.Current()
	if err != nil {
		return false, err
	}

	stdOut, stdErr, err = crcos.RunWithPrivilege("chown", currentUser.Username, resolverFile)
	if err != nil {
		return false, fmt.Errorf("Unable to change ownership of the resolver file: %s %v: %s", stdOut, err, stdErr)
	}

	err = os.Chmod(resolverFile, 0644)
	if err != nil {
		return false, fmt.Errorf("Unable to change permissions of the resolver file: %s %v: %s", stdOut, err, stdErr)
	}
	logging.DebugF("%s is readable/writable by current user", resolverFile)

	return true, nil
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
