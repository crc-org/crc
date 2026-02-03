package preflight

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/daemonclient"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/systemd"
	"github.com/crc-org/crc/v2/pkg/crc/systemd/states"
	crcos "github.com/crc-org/crc/v2/pkg/os"
	"github.com/crc-org/crc/v2/pkg/os/linux"
)

// lookupQemuKvm finds the qemu-kvm binary.
// On RHEL, qemu-kvm is located at /usr/libexec/qemu-kvm which is not in PATH.
func lookupQemuKvm() (string, error) {
	// First try to find qemu-kvm in PATH
	if path, err := exec.LookPath("qemu-kvm"); err == nil {
		return path, nil
	}

	// On RHEL, qemu-kvm is in /usr/libexec/ which is not in PATH
	rhelPath := "/usr/libexec/qemu-kvm"
	if _, err := os.Stat(rhelPath); err == nil {
		return rhelPath, nil
	}

	return "", fmt.Errorf("qemu-kvm not found in PATH or at %s", rhelPath)
}

func checkRunningInsideWSL2() error {
	version, err := os.ReadFile("/proc/version")
	if err != nil {
		return err
	}

	if strings.Contains(strings.ToLower(string(version)), "microsoft") {
		logging.Debugf("Running inside WSL2 environment")
		return fmt.Errorf("CRC is unsupported using WSL2")
	}

	return nil
}

func checkVirtualizationEnabled() error {
	if runtime.GOARCH == "arm64" {
		logging.Debug("Ignoring virtualization check for arm64")
		return nil
	}
	logging.Debug("Checking if the vmx/svm flags are present in /proc/cpuinfo")
	// Check if the cpu flags vmx or svm is present
	flags, err := getCPUFlags()
	if err != nil {
		return err
	}

	re := regexp.MustCompile(`(vmx|svm)`)

	cputype := re.FindString(flags)
	if cputype == "" {
		return fmt.Errorf("Virtualization is not available for your CPU")
	}
	logging.Debug("CPU virtualization flags are good")
	return nil
}

func fixVirtualizationEnabled() error {
	return fmt.Errorf("You need to enable virtualization in BIOS")
}

func checkKvmEnabled() error {
	logging.Debug("Checking if /dev/kvm exists")
	// Check if /dev/kvm exists
	if _, err := os.Stat("/dev/kvm"); os.IsNotExist(err) {
		return fmt.Errorf("kvm kernel module is not loaded")
	}
	logging.Debug("/dev/kvm was found")
	return nil
}

func fixKvmEnabled() error {
	logging.Debug("Trying to load kvm module")
	flags, err := getCPUFlags()
	if err != nil {
		return err
	}

	switch {
	case strings.Contains(flags, "vmx"):
		stdOut, stdErr, err := crcos.RunPrivileged("Loading kvm_intel kernel module", "modprobe", "kvm_intel")
		if err != nil {
			return fmt.Errorf("Failed to load kvm intel module: %s %v: %s", stdOut, err, stdErr)
		}
	case strings.Contains(flags, "svm"):
		stdOut, stdErr, err := crcos.RunPrivileged("Loading kvm_amd kernel module", "modprobe", "kvm_amd")
		if err != nil {
			return fmt.Errorf("Failed to load kvm amd module: %s %v: %s", stdOut, err, stdErr)
		}
	default:
		logging.Debug("Unable to detect processor details")
	}

	logging.Debug("kvm module loaded")
	return nil
}

func qemuSystemBinary() string {
	switch runtime.GOARCH {
	case "arm64":
		return "qemu-system-aarch64"
	default:
		return "qemu-system-x86_64"
	}
}

func checkQemuKvmInstalled() error {
	qemuBinary := qemuSystemBinary()
	logging.Debugf("Checking if '%s' is available", qemuBinary)

	// First check in CrcBinDir where we create the symlink (not in PATH)
	crcBinPath := filepath.Join(constants.CrcBinDir, qemuBinary)
	if _, err := os.Stat(crcBinPath); err == nil {
		logging.Debugf("'%s' was found in %s", qemuBinary, crcBinPath)
		return nil
	}

	// Fall back to checking in PATH
	path, err := exec.LookPath(qemuBinary)
	if err != nil {
		return fmt.Errorf("%s was not found in path", qemuBinary)
	}
	logging.Debugf("'%s' was found in %s", qemuBinary, path)

	return nil
}

