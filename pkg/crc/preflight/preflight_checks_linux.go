package preflight

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/Masterminds/semver/v3"
	"github.com/code-ready/crc/pkg/crc/cache"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/libvirt"
	"github.com/code-ready/crc/pkg/crc/systemd"
	"github.com/code-ready/crc/pkg/crc/systemd/states"
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/code-ready/crc/pkg/os/linux"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

const (
	// This is defined in https://github.com/code-ready/machine-driver-libvirt/blob/master/go.mod#L5
	minSupportedLibvirtVersion = "3.4.0"
)

func checkRunningInsideWSL2() error {
	version, err := ioutil.ReadFile("/proc/version")
	if err != nil {
		return err
	}

	if strings.Contains(string(version), "Microsoft") {
		logging.Debugf("Running inside WSL2 environment")
		return fmt.Errorf("CodeReady Containers is unsupported using WSL2")
	}

	return nil
}

func checkVirtualizationEnabled() error {
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

func getLibvirtCapabilities() (*libvirtxml.Caps, error) {
	stdOut, _, err := crcos.RunWithDefaultLocale("virsh", "--readonly", "--connect", "qemu:///system", "capabilities")
	if err != nil {
		stdOut, _, err = crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///session", "capabilities")
		if err != nil {
			return nil, fmt.Errorf("Failed to run 'virsh capabilities': %v", err)
		}
	}
	caps := &libvirtxml.Caps{}
	err = caps.Unmarshal(stdOut)
	if err != nil {
		return nil, fmt.Errorf("Error parsing 'virsh capabilities': %v", err)
	}

	return caps, nil
}

func checkLibvirtInstalled() error {
	logging.Debug("Checking if 'virsh' is available")
	path, err := exec.LookPath("virsh")
	if err != nil {
		return fmt.Errorf("Libvirt cli virsh was not found in path")
	}
	logging.Debug("'virsh' was found in ", path)

	logging.Debug("Checking 'virsh capabilities' for libvirtd/qemu availability")
	caps, err := getLibvirtCapabilities()
	if err != nil {
		return err
	}

	foundHvm := false
	for _, guest := range caps.Guests {
		if guest.OSType == "hvm" && guest.Arch.Name == caps.Host.CPU.Arch {
			logging.Debugf("Found %s hypervisor with 'hvm' capabilities", caps.Host.CPU.Arch)
			foundHvm = true
			break
		}
	}
	if !foundHvm {
		return fmt.Errorf("Could not find a %s hypervisor with 'hvm' capabilities", caps.Host.CPU.Arch)
	}

	return nil
}

func fixLibvirtInstalled(distro *linux.OsRelease) func() error {
	return func() error {
		logging.Debug("Trying to install libvirt")
		stdOut, stdErr, err := crcos.RunPrivileged("Installing virtualization packages", "/bin/sh", "-c", installLibvirtCommand(distro))
		if err != nil {
			return fmt.Errorf("Could not install required packages: %s %v: %s", stdOut, err, stdErr)
		}
		logging.Debug("libvirt was successfully installed")
		return nil
	}
}

func installLibvirtCommand(distro *linux.OsRelease) string {
	yumCommand := "yum install -y libvirt libvirt-daemon-kvm qemu-kvm"
	switch {
	case distroIsLike(distro, linux.Ubuntu):
		return "apt-get update && apt-get install -y libvirt-daemon libvirt-daemon-system libvirt-clients"
	case distroIsLike(distro, linux.Fedora):
		return yumCommand
	default:
		logging.Warnf("unsupported distribution %s, trying to install libvirt with yum", distro)
		return yumCommand
	}
}

func checkLibvirtVersion() error {
	logging.Debugf("Checking if libvirt version is >=%s", minSupportedLibvirtVersion)
	stdOut, _, err := crcos.RunWithDefaultLocale("virsh", "-v")
	if err != nil {
		return fmt.Errorf("Failed to run virsh")
	}
	installedLibvirtVersion, err := semver.NewVersion(strings.TrimSpace(stdOut))
	if err != nil {
		return fmt.Errorf("Unable to parse installed libvirt version %v", err)
	}
	supportedLibvirtVersion, err := semver.NewVersion(minSupportedLibvirtVersion)
	if err != nil {
		return fmt.Errorf("Unable to parse %s libvirt version %v", minSupportedLibvirtVersion, err)
	}

	if installedLibvirtVersion.LessThan(supportedLibvirtVersion) {
		return fmt.Errorf("libvirt version %s is installed, but %s or higher is required", installedLibvirtVersion.String(), minSupportedLibvirtVersion)
	}

	return nil
}

func checkUserPartOfLibvirtGroup() error {
	logging.Debug("Checking if current user is part of the libvirt group")

	currentUser, err := user.Current()
	if err != nil {
		logging.Debugf("user.Current() failed: %v", err)
		return fmt.Errorf("Failed to get current user id")
	}
	gids, err := currentUser.GroupIds()
	if err != nil {
		return fmt.Errorf("Failed to get the groups user '%s' belongs to", currentUser.Username)
	}
	for _, gid := range gids {
		group, err := user.LookupGroupId(gid)
		if err != nil {
			logging.Debugf("Failed to lookup group id %s: %v", gid, err)
			continue
		}
		if group.Name == "libvirt" {
			logging.Debug("Current user is already in the libvirt group")
			return nil
		}
	}

	return fmt.Errorf("%s is not part of the libvirt group", currentUser.Username)
}

func fixUserPartOfLibvirtGroup() error {
	logging.Debug("Adding current user to the libvirt group")
	currentUser, err := user.Current()
	if err != nil {
		logging.Debugf("user.Current() failed: %v", err)
		return fmt.Errorf("Failed to get current user id")
	}
	_, _, err = crcos.RunPrivileged("Adding user to the libvirt group", "usermod", "-a", "-G", "libvirt", currentUser.Username)
	if err != nil {
		return fmt.Errorf("Failed to add user to libvirt group")
	}
	logging.Debug("Current user is in the libvirt group")

	return err
}

func checkCurrentGroups(distro *linux.OsRelease) func() error {
	return func() error {
		if !distroIsLike(distro, linux.Ubuntu) {
			return nil
		}

		// After adding the user to the libvirt group, they need to relogin for the new group to be used by the currrent shell
		gids, err := os.Getgroups()
		if err != nil {
			return err
		}
		for _, gid := range gids {
			group, err := user.LookupGroupId(fmt.Sprintf("%d", gid))
			if err != nil {
				logging.Debugf("Failed to lookup group id %d: %v", gid, err)
				continue
			}
			if group.Name == "libvirt" {
				logging.Debug("libvirt group is active for the current user/process")
				return nil
			}
		}
		return fmt.Errorf("User in the currently active process is not part of the libvirt group")
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
	vsockUnitName = "crc-vsock.socket"
	vsockUnit     = `[Unit]
Description=CodeReady Containers vsock socket

[Socket]
ListenStream=vsock::1024
Service=crc-daemon.service
`

	httpUnitName = "crc-http.socket"
	httpUnit     = `[Unit]
Description=CodeReady Containers HTTP socket

[Socket]
ListenStream=%h/.crc/crc-http.sock
Service=crc-daemon.service
`

	daemonUnitName     = "crc-daemon.service"
	daemonUnitTemplate = `
[Unit]
Description=CodeReady Containers daemon
Requires=crc-http.socket
Requires=crc-vsock.socket

[Service]
# This allows systemd to know when startup is not complete (for example, because of a preflight failure)
# daemon.SdNotify(false, daemon.SdNotifyReady) must be called before the startup is successful
Type=notify
ExecStart=%s daemon
`
)

func checkSystemdUnit(unitName string, unitContent string, shouldBeRunning bool) error {
	sd := systemd.NewHostSystemdCommander().User()

	logging.Debugf("Checking if %s is running", unitName)
	running := systemdUnitRunning(sd, unitName)
	if !running && shouldBeRunning {
		return fmt.Errorf("%s is not running", unitName)
	} else if running && !shouldBeRunning {
		return fmt.Errorf("%s is running", unitName)
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
	return checkSystemdUnit(daemonUnitName, daemonUnitContent(), false)
}

func fixSystemdUnit(unitName string, unitContent string, shouldBeRunning bool) error {
	logging.Debugf("Setting up %s", unitName)

	sd := systemd.NewHostSystemdCommander().User()

	unitPath := systemd.UserUnitPath(unitName)
	if crcos.FileContentMatches(unitPath, []byte(unitContent)) != nil {
		logging.Debugf("Creating %s", unitPath)
		if err := ioutil.WriteFile(unitPath, []byte(unitContent), 0600); err != nil {
			return err
		}
		_ = sd.DaemonReload()
	}

	running := systemdUnitRunning(sd, unitName)
	if !running && shouldBeRunning {
		logging.Debugf("Starting %s", unitName)
		return sd.Start(unitName)
	} else if running && !shouldBeRunning {
		logging.Debugf("Stopping %s", unitName)
		return sd.Stop(unitName)
	}

	return nil
}

func fixDaemonSystemdSockets() error {
	logging.Debugf("Setting up crc daemon systemd socket units")
	if err := os.MkdirAll(systemd.UserUnitsDir(), 0750); err != nil {
		return err
	}

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

func checkLibvirtServiceRunning() error {
	logging.Debug("Checking if libvirtd service is running")
	sd := systemd.NewHostSystemdCommander()

	libvirtSystemdUnits := []string{"virtqemud.socket", "libvirtd.socket", "virtqemud.service", "libvirtd.service"}
	for _, unit := range libvirtSystemdUnits {
		if systemdUnitRunning(sd, unit) {
			return nil
		}
	}

	logging.Warnf("No active (running) libvirtd systemd unit could be found - make sure one of libvirt systemd units is enabled so that it's autostarted at boot time.")
	return fmt.Errorf("found no active libvirtd systemd unit")
}

func fixLibvirtServiceRunning() error {
	logging.Debug("Starting libvirtd.service")
	sd := systemd.NewHostSystemdCommander()
	/* split libvirt daemon is a bit tricky to startup properly as we'd
	* need to start multiple components by hand, so we just start the
	* monolithic daemon
	 */
	err := sd.Start("libvirtd")
	if err != nil {
		return fmt.Errorf("Failed to start libvirt service")
	}
	logging.Debug("libvirtd.service is running")
	return nil
}

func checkMachineDriverLibvirtInstalled() error {
	machineDriverLibvirt := cache.NewMachineDriverLibvirtCache()

	logging.Debugf("Checking if %s is installed", machineDriverLibvirt.GetExecutableName())

	if !machineDriverLibvirt.IsCached() {
		return fmt.Errorf("%s executable is not cached", machineDriverLibvirt.GetExecutableName())
	}
	if err := machineDriverLibvirt.CheckVersion(); err != nil {
		return err
	}
	logging.Debugf("%s is already installed", machineDriverLibvirt.GetExecutableName())
	return nil
}

func fixMachineDriverLibvirtInstalled() error {
	machineDriverLibvirt := cache.NewMachineDriverLibvirtCache()

	logging.Debugf("Installing %s", machineDriverLibvirt.GetExecutableName())

	if err := machineDriverLibvirt.EnsureIsCached(); err != nil {
		return fmt.Errorf("Unable to download %s: %v", machineDriverLibvirt.GetExecutableName(), err)
	}
	logging.Debugf("%s is installed in %s", machineDriverLibvirt.GetExecutableName(), filepath.Dir(machineDriverLibvirt.GetExecutablePath()))
	return nil
}

func checkLibvirtCrcNetworkAvailable() error {
	logging.Debug("Checking if libvirt 'crc' network exists")
	_, _, err := crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "net-info", "crc")
	if err != nil {
		return fmt.Errorf("Libvirt network crc not found")
	}

	return checkLibvirtCrcNetworkDefinition()
}

func getLibvirtNetworkXML() (string, error) {
	config := libvirt.NetworkConfig{
		NetworkName: libvirt.DefaultNetwork,
		MAC:         libvirt.MACAddress,
		IP:          libvirt.IPAddress,
	}
	t, err := template.New("netxml").Parse(libvirt.NetworkTemplate)
	if err != nil {
		return "", err
	}
	var netXMLDef strings.Builder
	err = t.Execute(&netXMLDef, config)
	if err != nil {
		return "", err
	}

	return netXMLDef.String(), nil
}

func fixLibvirtCrcNetworkAvailable() error {
	logging.Debug("Creating libvirt 'crc' network")

	netXMLDef, err := getLibvirtNetworkXML()
	if err != nil {
		logging.Debugf("getLibvirtNetworkXML() failed: %v", err)
		return fmt.Errorf("Failed to read libvirt 'crc' network definition")
	}

	// For time being we are going to override the crc network according what we have in our binary template.
	// We also don't care about the error or output from those commands atm.
	// #nosec G204
	_, _, _ = crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "net-destroy", libvirt.DefaultNetwork)
	// #nosec G204
	_, _, _ = crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "net-undefine", libvirt.DefaultNetwork)
	// Create the network according to our defined template
	cmd := exec.Command("virsh", "--connect", "qemu:///system", "net-define", "/dev/stdin")
	cmd.Stdin = strings.NewReader(netXMLDef)
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		logging.Debugf("%v : %s", err, buf.String())
		return fmt.Errorf("Failed to create libvirt 'crc' network: %v - %s", err, buf.String())
	}
	logging.Debug("libvirt 'crc' network created")
	return nil
}

