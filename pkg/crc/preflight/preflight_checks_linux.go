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

	"github.com/Masterminds/semver"
	"github.com/code-ready/crc/pkg/crc/cache"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/libvirt"
	"github.com/code-ready/crc/pkg/crc/systemd"
	"github.com/code-ready/crc/pkg/crc/systemd/states"
	crcos "github.com/code-ready/crc/pkg/os"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

const (
	crcDnsmasqConfigFile        = "crc.conf"
	crcNetworkManagerConfigFile = "crc-nm-dnsmasq.conf"
	// This is defined in https://github.com/code-ready/machine-driver-libvirt/blob/master/go.mod#L5
	minSupportedLibvirtVersion = "3.4.0"
)

var (
	crcDnsmasqConfigPath = filepath.Join(string(filepath.Separator), "etc", "NetworkManager", "dnsmasq.d", crcDnsmasqConfigFile)
	crcDnsmasqConfig     = `server=/apps-crc.testing/192.168.130.11
server=/crc.testing/192.168.130.11
`
	crcNetworkManagerConfigPath = filepath.Join(string(filepath.Separator), "etc", "NetworkManager", "conf.d", crcNetworkManagerConfigFile)
	crcNetworkManagerConfig     = `[main]
dns=dnsmasq
`
)

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
		stdOut, stdErr, err := crcos.RunWithPrivilege("Load kvm_intel kernel module", "modprobe", "kvm_intel")
		if err != nil {
			return fmt.Errorf("Failed to load kvm intel module: %s %v: %s", stdOut, err, stdErr)
		}
	case strings.Contains(flags, "svm"):
		stdOut, stdErr, err := crcos.RunWithPrivilege("Load kvm_amd kernel module", "modprobe", "kvm_amd")
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
	stdOut, _, err := crcos.RunWithDefaultLocale("virsh", "capabilities")
	if err != nil {
		return nil, fmt.Errorf("Failed to run 'virsh capabilities': %v", err)
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

func fixLibvirtInstalled() error {
	logging.Debug("Trying to install libvirt")
	stdOut, stdErr, err := crcos.RunWithPrivilege("install virtualization related packages", "yum", "install", "-y", "libvirt", "libvirt-daemon-kvm", "qemu-kvm")
	if err != nil {
		return fmt.Errorf("Could not install required packages: %s %v: %s", stdOut, err, stdErr)
	}
	logging.Debug("libvirt was successfully installed")
	return nil
}

func fixLibvirtVersion() error {
	return fmt.Errorf("libvirt v%s or newer is required and must be updated manually", minSupportedLibvirtVersion)
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
	// check if user is part of libvirt group
	currentUser, err := user.Current()
	if err != nil {
		logging.Debugf("user.Current() failed: %v", err)
		return fmt.Errorf("Failed to get current user id")
	}
	path, err := exec.LookPath("groups")
	if err != nil {
		return fmt.Errorf("Failed to locate 'groups' command")
	}
	stdOut, _, err := crcos.RunWithDefaultLocale(path, currentUser.Username)
	if err != nil {
		return fmt.Errorf("Failed to look up current user's groups")
	}
	if strings.Contains(stdOut, "libvirt") {
		logging.Debug("Current user is already in the libvirt group")
		return nil
	}
	return fmt.Errorf("%s not part of libvirtd group", currentUser.Username)
}

func fixUserPartOfLibvirtGroup() error {
	logging.Debug("Adding current user to the libvirt group")
	currentUser, err := user.Current()
	if err != nil {
		logging.Debugf("user.Current() failed: %v", err)
		return fmt.Errorf("Failed to get current user id")
	}
	_, _, err = crcos.RunWithPrivilege("add user to libvirt group", "usermod", "-a", "-G", "libvirt", currentUser.Username)
	if err != nil {
		return fmt.Errorf("Failed to add user to libvirt group")
	}
	logging.Debug("Current user is in the libvirt group")
	return nil
}

func checkLibvirtServiceRunning() error {
	logging.Debug("Checking if libvirtd service is running")
	sd := systemd.NewHostSystemdCommander()

	libvirtSystemdUnits := []string{"virtqemud.socket", "libvirtd.socket", "virtqemud.service", "libvirtd.service"}
	for _, unit := range libvirtSystemdUnits {
		status, err := sd.Status(unit)
		if err == nil && status == states.Running {
			logging.Debugf("%s is running", unit)
			return nil
		}

		logging.Debugf("%s is not running", unit)
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

	logging.Debugf("Checking if %s is installed", machineDriverLibvirt.GetBinaryName())

	if !machineDriverLibvirt.IsCached() {
		return fmt.Errorf("%s executable is not cached", machineDriverLibvirt.GetBinaryName())
	}
	if err := machineDriverLibvirt.CheckVersion(); err != nil {
		return err
	}
	logging.Debugf("%s is already installed", machineDriverLibvirt.GetBinaryName())
	return nil
}

func fixMachineDriverLibvirtInstalled() error {
	machineDriverLibvirt := cache.NewMachineDriverLibvirtCache()

	logging.Debugf("Installing %s", machineDriverLibvirt.GetBinaryName())

	if err := machineDriverLibvirt.EnsureIsCached(); err != nil {
		return fmt.Errorf("Unable to download %s: %v", machineDriverLibvirt.GetBinaryName(), err)
	}
	logging.Debugf("%s is installed in %s", machineDriverLibvirt.GetBinaryName(), filepath.Dir(machineDriverLibvirt.GetBinaryPath()))
	return nil
}

/* These 2 checks can be removed after a few releases */
func checkOldMachineDriverLibvirtInstalled() error {
	logging.Debugf("Checking if an older libvirt driver %s is installed", libvirt.MachineDriverCommand)
	oldLibvirtDriverPath := filepath.Join("/usr/local/bin/", libvirt.MachineDriverCommand)
	if _, err := os.Stat(oldLibvirtDriverPath); !os.IsNotExist(err) {
		return fmt.Errorf("Found old system-wide crc-machine-driver executable")
	}
	logging.Debugf("No older %s installation found", libvirt.MachineDriverCommand)

	return nil
}

func fixOldMachineDriverLibvirtInstalled() error {
	oldLibvirtDriverPath := filepath.Join("/usr/local/bin/", libvirt.MachineDriverCommand)
	logging.Debugf("Removing %s", oldLibvirtDriverPath)
	_, _, err := crcos.RunWithPrivilege("remove old libvirt driver", "rm", "-f", oldLibvirtDriverPath)
	if err != nil {
		logging.Debugf("Removal of %s failed", oldLibvirtDriverPath)
		/* Ignoring error, an obsolete file being still present is not a fatal error */
	} else {
		logging.Debugf("%s successfully removed", oldLibvirtDriverPath)
	}

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
		return fmt.Errorf("Failed to create libvirt 'crc' network")
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
	logging.Debug("Removing 'crc' VM")

	if err := os.RemoveAll(constants.MachineInstanceDir); err != nil {
		logging.Debugf("Error removing %s dir: %v", constants.MachineInstanceDir, err)
		return fmt.Errorf("Error removing %s dir", constants.MachineInstanceDir)
	}

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

func checkCrcDnsmasqConfigFile() error {
	logging.Debug("Checking dnsmasq configuration")
	c := []byte(crcDnsmasqConfig)
	_, err := os.Stat(crcDnsmasqConfigPath)
	if err != nil {
		return fmt.Errorf("File not found: %s: %s", crcDnsmasqConfigPath, err.Error())
	}
	config, err := ioutil.ReadFile(filepath.Clean(crcDnsmasqConfigPath))
	if err != nil {
		return fmt.Errorf("Error opening file: %s: %s", crcDnsmasqConfigPath, err.Error())
	}
	if !bytes.Equal(config, c) {
		return fmt.Errorf("Config file contains changes: %s", crcDnsmasqConfigPath)
	}
	logging.Debug("dnsmasq configuration is good")
	return nil
}

func fixCrcDnsmasqConfigFile() error {
	logging.Debug("Fixing dnsmasq configuration")
	err := crcos.WriteToFileAsRoot(
		fmt.Sprintf("write dnsmasq configuration in %s", crcDnsmasqConfigPath),
		crcDnsmasqConfig,
		crcDnsmasqConfigPath,
		0644,
	)
	if err != nil {
		return fmt.Errorf("Failed to write dnsmasq config file: %s: %v", crcDnsmasqConfigPath, err)
	}

	logging.Debug("Reloading NetworkManager")
	sd := systemd.NewHostSystemdCommander()
	if err := sd.Reload("NetworkManager"); err != nil {
		return fmt.Errorf("Failed to restart NetworkManager: %v", err)
	}

	logging.Debug("dnsmasq configuration fixed")
	return nil
}

func removeCrcDnsmasqConfigFile() error {
	if err := checkNetworkManagerInstalled(); err != nil {
		// When NetworkManager is not installed, this file won't exist
		return nil
	}
	// Delete the `crcDnsmasqConfigPath` file if exists,
	// ignore all the os PathError except `IsNotExist` one.
	if _, err := os.Stat(crcDnsmasqConfigPath); !os.IsNotExist(err) {
		logging.Debug("Removing dnsmasq configuration")
		err := crcos.RemoveFileAsRoot(
			fmt.Sprintf("removing dnsmasq configuration in %s", crcDnsmasqConfigPath),
			crcDnsmasqConfigPath,
		)
		if err != nil {
			return fmt.Errorf("Failed to remove dnsmasq config file: %s: %v", crcDnsmasqConfigPath, err)
		}

		logging.Debug("Reloading NetworkManager")
		sd := systemd.NewHostSystemdCommander()
		if err := sd.Reload("NetworkManager"); err != nil {
			return fmt.Errorf("Failed to restart NetworkManager: %v", err)
		}
	}
	return nil
}

func checkCrcNetworkManagerConfig() error {
	logging.Debug("Checking NetworkManager configuration")
	c := []byte(crcNetworkManagerConfig)
	_, err := os.Stat(crcNetworkManagerConfigPath)
	if err != nil {
		return fmt.Errorf("File not found: %s: %s", crcNetworkManagerConfigPath, err.Error())
	}
	config, err := ioutil.ReadFile(filepath.Clean(crcNetworkManagerConfigPath))
	if err != nil {
		return fmt.Errorf("Error opening file: %s: %s", crcNetworkManagerConfigPath, err.Error())
	}
	if !bytes.Equal(config, c) {
		return fmt.Errorf("Config file contains changes: %s", crcNetworkManagerConfigPath)
	}
	logging.Debug("NetworkManager configuration is good")
	return nil
}

func fixCrcNetworkManagerConfig() error {
	logging.Debug("Fixing NetworkManager configuration")
	err := crcos.WriteToFileAsRoot(
		fmt.Sprintf("write NetworkManager config in %s", crcNetworkManagerConfigPath),
		crcNetworkManagerConfig,
		crcNetworkManagerConfigPath,
		0644,
	)
	if err != nil {
		return fmt.Errorf("Failed to write NetworkManager config file: %s: %v", crcNetworkManagerConfigPath, err)
	}

	logging.Debug("Reloading NetworkManager")
	sd := systemd.NewHostSystemdCommander()
	if err := sd.Reload("NetworkManager"); err != nil {
		return fmt.Errorf("Failed to restart NetworkManager: %v", err)
	}

	logging.Debug("NetworkManager configuration fixed")
	return nil
}

func removeCrcNetworkManagerConfig() error {
	if err := checkNetworkManagerInstalled(); err != nil {
		// When NetworkManager is not installed, this file won't exist
		return nil
	}
	if _, err := os.Stat(crcNetworkManagerConfigPath); !os.IsNotExist(err) {
		logging.Debug("Removing NetworkManager configuration")
		err := crcos.RemoveFileAsRoot(
			fmt.Sprintf("Removing NetworkManager config in %s", crcNetworkManagerConfigPath),
			crcNetworkManagerConfigPath,
		)
		if err != nil {
			return fmt.Errorf("Failed to remove NetworkManager config file: %s: %v", crcNetworkManagerConfigPath, err)
		}

		logging.Debug("Reloading NetworkManager")
		sd := systemd.NewHostSystemdCommander()
		if err := sd.Reload("NetworkManager"); err != nil {
			return fmt.Errorf("Failed to restart NetworkManager: %v", err)
		}
	}
	return nil
}

func checkNetworkManagerInstalled() error {
	logging.Debug("Checking if 'nmcli' is available")
	path, err := exec.LookPath("nmcli")
	if err != nil {
		return fmt.Errorf("NetworkManager cli nmcli was not found in path")
	}
	logging.Debug("'nmcli' was found in ", path)
	return nil
}

func fixNetworkManagerInstalled() error {
	return fmt.Errorf("NetworkManager is required and must be installed manually")
}

func checkNetworkManagerIsRunning() error {
	logging.Debug("Checking if NetworkManager.service is running")
	sd := systemd.NewHostSystemdCommander()
	status, err := sd.Status("NetworkManager")
	if err != nil {
		return err
	}
	if status != states.Running {
		return fmt.Errorf("NetworkManager.service is not running")
	}
	logging.Debug("NetworkManager.service is already running")
	return nil
}

func fixNetworkManagerIsRunning() error {
	return fmt.Errorf("NetworkManager is required. Please make sure it is installed and running manually")
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
