package preflight

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/libvirt"
	"github.com/code-ready/crc/pkg/crc/systemd"
	"github.com/code-ready/crc/pkg/download"

	"github.com/Masterminds/semver"

	crcos "github.com/code-ready/crc/pkg/os"
)

const (
	libvirtDriverCommand        = "crc-driver-libvirt"
	libvirtDriverVersion        = "0.12.6"
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
	libvirtDriverDownloadURL = fmt.Sprintf("https://github.com/code-ready/machine-driver-libvirt/releases/download/%s/crc-driver-libvirt", libvirtDriverVersion)
)

func checkVirtualizationEnabled() error {
	logging.Debug("Checking if the vmx/svm flags are present in /proc/cpuinfo")
	// Check if the cpu flags vmx or svm is present
	out, err := ioutil.ReadFile("/proc/cpuinfo")
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`flags.*:.*`)

	flags := re.FindString(string(out))
	if flags == "" {
		return errors.New("Could not find cpu flags from /proc/cpuinfo")
	}

	re = regexp.MustCompile(`(vmx|svm)`)

	cputype := re.FindString(flags)
	if cputype == "" {
		return errors.New("Virtualization is not available for you CPU")
	}
	logging.Debug("CPU virtualization flags are good")
	return nil
}

func fixVirtualizationEnabled() error {
	return errors.New("You need to enable virtualization in BIOS")
}

func checkKvmEnabled() error {
	logging.Debug("Checking if /dev/kvm exists")
	// Check if /dev/kvm exists
	if _, err := os.Stat("/dev/kvm"); os.IsNotExist(err) {
		return errors.New("kvm kernel module is not loaded")
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
		return fmt.Errorf("%v : %s", err, buf.String())
	}
	logging.Debug("kvm module loaded")
	return nil
}