func removeLibvirtCrcNetwork() error {
	logging.Debug("Removing libvirt 'crc' network")
	_, _, err := crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "net-info", libvirt.DefaultNetwork)
	if err != nil {
		// Ignore if no crc network exists for libvirt
		// User may have manually deleted the `crc` network from libvirt
		return nil
	}
	_, stderr, err := crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "net-destroy", libvirt.DefaultNetwork)
	if err != nil {
		logging.Debugf("%v : %s", err, stderr)
		return fmt.Errorf("Failed to destroy libvirt 'crc' network")
	}

	_, stderr, err = crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "net-undefine", libvirt.DefaultNetwork)
	if err != nil {
		logging.Debugf("%v : %s", err, stderr)
		return fmt.Errorf("Failed to undefine libvirt 'crc' network")
	}
	logging.Debug("libvirt 'crc' network removed")
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
	_, stderr, err := crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "undefine", constants.DefaultName)
	if err != nil {
		logging.Debugf("%v : %s", err, stderr)
		return fmt.Errorf("Failed to undefine 'crc' VM")
	}
	logging.Debug("'crc' VM is removed")
	return nil
}

func trimSpacesFromXML(str string) string {
	strs := strings.Split(str, "\n")
	var builder strings.Builder
	for _, s := range strs {
		builder.WriteString(strings.TrimSpace(s))
	}

	return builder.String()
}