func fixQemuKvmInstalled(distro *linux.OsRelease) func() error {
	return func() error {
		qemuBinary := qemuSystemBinary()

		// First check if qemu-kvm exists and we can create a symlink
		qemuKvmPath, err := lookupQemuKvm()
		if err == nil {
			logging.Debugf("'qemu-kvm' found at %s, creating symlink for %s", qemuKvmPath, qemuBinary)
			symlinkPath := filepath.Join(constants.CrcBinDir, qemuBinary)

			// Ensure CrcBinDir exists
			if err := os.MkdirAll(constants.CrcBinDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", constants.CrcBinDir, err)
			}

			// Remove existing symlink if present
			_ = os.Remove(symlinkPath)

			// Create symlink: qemu-system-* -> qemu-kvm
			if err := os.Symlink(qemuKvmPath, symlinkPath); err != nil {
				return fmt.Errorf("failed to create symlink %s -> %s: %v", symlinkPath, qemuKvmPath, err)
			}
			logging.Debugf("Created symlink %s -> %s", symlinkPath, qemuKvmPath)
			return nil
		}

		// qemu-kvm not found, try to install it via package manager
		logging.Debug("Trying to install qemu-kvm")
		stdOut, stdErr, err := crcos.RunPrivileged("Installing qemu-kvm", "/bin/sh", "-c", installQemuKvmCommand(distro))
		if err != nil {
			return fmt.Errorf("Could not install qemu-kvm: %s %v: %s", stdOut, err, stdErr)
		}
		logging.Debug("qemu-kvm was successfully installed")

		// After installation, check again and create symlink if needed
		qemuKvmPath, err = lookupQemuKvm()
		if err == nil {
			symlinkPath := filepath.Join(constants.CrcBinDir, qemuBinary)
			if err := os.MkdirAll(constants.CrcBinDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", constants.CrcBinDir, err)
			}
			_ = os.Remove(symlinkPath)
			if err := os.Symlink(qemuKvmPath, symlinkPath); err != nil {
				return fmt.Errorf("failed to create symlink %s -> %s: %v", symlinkPath, qemuKvmPath, err)
			}
			logging.Debugf("Created symlink %s -> %s", symlinkPath, qemuKvmPath)
		}

		return nil
	}
}

func removeQemuKvmSymlink() error {
	qemuBinary := qemuSystemBinary()
	symlinkPath := filepath.Join(constants.CrcBinDir, qemuBinary)
	if err := os.Remove(symlinkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove symlink %s: %v", symlinkPath, err)
	}
	logging.Debugf("Removed symlink %s", symlinkPath)
	return nil
}

func installQemuKvmCommand(distro *linux.OsRelease) string {
	dnfCommand := "dnf install -y /usr/bin/qemu-kvm"
	switch {
	case distroIsLike(distro, linux.Ubuntu):
		return "apt-get update && apt-get install -y qemu-kvm"
	case distroIsLike(distro, linux.Fedora):
		return dnfCommand
	default:
		logging.Warnf("unsupported distribution %s, trying to install qemu-kvm with dnf", distro)
		return dnfCommand
	}
}

func systemdUnitRunning(sd *systemd.Commander, unitName string) bool {
	status, err := sd.Status(unitName)
	if err != nil {
		logging.Debugf("Could not get %s  status: %v", unitName, err)
		return false
	}
	switch status {
	case states.Running:
		logging.Debugf("%s is running", unitName)
		return true
	case states.Listening:
		logging.Debugf("%s is listening", unitName)
		return true
	default:
		logging.Debugf("%s is neither running nor listening", unitName)
		return false
	}
}

const (
	vsockUnitName     = "crc-vsock.socket"
	vsockUnitTemplate = `[Unit]
Description=CRC vsock socket

[Socket]
ListenStream=vsock::%d
Service=crc-daemon.service

[Install]
WantedBy=default.target
`

	httpUnitName = "crc-http.socket"
	httpUnit     = `[Unit]
Description=CRC HTTP socket

[Socket]
ListenStream=%h/.crc/crc-http.sock
Service=crc-daemon.service

[Install]
WantedBy=default.target
`

	daemonUnitName     = "crc-daemon.service"
	daemonUnitTemplate = `
[Unit]
Description=CRC daemon
Requires=crc-http.socket
Requires=crc-vsock.socket

[Service]
# This allows systemd to know when startup is not complete (for example, because of a preflight failure)
# daemon.SdNotify(false, daemon.SdNotifyReady) must be called before the startup is successful
Type=notify
ExecStart=%s daemon
`
)

var vsockUnit = fmt.Sprintf(vsockUnitTemplate, constants.DaemonVsockPort)

func checkSystemdUnit(unitName string, unitContent string, shouldBeRunning bool) error {
	sd := systemd.NewHostSystemdCommander().User()

	logging.Debugf("Checking if %s is running", unitName)
	running := systemdUnitRunning(sd, unitName)
	if !running && shouldBeRunning {
		return unitShouldBeRunningErr(unitName)
	} else if running && !shouldBeRunning {
		return unitShouldNotBeRunningErr(unitName)
	}

	logging.Debugf("Checking if %s has the expected content", unitName)
	unitPath := systemd.UserUnitPath(unitName)
	return crcos.FileContentMatches(unitPath, []byte(unitContent))
}

func daemonUnitContent() string {
	return fmt.Sprintf(daemonUnitTemplate, constants.CrcSymlinkPath)
}