func checkLibvirtInstalled() error {
	logging.Debug("Checking if 'virsh' is available")
	path, err := exec.LookPath("virsh")
	if err != nil {
		return errors.New("Libvirt cli virsh was not found in path")
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
	stdOut, stdErr, err := crcos.RunWithDefaultLocale(path, "is-enabled", "libvirtd")
	if err != nil {
		return fmt.Errorf("%v : %s", err, stdErr)
	}
	if strings.TrimSpace(stdOut) != "enabled" {
		return errors.New("libvirtd.service is not enabled")
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
	stdOut, stdErr, err := crcos.RunWithPrivilege("enable libvirtd service", path, "enable", "libvirtd")
	if err != nil {
		return fmt.Errorf("%s, %v : %s", stdOut, err, stdErr)
	}
	logging.Debug("libvirtd.service is enabled")
	return nil
}

func fixLibvirtVersion() error {
	return fmt.Errorf("libvirt v%s or newer is required and must be updated manually", minSupportedLibvirtVersion)
}

func checkLibvirtVersion() error {
	logging.Debugf("Checking if libvirt version is >=%s", minSupportedLibvirtVersion)
	stdOut, stdErr, err := crcos.RunWithDefaultLocale("virsh", "-v")
	if err != nil {
		return fmt.Errorf("%v : %s", err, stdErr)
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
		return err
	}
	path, err := exec.LookPath("groups")
	if err != nil {
		return err
	}
	stdOut, stdErr, err := crcos.RunWithDefaultLocale(path, currentUser.Username)
	if err != nil {
		return fmt.Errorf("%+v : %s", err, stdErr)
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
		return err
	}
	stdOut, stdErr, err := crcos.RunWithPrivilege("add user to libvirt group", "usermod", "-a", "-G", "libvirt", currentUser.Username)
	if err != nil {
		return fmt.Errorf("%s %v : %s", stdOut, err, stdErr)
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
	stdOut, stdErr, err := crcos.RunWithDefaultLocale(path, "is-active", "libvirtd")
	if err != nil {
		return fmt.Errorf("%v : %s", err, stdErr)
	}
	if strings.TrimSpace(stdOut) != "active" {
		return errors.New("libvirtd.service is not running")
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
	stdOut, stdErr, err := crcos.RunWithPrivilege("start libvirtd service", path, "start", "libvirtd")
	if err != nil {
		return fmt.Errorf("%s %v : %s", stdOut, err, stdErr)
	}
	logging.Debug("libvirtd.service is running")
	return nil
}

func checkMachineDriverLibvirtInstalled() error {
	logging.Debugf("Checking if %s is installed", libvirtDriverCommand)

	// Check if crc-driver-libvirt is available
	libvirtDriverPath := filepath.Join(constants.CrcBinDir, libvirtDriverCommand)
	err := unix.Access(libvirtDriverPath, unix.X_OK)
	if err != nil {
		logging.Debugf("%s not executable", libvirtDriverPath)
		return err
	}

	// Check the version of driver if it matches to supported one
	stdOut, stdErr, err := crcos.RunWithDefaultLocale(libvirtDriverPath, "version")
	if err != nil {
		return fmt.Errorf("%v : %s", err, stdErr)
	}
	if !strings.Contains(stdOut, libvirtDriverVersion) {
		return fmt.Errorf("crc-driver-libvirt does not have right version \n Required: %s \n Got: %s use 'crc setup' command.\n %v\n", libvirtDriverVersion, stdOut, stdErr)
	}
	logging.Debugf("%s is already installed in %s", libvirtDriverCommand, libvirtDriverPath)
	return nil
}

func fixMachineDriverLibvirtInstalled() error {
	logging.Debugf("Installing %s", libvirtDriverCommand)
	_, err := download.Download(libvirtDriverDownloadURL, constants.CrcBinDir, 0755)
	if err != nil {
		return err
	}
	logging.Debugf("%s is installed in %s", libvirtDriverCommand, constants.CrcBinDir)

	return nil
}

/* These 2 checks can be removed after a few releases */
func checkOldMachineDriverLibvirtInstalled() error {
	logging.Debugf("Checking if an older libvirt driver %s is installed", libvirtDriverCommand)
	oldLibvirtDriverPath := filepath.Join("/usr/local/bin/", libvirtDriverCommand)
	if _, err := os.Stat(oldLibvirtDriverPath); !os.IsNotExist(err) {
		return fmt.Errorf("Found old system-wide crc-machine-driver binary")
	}
	logging.Debugf("No older %s installation found", libvirtDriverCommand)

	return nil
}

func fixOldMachineDriverLibvirtInstalled() error {
	oldLibvirtDriverPath := filepath.Join("/usr/local/bin/", libvirtDriverCommand)
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
	stdOut, stdErr, err := crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "net-list")
	if err != nil {
		return fmt.Errorf("%+v: %s", err, stdErr)
	}
	outputSlice := strings.Split(stdOut, "\n")
	for _, stdOut = range outputSlice {
		stdOut = strings.TrimSpace(stdOut)
		if strings.HasPrefix(stdOut, "crc") {
			logging.Debug("libvirt 'crc' network exists")
			return nil
		}
	}
	return errors.New("Libvirt network crc not found")
}

func fixLibvirtCrcNetworkAvailable() error {
	logging.Debug("Creating libvirt 'crc' network")
	config := libvirt.NetworkConfig{
		NetworkName: libvirt.DefaultNetwork,
		MAC:         libvirt.MACAddress,
		IP:          libvirt.IPAddress,
	}

	t, err := template.New("netxml").Parse(libvirt.NetworkTemplate)
	if err != nil {
		return err
	}
	var netXMLDef strings.Builder
	err = t.Execute(&netXMLDef, config)
	if err != nil {
		return err
	}
	// For time being we are going to override the crc network according what we have in our binary template.
	// We also don't care about the error or output from those commands atm.
	cmd := exec.Command("virsh", "--connect", "qemu:///system", "net-destroy", libvirt.DefaultNetwork)
	_ = cmd.Run()
	cmd = exec.Command("virsh", "--connect", "qemu:///system", "net-undefine", libvirt.DefaultNetwork)
	_ = cmd.Run()
	// Create the network according to our defined template
	cmd = exec.Command("virsh", "--connect", "qemu:///system", "net-define", "/dev/stdin")
	cmd.Stdin = strings.NewReader(netXMLDef.String())
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%v : %s", err, buf.String())
	}
	logging.Debug("libvirt 'crc' network created")
	return nil
}

func checkLibvirtCrcNetworkActive() error {
	logging.Debug("Checking if libvirt 'crc' network is active")
	stdOut, stdErr, err := crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "net-list")
	if err != nil {
		return fmt.Errorf("%+v: %s", err, stdErr)
	}
	outputSlice := strings.Split(stdOut, "\n")

	for _, stdOut = range outputSlice {
		stdOut = strings.TrimSpace(stdOut)
		if strings.HasPrefix(stdOut, "crc") && strings.Contains(stdOut, "active") {
			logging.Debug("libvirt 'crc' network is already active")
			return nil
		}
	}
	return errors.New("Libvirt crc network is not active")
}

func fixLibvirtCrcNetworkActive() error {
	logging.Debug("Starting libvirt 'crc' network")
	cmd := exec.Command("virsh", "--connect", "qemu:///system", "net-start", "crc")
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%v : %s", err, buf.String())
	}
	cmd = exec.Command("virsh", "--connect", "qemu:///system", "net-autostart", "crc")
	err = cmd.Run()
	if err != nil {
		return err
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

func checkNetworkManagerInstalled() error {
	logging.Debug("Checking if 'nmcli' is available")
	path, err := exec.LookPath("nmcli")
	if err != nil {
		return errors.New("NetworkManager cli nmcli was not found in path")
	}
	logging.Debug("'nmcli' was found in ", path)
	return nil
}

func fixNetworkManagerInstalled() error {
	return fmt.Errorf("NetworkManager is required and must be installed manually")
}

func CheckNetworkManagerIsRunning() error {
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
		return errors.New("NetworkManager.service is not running")
	}
	logging.Debug("NetworkManager.service is already running")
	return nil
}

func fixNetworkManagerIsRunning() error {
	return fmt.Errorf("NetworkManager is required. Please make sure it is installed and running manually")
}