func checkLibvirtCrcNetworkDefinition() error {
	logging.Debug("Checking if libvirt 'crc' definition is up to date")
	stdOut, _, err := crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "net-dumpxml", "--inactive", "crc")
	if err != nil {
		return fmt.Errorf("Failed to get 'crc' network XML: %s", err)
	}
	stdOut = trimSpacesFromXML(stdOut)

	netXMLDef, err := getLibvirtNetworkXML()
	if err != nil {
		return fmt.Errorf("Failed to generate 'crc' network XML from template: %s", err)
	}
	netXMLDef = trimSpacesFromXML(netXMLDef)

	if stdOut != netXMLDef {
		logging.Debugf("libvirt 'crc' network definition does not have the expected value")
		logging.Debugf("expected: %s", netXMLDef)
		logging.Debugf("current: %s", stdOut)
		return fmt.Errorf("libvirt 'crc' network definition is incorrect")
	}
	logging.Debugf("libvirt 'crc' network has the expected value")
	return nil
}

func checkLibvirtCrcNetworkActive() error {
	logging.Debug("Checking if libvirt 'crc' network is active")
	stdOut, _, err := crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "net-info", "crc")
	if err != nil {
		return fmt.Errorf("Failed to query 'crc' network information")
	}
	outputSlice := strings.Split(stdOut, "\n")

	for _, stdOut = range outputSlice {
		stdOut = strings.TrimSpace(stdOut)
		if strings.HasPrefix(stdOut, "Active") && strings.Contains(stdOut, "yes") {
			logging.Debug("libvirt 'crc' network is already active")
			return nil
		}
	}
	return fmt.Errorf("Libvirt crc network is not active")
}

