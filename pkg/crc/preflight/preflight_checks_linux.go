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
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/libvirt"
	"github.com/code-ready/crc/pkg/crc/systemd"
	"github.com/code-ready/crc/pkg/download"
	crcos "github.com/code-ready/crc/pkg/os"
	"golang.org/x/sys/unix"
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
	out, err := ioutil.ReadFile("/proc/cpuinfo")
	if err != nil {
		logging.Debugf("Failed to read /proc/cpuinfo: %v", err)
		return fmt.Errorf("Failed to read /proc/cpuinfo")
	}
	re := regexp.MustCompile(`flags.*:.*`)

	flags := re.FindString(string(out))
	if flags == "" {
		return fmt.Errorf("Could not find cpu flags from /proc/cpuinfo")
	}

	re = regexp.MustCompile(`(vmx|svm)`)

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
	cmd := exec.Command("modprobe", "kvm")
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		logging.Debugf("%v : %s", err, buf.String())
		return fmt.Errorf("Failed to load kvm module")
	}
	logging.Debug("kvm module loaded")
	return nil
}

func checkLibvirtInstalled() error {
	logging.Debug("Checking if 'virsh' is available")
	path, err := exec.LookPath("virsh")
	if err != nil {
		return fmt.Errorf("Libvirt cli virsh was not found in path")
	}
	logging.Debug("'virsh' was found in ", path)
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

func checkLibvirtEnabled() error {
	logging.Debug("Checking if libvirtd.service is enabled")
	// check if libvirt service is enabled
	path, err := exec.LookPath("systemctl")
	if err != nil {
		return fmt.Errorf("systemctl not found on path: %s", err.Error())
	}
	stdOut, _, err := crcos.RunWithDefaultLocale(path, "is-enabled", "libvirtd")
	if err != nil {
		return fmt.Errorf("Error checking if libvirtd service is enabled")
	}
	if strings.TrimSpace(stdOut) != "enabled" {
		return fmt.Errorf("libvirtd.service is not enabled")
	}
	logging.Debug("libvirtd.service is already enabled")
	return nil
}

func fixLibvirtEnabled() error {
	logging.Debug("Enabling libvirtd.service")
	// Start libvirt service
	path, err := exec.LookPath("systemctl")
	if err != nil {
		return err
	}
	_, _, err = crcos.RunWithPrivilege("enable libvirtd service", path, "enable", "libvirtd")
	if err != nil {
		return fmt.Errorf("Failed to enable libvirtd service")
	}
	logging.Debug("libvirtd.service is enabled")
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
	logging.Debug("Checking if libvirtd.service is running")
	path, err := exec.LookPath("systemctl")
	if err != nil {
		return err
	}
	stdOut, _, err := crcos.RunWithDefaultLocale(path, "is-active", "libvirtd")
	if err != nil {
		return fmt.Errorf("Failed to check if libvirtd service is active")
	}
	if strings.TrimSpace(stdOut) != "active" {
		return fmt.Errorf("libvirtd.service is not running")
	}
	logging.Debug("libvirtd.service is already running")
	return nil
}

func fixLibvirtServiceRunning() error {
	logging.Debug("Starting libvirtd.service")
	path, err := exec.LookPath("systemctl")
	if err != nil {
		return err
	}
	_, _, err = crcos.RunWithPrivilege("start libvirtd service", path, "start", "libvirtd")
	if err != nil {
		return fmt.Errorf("Failed to start libvirt service")
	}
	logging.Debug("libvirtd.service is running")
	return nil
}

func checkMachineDriverLibvirtInstalled() error {
	logging.Debugf("Checking if %s is installed", libvirt.MachineDriverCommand)

	// Check if crc-driver-libvirt is available
	libvirtDriverPath := filepath.Join(constants.CrcBinDir, libvirt.MachineDriverCommand)
	err := unix.Access(libvirtDriverPath, unix.X_OK)
	if err != nil {
		return fmt.Errorf("%s is not executable", libvirtDriverPath)
	}

	// Check the version of driver if it matches to supported one
	stdOut, stdErr, err := crcos.RunWithDefaultLocale(libvirtDriverPath, "version")
	if err != nil {
		return fmt.Errorf("Failed to check libvirt machine driver's version")
	}
	if !strings.Contains(stdOut, libvirt.MachineDriverVersion) {
		return fmt.Errorf("crc-driver-libvirt does not have right version \n Required: %s \n Got: %s use 'crc setup' command.\n %v\n", libvirt.MachineDriverVersion, stdOut, stdErr)
	}
	logging.Debugf("%s is already installed in %s", libvirt.MachineDriverCommand, libvirtDriverPath)
	return nil
}

func fixMachineDriverLibvirtInstalled() error {
	logging.Debugf("Installing %s", libvirt.MachineDriverCommand)
	_, err := extractBinary(libvirt.MachineDriverCommand, 0755)
	if err != nil {
		_, err = download.Download(libvirt.MachineDriverDownloadUrl, constants.CrcBinDir, 0755)
		if err != nil {
			logging.Debugf("download.Download() failed: %v", err)
			return fmt.Errorf("Failed to download libvirt machine driver")
		}
	}
	logging.Debugf("%s is installed in %s", libvirt.MachineDriverCommand, constants.CrcBinDir)

	return nil
}

/* These 2 checks can be removed after a few releases */
func checkOldMachineDriverLibvirtInstalled() error {
	logging.Debugf("Checking if an older libvirt driver %s is installed", libvirt.MachineDriverCommand)
	oldLibvirtDriverPath := filepath.Join("/usr/local/bin/", libvirt.MachineDriverCommand)
	if _, err := os.Stat(oldLibvirtDriverPath); !os.IsNotExist(err) {
		return fmt.Errorf("Found old system-wide crc-machine-driver binary")
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

func getLibvirtNetworkXml() (string, error) {
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

	netXMLDef, err := getLibvirtNetworkXml()
	if err != nil {
		logging.Debugf("getLibvirtNetworkXml() failed: %v", err)
		return fmt.Errorf("Failed to read libvirt 'crc' network definition")
	}

	// For time being we are going to override the crc network according what we have in our binary template.
	// We also don't care about the error or output from those commands atm.
	cmd := exec.Command("virsh", "--connect", "qemu:///system", "net-destroy", libvirt.DefaultNetwork)
	_ = cmd.Run()
	cmd = exec.Command("virsh", "--connect", "qemu:///system", "net-undefine", libvirt.DefaultNetwork)
	_ = cmd.Run()
	// Create the network according to our defined template
	cmd = exec.Command("virsh", "--connect", "qemu:///system", "net-define", "/dev/stdin")
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
	_, _, err := crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "dominfo", constants.DefaultName)
	if err != nil {
		//  User may have run `crc delete` before `crc cleanup`
		//  in that case there is no crc vm so return early.
		return nil
	}
	_, stderr, err := crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "destroy", constants.DefaultName)
	if err != nil {
		logging.Debugf("%v : %s", err, stderr)
		return fmt.Errorf("Failed to destroy 'crc' VM")
	}
	_, stderr, err = crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "undefine", constants.DefaultName)
	if err != nil {
		logging.Debugf("%v : %s", err, stderr)
		return fmt.Errorf("Failed to undefine 'crc' VM")
	}
	if err := os.RemoveAll(constants.MachineInstanceDir); err != nil {
		logging.Debugf("Error removing %s dir: %v", constants.MachineInstanceDir, err)
		return fmt.Errorf("Error removing %s dir", constants.MachineInstanceDir)
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

	netXMLDef, err := getLibvirtNetworkXml()
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
	cmd := exec.Command("virsh", "--connect", "qemu:///system", "net-start", "crc")
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		logging.Debugf("%v : %s", err, buf.String())
		return fmt.Errorf("Failed to start libvirt 'crc' network")
	}
	cmd = exec.Command("virsh", "--connect", "qemu:///system", "net-autostart", "crc")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to autostart libvirt 'crc' network")
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
	)
	if err != nil {
		return fmt.Errorf("Failed to write dnsmasq config file: %s: %v", crcDnsmasqConfigPath, err)
	}

	logging.Debug("Reloading NetworkManager")
	sd := systemd.NewHostSystemdCommander()
	if _, err := sd.Reload("NetworkManager"); err != nil {
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
		if _, err := sd.Reload("NetworkManager"); err != nil {
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
	)
	if err != nil {
		return fmt.Errorf("Failed to write NetworkManager config file: %s: %v", crcNetworkManagerConfigPath, err)
	}

	logging.Debug("Reloading NetworkManager")
	sd := systemd.NewHostSystemdCommander()
	if _, err := sd.Reload("NetworkManager"); err != nil {
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
		if _, err := sd.Reload("NetworkManager"); err != nil {
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
	path, err := exec.LookPath("systemctl")
	if err != nil {
		return err
	}
	stdOut, stdErr, err := crcos.RunWithDefaultLocale(path, "is-active", "NetworkManager")
	if err != nil {
		return fmt.Errorf("%v : %s", err, stdErr)
	}
	if strings.TrimSpace(stdOut) != "active" {
		return fmt.Errorf("NetworkManager.service is not running")
	}
	logging.Debug("NetworkManager.service is already running")
	return nil
}

func fixNetworkManagerIsRunning() error {
	return fmt.Errorf("NetworkManager is required. Please make sure it is installed and running manually")
}
