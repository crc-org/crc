package preflight

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/cache"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"
	crcos "github.com/code-ready/crc/pkg/os"
	"golang.org/x/sys/unix"
)

const (
	resolverDir  = "/etc/resolver"
	resolverFile = "/etc/resolver/testing"
	hostsFile    = "/etc/hosts"
)

func checkHyperKitInstalled(networkMode network.Mode) func() error {
	return func() error {
		h := cache.NewHyperKitCache()
		if !h.IsCached() {
			return fmt.Errorf("%s executable is not cached", h.GetExecutableName())
		}
		hyperkitPath := h.GetExecutablePath()
		err := unix.Access(hyperkitPath, unix.X_OK)
		if err != nil {
			return fmt.Errorf("%s not executable", hyperkitPath)
		}
		if err := h.CheckVersion(); err != nil {
			return err
		}
		if networkMode == network.VSockMode {
			return nil
		}
		return checkSuid(hyperkitPath)
	}
}

func fixHyperKitInstallation(networkMode network.Mode) func() error {
	return func() error {
		h := cache.NewHyperKitCache()

		logging.Debugf("Installing %s", h.GetExecutableName())

		if err := h.EnsureIsCached(); err != nil {
			return fmt.Errorf("Unable to download %s : %v", h.GetExecutableName(), err)
		}
		if networkMode == network.VSockMode {
			return nil
		}
		return setSuid(h.GetExecutablePath())
	}
}

func checkMachineDriverHyperKitInstalled(networkMode network.Mode) func() error {
	return func() error {
		hyperkitDriver := cache.NewMachineDriverHyperKitCache()

		logging.Debugf("Checking if %s is installed", hyperkitDriver.GetExecutableName())
		if !hyperkitDriver.IsCached() {
			return fmt.Errorf("%s executable is not cached", hyperkitDriver.GetExecutableName())
		}

		if err := hyperkitDriver.CheckVersion(); err != nil {
			return err
		}
		if networkMode == network.VSockMode {
			return nil
		}
		return checkSuid(hyperkitDriver.GetExecutablePath())
	}
}

func fixMachineDriverHyperKitInstalled(networkMode network.Mode) func() error {
	return func() error {
		hyperkitDriver := cache.NewMachineDriverHyperKitCache()

		logging.Debugf("Installing %s", hyperkitDriver.GetExecutableName())

		if err := hyperkitDriver.EnsureIsCached(); err != nil {
			return fmt.Errorf("Unable to download %s : %v", hyperkitDriver.GetExecutableName(), err)
		}
		if networkMode == network.VSockMode {
			return nil
		}
		return setSuid(hyperkitDriver.GetExecutablePath())
	}
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

func stopCRCHyperkitProcess() error {
	pgrepPath, err := exec.LookPath("pgrep")
	if err != nil {
		return fmt.Errorf("Could not find 'pgrep'. %w", err)
	}
	if _, _, err := crcos.RunWithDefaultLocale(pgrepPath, "-f", filepath.Join(constants.CrcBinDir, "hyperkit")); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			/* 1: no processes matched */
			if exitErr.ExitCode() == 1 {
				logging.Debugf("No running 'hyperkit' process started by crc")
				return nil
			}
		}
		logging.Debugf("Failed to find 'hyperkit' process. %v", err)
		/* Unclear what pgrep failure was, don't return, maybe pkill will be more successful */
	}

	pkillPath, err := exec.LookPath("pkill")
	if err != nil {
		return fmt.Errorf("Could not find 'pkill'. %w", err)
	}
	if _, _, err := crcos.RunWithDefaultLocale(pkillPath, "-f", filepath.Join(constants.CrcBinDir, "hyperkit")); err != nil {
		return fmt.Errorf("Failed to kill 'hyperkit' process. %w", err)
	}
	return nil
}
