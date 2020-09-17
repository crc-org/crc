package preflight

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/cache"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/hyperkit"
	crcos "github.com/code-ready/crc/pkg/os"
	"golang.org/x/sys/unix"
)

const (
	resolverDir  = "/etc/resolver"
	resolverFile = "/etc/resolver/testing"
	hostsFile    = "/etc/hosts"
)

func checkHyperKitInstalled() error {
	h := cache.NewHyperkitCache()
	if !h.IsCached() {
		return fmt.Errorf("%s binary is not cached", hyperkit.HyperkitCommand)
	}
	hyperkitPath := filepath.Join(constants.CrcBinDir, hyperkit.HyperkitCommand)
	err := unix.Access(hyperkitPath, unix.X_OK)
	if err != nil {
		return fmt.Errorf("%s not executable", hyperkitPath)
	}
	return checkSuid(hyperkitPath)
}

func fixHyperKitInstallation() error {
	logging.Debugf("Installing %s", hyperkit.HyperkitCommand)
	h := cache.NewHyperkitCache()
	if err := h.EnsureIsCached(); err != nil {
		return fmt.Errorf("Unable to download %s : %v", hyperkit.HyperkitCommand, err)
	}
	return setSuid(filepath.Join(constants.CrcBinDir, hyperkit.HyperkitCommand))
}

func checkMachineDriverHyperKitInstalled() error {
	logging.Debugf("Checking if %s is installed", hyperkit.MachineDriverCommand)
	hyperkitDriver := cache.NewMachineDriverHyperkitCache()
	if !hyperkitDriver.IsCached() {
		return fmt.Errorf("%s binary is not cached", hyperkit.MachineDriverCommand)
	}

	if err := hyperkitDriver.CheckVersion(); err != nil {
		return err
	}
	return checkSuid(filepath.Join(constants.CrcBinDir, hyperkit.MachineDriverCommand))
}

func fixMachineDriverHyperKitInstalled() error {
	logging.Debugf("Installing %s", hyperkit.MachineDriverCommand)
	hyperkitDriver := cache.NewMachineDriverHyperkitCache()
	if err := hyperkitDriver.EnsureIsCached(); err != nil {
		return fmt.Errorf("Unable to download %s : %v", hyperkit.MachineDriverCommand, err)
	}
	return setSuid(filepath.Join(constants.CrcBinDir, hyperkit.MachineDriverCommand))
}

func checkEtcHostsFilePermissions() error {
	logging.Debugf("Checking if /etc/hosts ownership/permissions need to be adjusted after crc upgrade")
	fileinfo, err := os.Stat(hostsFile)
	if err != nil {
		return err
	}
	// Older crc releases were setting /etc/hosts permissions to 0600 and ownership to the current user
	// This will cause issues if ownership is reset to root:wheel with permissions
	// issue if other tools
	if fileinfo.Mode().Perm() == 0600 {
		return fmt.Errorf("%s permissions are not 0644", hostsFile)
	}
	return nil
}

func fixEtcHostsFilePermissions() error {
	stdOut, stdErr, err := crcos.RunWithPrivilege(fmt.Sprintf("change ownership of %s", hostsFile), "chown", "root:wheel", hostsFile)
	if err != nil {
		return fmt.Errorf("Unable to change ownership of %s: %s %v: %s", hostsFile, stdOut, err, stdErr)
	}

	stdOut, stdErr, err = crcos.RunWithPrivilege(fmt.Sprintf("change permissions of %s", hostsFile), "chmod", "644", hostsFile)
	if err != nil {
		return fmt.Errorf("Unable to change permissions of %s to 0644: %s %v: %s", hostsFile, stdOut, err, stdErr)
	}

	return nil
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

func removeResolverFile() error {
	// Check if the resolver file exist or not
	if _, err := os.Stat(resolverFile); !os.IsNotExist(err) {
		logging.Debugf("Removing %s file", resolverFile)
		_, stdErr, err := crcos.RunWithPrivilege(fmt.Sprintf("Remove file %s", resolverFile), "rm", "-f", resolverFile)
		if err != nil {
			return fmt.Errorf("Unable to delete the resolver File: %s %v: %s", resolverFile, err, stdErr)
		}
	}
	return nil
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
		logging.Debugf("user.Current() failed: %v", err)
		return fmt.Errorf("Failed to get current user id")
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
