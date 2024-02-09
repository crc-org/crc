package preflight

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/cache"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine/vfkit"
	"github.com/crc-org/crc/v2/pkg/crc/version"
	crcos "github.com/crc-org/crc/v2/pkg/os"
	"github.com/crc-org/crc/v2/pkg/os/darwin"
	"github.com/crc-org/crc/v2/pkg/os/darwin/launchd"
	"github.com/klauspost/cpuid/v2"
	"golang.org/x/sys/unix"
)

const (
	resolverDir  = "/etc/resolver"
	resolverFile = "/etc/resolver/testing"
)

func checkAppleSilicon() error {
	if strings.HasPrefix(cpuid.CPU.BrandName, "VirtualApple") {
		logging.Debugf("Running with an emulated x86_64 CPU")
		return fmt.Errorf("This version of CRC for AMD64/Intel64 CPUs is unsupported on Apple M1 hardware")
	}

	return nil
}

func checkVfkitInstalled() error {
	if version.IsInstaller() {
		return nil
	}

	h := cache.NewVfkitCache()
	if !h.IsCached() {
		return fmt.Errorf("%s executable is not cached", h.GetExecutableName())
	}
	vfkitPath := h.GetExecutablePath()
	err := unix.Access(vfkitPath, unix.X_OK)
	if err != nil {
		return fmt.Errorf("%s not executable", vfkitPath)
	}
	return h.CheckVersion()
}

func fixVfkitInstallation() error {
	if version.IsInstaller() {
		return nil
	}

	h := cache.NewVfkitCache()

	logging.Debugf("Installing %s", h.GetExecutableName())

	if err := h.EnsureIsCached(); err != nil {
		return fmt.Errorf("Unable to download %s : %v", h.GetExecutableName(), err)
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
		stdOut, stdErr, err := crcos.RunPrivileged(fmt.Sprintf("Creating dir %s", resolverDir), "mkdir", resolverDir)
		if err != nil {
			return fmt.Errorf("Unable to create the resolver Dir: %s %v: %s", stdOut, err, stdErr)
		}
	}
	logging.Debugf("Making %s readable/writable by the current user", resolverFile)
	stdOut, stdErr, err := crcos.RunPrivileged(fmt.Sprintf("Creating file %s", resolverFile), "touch", resolverFile)
	if err != nil {
		return fmt.Errorf("Unable to create the resolver file: %s %v: %s", stdOut, err, stdErr)
	}

	return addFileWritePermissionToUser(resolverFile)
}

func removeResolverFile() error {
	// Check if the resolver file exist or not
	if _, err := os.Stat(resolverFile); !os.IsNotExist(err) {
		logging.Debugf("Removing %s file", resolverFile)
		err := crcos.RemoveFileAsRoot(fmt.Sprintf("Removing file %s", resolverFile), resolverFile)
		if err != nil {
			return fmt.Errorf("Unable to delete the resolver File: %s %v", resolverFile, err)
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

	stdOut, stdErr, err := crcos.RunPrivileged(fmt.Sprintf("Changing ownership of %s", filename), "chown", currentUser.Username, filename)
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

func killVfkitProcess() error {
	pgrepPath, err := exec.LookPath("pgrep")
	if err != nil {
		return fmt.Errorf("Could not find 'pgrep'. %w", err)
	}
	if _, _, err := crcos.RunWithDefaultLocale(pgrepPath, "-f", vfkit.ExecutablePath()); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			/* 1: no processes matched */
			if exitErr.ExitCode() == 1 {
				logging.Debugf("No running 'vfkit' process started by crc")
				return nil
			}
		}
		logging.Debugf("Failed to find 'vfkit' process. %v", err)
		/* Unclear what pgrep failure was, don't return, maybe pkill will be more successful */
	}

	pkillPath, err := exec.LookPath("pkill")
	if err != nil {
		return fmt.Errorf("Could not find 'pkill'. %w", err)
	}
	if _, _, err := crcos.RunWithDefaultLocale(pkillPath, "-SIGKILL", "-f", vfkit.ExecutablePath()); err != nil {
		return fmt.Errorf("Failed to kill 'vfkit' process. %w", err)
	}
	return nil
}

func getDaemonConfig() (*launchd.AgentConfig, error) {
	logFilePath := filepath.Join(constants.CrcBaseDir, ".launchd-crcd.log")

	env := map[string]string{"Version": version.GetCRCVersion()}
	daemonConfig := launchd.AgentConfig{
		Label:          constants.DaemonAgentLabel,
		ExecutablePath: constants.CrcSymlinkPath,
		StdOutFilePath: logFilePath,
		StdErrFilePath: logFilePath,
		Args:           []string{"daemon", "--log-level=debug"},
		Env:            env,
	}

	return &daemonConfig, nil
}

func checkIfDaemonPlistFileExists() error {
	daemonConfig, err := getDaemonConfig()
	if err != nil {
		return err
	}
	if err := launchd.CheckPlist(*daemonConfig); err != nil {
		return err
	}
	if !launchd.AgentRunning(daemonConfig.Label) && !daemonRunning() {
		return fmt.Errorf("launchd agent '%s' is not running", daemonConfig.Label)
	}
	return nil
}

func fixDaemonPlistFileExists() error {
	if daemonRunning() {
		if err := killDaemonProcess(); err != nil {
			return err
		}
	}
	daemonConfig, err := getDaemonConfig()
	if err != nil {
		return err
	}
	return fixPlistFileExists(*daemonConfig)
}

func removeDaemonPlistFile() error {
	if err := launchd.UnloadPlist(constants.DaemonAgentLabel); err != nil {
		return err
	}
	return launchd.RemovePlist(constants.DaemonAgentLabel)
}

func fixPlistFileExists(agentConfig launchd.AgentConfig) error {
	logging.Debugf("Creating plist for %s", agentConfig.Label)
	err := launchd.CreatePlist(agentConfig)
	if err != nil {
		return err
	}
	if err := launchd.LoadPlist(agentConfig.Label); err != nil {
		logging.Debugf("failed to load launchd agent '%s': %v", agentConfig.Label, err.Error())
		return err
	}
	if err := launchd.RestartAgent(agentConfig.Label); err != nil {
		logging.Debugf("failed to restart launchd agent '%s': %v", agentConfig.Label, err.Error())
		return err
	}
	return waitForDaemonRunning()
}

func deprecationNotice() error {
	supports, err := darwin.AtLeast("13.0.0")
	if err != nil {
		return err
	}
	if !supports {
		logging.Warnf("This version of macOS is going to be unsupported for CRC, Please update to macOS 13 or newer")
	}
	return nil
}