func checkDaemonSystemdSockets() error {
	logging.Debug("Checking crc daemon systemd socket units")

	if err := checkSystemdUnit(httpUnitName, httpUnit, true); err != nil {
		return err
	}

	return checkSystemdUnit(vsockUnitName, vsockUnit, true)
}

func checkDaemonSystemdService() error {
	logging.Debug("Checking crc daemon systemd service")

	// the daemon should not be running at the end of setup, as it must be restarted on upgrades
	shouldNotBeRunningErr := checkSystemdUnit(daemonUnitName, daemonUnitContent(), false)
	if shouldNotBeRunningErr == nil {
		return nil
	}
	if !errors.Is(shouldNotBeRunningErr, unitShouldNotBeRunningErr(daemonUnitName)) {
		return shouldNotBeRunningErr
	}
	// daemon is running, check its version
	version, err := daemonclient.GetVersionFromDaemonAPI()
	if err != nil {
		return shouldNotBeRunningErr
	}
	return daemonclient.CheckVersionMismatch(version)
}

func fixSystemdUnit(unitName string, unitContent string, shouldBeRunning bool) error {
	logging.Debugf("Setting up %s", unitName)

	sd := systemd.NewHostSystemdCommander().User()

	if err := os.MkdirAll(systemd.UserUnitsDir(), 0750); err != nil {
		return err
	}
	unitPath := systemd.UserUnitPath(unitName)
	if crcos.FileContentMatches(unitPath, []byte(unitContent)) != nil {
		logging.Debugf("Creating %s", unitPath)
		if err := os.WriteFile(unitPath, []byte(unitContent), 0600); err != nil {
			return err
		}
		_ = sd.DaemonReload()
	}

	running := systemdUnitRunning(sd, unitName)
	if !running && shouldBeRunning {
		logging.Debugf("Starting %s", unitName)
		if err := sd.Enable(unitName); err != nil {
			return err
		}
		return sd.Start(unitName)
	} else if running && !shouldBeRunning {
		logging.Debugf("Stopping %s", unitName)
		return sd.Stop(unitName)
	}

	return nil
}

func fixDaemonSystemdSockets() error {
	logging.Debugf("Setting up crc daemon systemd socket units")
	if err := fixSystemdUnit(httpUnitName, httpUnit, true); err != nil {
		return err
	}

	return fixSystemdUnit(vsockUnitName, vsockUnit, true)
}

func fixDaemonSystemdService() error {
	logging.Debugf("Setting up crc daemon systemd unit")
	return fixSystemdUnit(daemonUnitName, daemonUnitContent(), false)
}

func removeDaemonSystemdSockets() error {
	logging.Debugf("Removing crc daemon systemd socket units")

	sd := systemd.NewHostSystemdCommander().User()

	_ = sd.Stop(httpUnitName)
	os.Remove(systemd.UserUnitPath(httpUnitName))

	_ = sd.Stop(vsockUnitName)
	os.Remove(systemd.UserUnitPath(vsockUnitName))

	return nil
}

func removeDaemonSystemdService() error {
	logging.Debugf("Removing crc daemon systemd service")

	sd := systemd.NewHostSystemdCommander().User()

	_ = sd.Stop(daemonUnitName)
	os.Remove(systemd.UserUnitPath(daemonUnitName))

	return nil
}

func warnNoDaemonAutostart() error {
	// only purpose of this check is to trigger a warning for RHEL7/CentOS7 users
	logging.Warnf("systemd --user is not available, crc daemon won't be autostarted and must be run manually before using CRC")
	return nil
}

func removeCrcVM() error {
	stdout, _, err := crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "domstate", constants.DefaultName)
	if err != nil {
		//  User may have run `crc delete` before `crc cleanup`
		//  in that case there is no crc vm so return early.
		return nil
	}
	if strings.TrimSpace(stdout) == "running" {
		_, stderr, err := crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "destroy", constants.DefaultName)
		if err != nil {
			logging.Debugf("%v : %s", err, stderr)
			return fmt.Errorf("Failed to destroy 'crc' VM")
		}
	}
	_, stderr, err := crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "undefine", "--nvram", constants.DefaultName)
	if err != nil {
		logging.Debugf("%v : %s", err, stderr)
		return fmt.Errorf("Failed to undefine 'crc' VM")
	}
	logging.Debug("'crc' VM is removed")
	return nil
}

func getCPUFlags() (string, error) {
	// Check if the cpu flags vmx or svm is present
	out, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		logging.Debugf("Failed to read /proc/cpuinfo: %v", err)
		return "", fmt.Errorf("Failed to read /proc/cpuinfo")
	}
	re := regexp.MustCompile(`flags.*:.*`)

	flags := re.FindString(string(out))
	if flags == "" {
		return "", fmt.Errorf("Could not find cpu flags from /proc/cpuinfo")
	}
	return flags, nil
}