func fixLibvirtCrcNetworkActive() error {
	logging.Debug("Starting libvirt 'crc' network")
	stdOut, stdErr, err := crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "net-start", "crc")
	if err != nil {
		return fmt.Errorf("Failed to start libvirt 'crc' network %s %v: %s", stdOut, err, stdErr)
	}
	stdOut, stdErr, err = crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "net-autostart", "crc")
	if err != nil {
		return fmt.Errorf("Failed to autostart libvirt 'crc' network %s %v: %s", stdOut, err, stdErr)
	}
	logging.Debug("libvirt 'crc' network started")
	return nil
}

func getCPUFlags() (string, error) {
	// Check if the cpu flags vmx or svm is present
	out, err := ioutil.ReadFile("/proc/cpuinfo")
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

func runtimeExecutablePath() (string, error) {
	path, err := exec.LookPath(os.Args[0])
	if err != nil {
		// os.Args[0] is not in $PATH, crc must have been started by specifying the path to its binary
		path = os.Args[0]
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}

func checkCrcSymlink() error {
	runtimePath, err := runtimeExecutablePath()
	if err != nil {
		return err
	}
	symlinkPath, err := filepath.EvalSymlinks(constants.CrcSymlinkPath)
	if err != nil {
		return err
	}
	if symlinkPath != runtimePath {
		return fmt.Errorf("%s points to %s, not to %s", constants.CrcSymlinkPath, symlinkPath, runtimePath)
	}

	return nil
}

func fixCrcSymlink() error {
	_ = os.Remove(constants.CrcSymlinkPath)

	runtimePath, err := runtimeExecutablePath()
	if err != nil {
		return err
	}
	logging.Debugf("symlinking %s to %s", runtimePath, constants.CrcSymlinkPath)
	return os.Symlink(runtimePath, constants.CrcSymlinkPath)
}

func removeCrcSymlink() error {
	if crcos.FileExists(constants.CrcSymlinkPath) {
		return os.Remove(constants.CrcSymlinkPath)
	}
	return nil
}
