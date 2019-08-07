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

	crcos "github.com/code-ready/crc/pkg/os"
)

const (
	libvirtDriverCommand        = "crc-driver-libvirt"
	libvirtDriverVersion        = "0.12.5"
	crcDnsmasqConfigFile        = "crc.conf"
	crcNetworkManagerConfigFile = "crc-nm-dnsmasq.conf"
)

var (
	crcDnsmasqConfigPath = filepath.Join(string(filepath.Separator), "etc", "NetworkManager", "dnsmasq.d", crcDnsmasqConfigFile)
	crcDnsmasqConfig     = `server=/crc.testing/192.168.130.1
address=/apps-crc.testing/192.168.130.11
`
	crcNetworkManagerConfigPath = filepath.Join(string(filepath.Separator), "etc", "NetworkManager", "conf.d", crcNetworkManagerConfigFile)
	crcNetworkManagerConfig     = `[main]
dns=dnsmasq
`
	libvirtDriverDownloadURL = fmt.Sprintf("https://github.com/code-ready/machine-driver-libvirt/releases/download/%s/crc-driver-libvirt", libvirtDriverVersion)
)

func checkVirtualizationEnabled() (bool, error) {
	logging.Debug("Checking if the vmx/svm flags are present in /proc/cpuinfo")
	// Check if the cpu flags vmx or svm is present
	out, err := ioutil.ReadFile("/proc/cpuinfo")
	if err != nil {
		return false, err
	}
	re := regexp.MustCompile(`flags.*:.*`)

	flags := re.FindString(string(out))
	if flags == "" {
		return false, errors.New("Could not find cpu flags from /proc/cpuinfo")
	}

	re = regexp.MustCompile(`(vmx|svm)`)

	cputype := re.FindString(flags)
	if cputype == "" {
		return false, errors.New("Virtualization is not available for you CPU")
	}
	logging.Debug("CPU virtualization flags are good")
	return true, nil
}

func fixVirtualizationEnabled() (bool, error) {
	return false, errors.New("You need to enable virtualization in BIOS")
}

func checkKvmEnabled() (bool, error) {
	logging.Debug("Checking if /dev/kvm exists")
	// Check if /dev/kvm exists
	if _, err := os.Stat("/dev/kvm"); os.IsNotExist(err) {
		return false, errors.New("kvm kernel module is not loaded")
	}
	logging.Debug("/dev/kvm was found")
	return true, nil
}

func fixKvmEnabled() (bool, error) {
	logging.Debug("Trying to load kvm module")
	cmd := exec.Command("modprobe", "kvm")
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("%v : %s", err, buf.String())
	}
	logging.Debug("kvm module loaded")
	return true, nil
}

func checkLibvirtInstalled() (bool, error) {
	logging.Debug("Checking if 'virsh' is available")
	path, err := exec.LookPath("virsh")
	if err != nil {
		return false, errors.New("Libvirt cli virsh was not found in path")
	}
	logging.Debug("'virsh' was found in ", path)
	return true, nil
}

func fixLibvirtInstalled() (bool, error) {
	logging.Debug("Trying to install libvirt")
	stdOut, stdErr, err := crcos.RunWithPrivilege("yum", "install", "-y", "libvirt", "libvirt-daemon-kvm", "qemu-kvm")
	if err != nil {
		return false, fmt.Errorf("Could not install required packages: %s %v: %s", stdOut, err, stdErr)
	}
	logging.Debug("libvirt was successfully installed")
	return true, nil
}

func checkLibvirtEnabled() (bool, error) {
	logging.Debug("Checking if libvirtd.service is enabled")
	// check if libvirt service is enabled
	path, err := exec.LookPath("systemctl")
	if err != nil {
		return false, fmt.Errorf("systemctl not found on path: %s", err.Error())
	}
	stdOut, stdErr, err := crcos.RunWithDefaultLocale(path, "is-enabled", "libvirtd")
	if err != nil {
		return false, fmt.Errorf("%v : %s", err, stdErr)
	}
	if strings.TrimSpace(stdOut) != "enabled" {
		return false, errors.New("libvirtd.service is not enabled")
	}
	logging.Debug("libvirtd.service is already enabled")
	return true, nil
}

func fixLibvirtEnabled() (bool, error) {
	logging.Debug("Enabling libvirtd.service")
	// Start libvirt service
	path, err := exec.LookPath("systemctl")
	if err != nil {
		return false, err
	}
	stdOut, stdErr, err := crcos.RunWithPrivilege(path, "enable", "libvirtd")
	if err != nil {
		return false, fmt.Errorf("%s, %v : %s", stdOut, err, stdErr)
	}
	logging.Debug("libvirtd.service is enabled")
	return true, nil
}

func checkUserPartOfLibvirtGroup() (bool, error) {
	logging.Debug("Checking if current user is part of the libvirt group")
	// check if user is part of libvirt group
	currentUser, err := user.Current()
	if err != nil {
		return false, err
	}
	path, err := exec.LookPath("groups")
	if err != nil {
		return false, err
	}
	stdOut, stdErr, err := crcos.RunWithDefaultLocale(path, currentUser.Username)
	if err != nil {
		return false, fmt.Errorf("%+v : %s", err, stdErr)
	}
	if strings.Contains(stdOut, "libvirt") {
		logging.Debug("Current user is already in the libvirt group")
		return true, nil
	}
	return false, fmt.Errorf("%s not part of libvirtd group", currentUser.Username)
}

func fixUserPartOfLibvirtGroup() (bool, error) {
	logging.Debug("Adding current user to the libvirt group")
	// Add user to libvirt/libvirtd group based on distro
	currentUser, err := user.Current()
	if err != nil {
		return false, err
	}
	stdOut, stdErr, err := crcos.RunWithPrivilege("usermod", "-a", "-G", "libvirt", currentUser.Username)
	if err != nil {
		return false, fmt.Errorf("%s %v : %s", stdOut, err, stdErr)
	}
	logging.Debug("Current user is in the libvirt group")
	return true, nil
}

func checkLibvirtServiceRunning() (bool, error) {
	logging.Debug("Checking if libvirtd.service is running")
	path, err := exec.LookPath("systemctl")
	if err != nil {
		return false, err
	}
	stdOut, stdErr, err := crcos.RunWithDefaultLocale(path, "is-active", "libvirtd")
	if err != nil {
		return false, fmt.Errorf("%v : %s", err, stdErr)
	}
	if strings.TrimSpace(stdOut) != "active" {
		return false, errors.New("libvirtd.service is not running")
	}
	logging.Debug("libvirtd.service is already running")
	return true, nil
}

func fixLibvirtServiceRunning() (bool, error) {
	logging.Debug("Starting libvirtd.service")
	path, err := exec.LookPath("systemctl")
	if err != nil {
		return false, err
	}
	stdOut, stdErr, err := crcos.RunWithPrivilege(path, "start", "libvirtd")
	if err != nil {
		return false, fmt.Errorf("%s %v : %s", stdOut, err, stdErr)
	}
	logging.Debug("libvirtd.service is running")
	return true, nil
}

func checkMachineDriverLibvirtInstalled() (bool, error) {
	logging.Debugf("Checking if %s is installed", libvirtDriverCommand)

	// Check if crc-driver-libvirt is available
	libvirtDriverPath := filepath.Join(constants.CrcBinDir, libvirtDriverCommand)
	err := unix.Access(libvirtDriverPath, unix.X_OK)
	if err != nil {
		logging.Debugf("%s not executable", libvirtDriverPath)
		return false, err
	}

	// Check the version of driver if it matches to supported one
	stdOut, stdErr, err := crcos.RunWithDefaultLocale(libvirtDriverPath, "version")
	if err != nil {
		return false, fmt.Errorf("%v : %s", err, stdErr)
	}
	if !strings.Contains(stdOut, libvirtDriverVersion) {
		return false, fmt.Errorf("crc-driver-libvirt does not have right version \n Required: %s \n Got: %s use 'crc setup' command.\n %v\n", libvirtDriverVersion, stdOut, stdErr)
	}
	logging.Debugf("%s is already installed in %s", libvirtDriverCommand, libvirtDriverPath)
	return true, nil
}

func fixMachineDriverLibvirtInstalled() (bool, error) {
	logging.Debugf("Installing %s", libvirtDriverCommand)
	_, err := download.Download(libvirtDriverDownloadURL, constants.CrcBinDir, 0755)
	if err != nil {
		return false, err
	}
	logging.Debugf("%s is installed in %s", libvirtDriverCommand, constants.CrcBinDir)

	return true, nil
}

/* These 2 checks can be removed after a few releases */
func checkOldMachineDriverLibvirtInstalled() (bool, error) {
	logging.Debugf("Checking if an older libvirt driver %s is installed", libvirtDriverCommand)
	oldLibvirtDriverPath := filepath.Join("/usr/local/bin/", libvirtDriverCommand)
	if _, err := os.Stat(oldLibvirtDriverPath); !os.IsNotExist(err) {
		return false, fmt.Errorf("Found old system-wide crc-machine-driver binary")
	}
	logging.Debugf("No older %s installation found", libvirtDriverCommand)

	return true, nil
}

func fixOldMachineDriverLibvirtInstalled() (bool, error) {
	oldLibvirtDriverPath := filepath.Join("/usr/local/bin/", libvirtDriverCommand)
	logging.Debugf("Removing %s", oldLibvirtDriverPath)
	_, _, err := crcos.RunWithPrivilege("rm", "-f", oldLibvirtDriverPath)
	if err != nil {
		logging.Debugf("Removal of %s failed", oldLibvirtDriverPath)
		/* Ignoring error, an obsolete file being still present is not a fatal error */
	} else {
		logging.Debugf("%s successfully removed", oldLibvirtDriverPath)
	}

	return true, nil
}

func checkLibvirtCrcNetworkAvailable() (bool, error) {
	logging.Debug("Checking if libvirt 'crc' network exists")
	stdOut, stdErr, err := crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "net-list")
	if err != nil {
		return false, fmt.Errorf("%+v: %s", err, stdErr)
	}
	outputSlice := strings.Split(stdOut, "\n")
	for _, stdOut = range outputSlice {
		stdOut = strings.TrimSpace(stdOut)
		if strings.HasPrefix(stdOut, "crc") {
			logging.Debug("libvirt 'crc' network exists")
			logging.Debug("Checking if libvirt 'crc' network definition contains ", constants.DefaultHostname)
			// Check if the network have defined hostname. It might be possible that User already have an outdated crc network
			stdOut, stdErr, err = crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "net-dumpxml", libvirt.DefaultNetwork)
			if err != nil {
				return false, fmt.Errorf("%+v: %s", err, stdErr)
			}
			if !strings.Contains(stdOut, constants.DefaultHostname) {
				return false, fmt.Errorf("crc network is not updated with %s hostname, use 'crc setup' to update it.", constants.DefaultHostname)
			}
			logging.Debug("libvirt 'crc' network definition contains ", constants.DefaultHostname)
			return true, nil
		}
	}
	return false, errors.New("Libvirt network crc not found")
}

func fixLibvirtCrcNetworkAvailable() (bool, error) {
	logging.Debug("Creating libvirt 'crc' network")
	config := libvirt.NetworkConfig{
		NetworkName: libvirt.DefaultNetwork,
		HostName:    constants.DefaultHostname,
		MAC:         libvirt.MACAddress,
		IP:          libvirt.IPAddress,
	}

	t, err := template.New("netxml").Parse(libvirt.NetworkTemplate)
	if err != nil {
		return false, err
	}
	var netXMLDef strings.Builder
	err = t.Execute(&netXMLDef, config)
	if err != nil {
		return false, err
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
		return false, fmt.Errorf("%v : %s", err, buf.String())
	}
	logging.Debug("libvirt 'crc' network created")
	return true, nil
}

func checkLibvirtCrcNetworkActive() (bool, error) {
	logging.Debug("Checking if libvirt 'crc' network is active")
	stdOut, stdErr, err := crcos.RunWithDefaultLocale("virsh", "--connect", "qemu:///system", "net-list")
	if err != nil {
		return false, fmt.Errorf("%+v: %s", err, stdErr)
	}
	outputSlice := strings.Split(stdOut, "\n")

	for _, stdOut = range outputSlice {
		stdOut = strings.TrimSpace(stdOut)
		if strings.HasPrefix(stdOut, "crc") && strings.Contains(stdOut, "active") {
			logging.Debug("libvirt 'crc' network is already active")
			return true, nil
		}
	}
	return false, errors.New("Libvirt crc network is not active")
}

func fixLibvirtCrcNetworkActive() (bool, error) {
	logging.Debug("Starting libvirt 'crc' network")
	cmd := exec.Command("virsh", "--connect", "qemu:///system", "net-start", "crc")
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("%v : %s", err, buf.String())
	}
	cmd = exec.Command("virsh", "--connect", "qemu:///system", "net-autostart", "crc")
	err = cmd.Run()
	if err != nil {
		return false, err
	}
	logging.Debug("libvirt 'crc' network started")
	return true, nil
}

func checkCrcDnsmasqConfigFile() (bool, error) {
	logging.Debug("Checking dnsmasq configuration")
	c := []byte(crcDnsmasqConfig)
	_, err := os.Stat(crcDnsmasqConfigPath)
	if err != nil {
		return false, fmt.Errorf("File not found: %s: %s", crcDnsmasqConfigPath, err.Error())
	}
	config, err := ioutil.ReadFile(crcDnsmasqConfigPath)
	if err != nil {
		return false, fmt.Errorf("Error opening file: %s: %s", crcDnsmasqConfigPath, err.Error())
	}
	if !bytes.Equal(config, c) {
		return false, fmt.Errorf("Config file contains changes: %s", crcDnsmasqConfigPath)
	}
	logging.Debug("dnsmasq configuration is good")
	return true, nil
}

func fixCrcDnsmasqConfigFile() (bool, error) {
	logging.Debug("Fixing dnsmasq configuration")
	cmd := exec.Command("sudo", "tee", crcDnsmasqConfigPath)
	cmd.Stdin = strings.NewReader(crcDnsmasqConfig)
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("Failed to write dnsmasq config file: %s: %s: %v", crcDnsmasqConfigPath, buf.String(), err)
	}

	logging.Debug("Reloading NetworkManager")
	sd := systemd.NewHostSystemdCommander()
	if _, err := sd.Reload("NetworkManager"); err != nil {
		return false, fmt.Errorf("Failed to restart NetworkManager: %v", err)
	}

	logging.Debug("dnsmasq configuration fixed")
	return true, nil
}
func checkCrcNetworkManagerConfig() (bool, error) {
	logging.Debug("Checking NetworkManager configuration")
	c := []byte(crcNetworkManagerConfig)
	_, err := os.Stat(crcNetworkManagerConfigPath)
	if err != nil {
		return false, fmt.Errorf("File not found: %s: %s", crcNetworkManagerConfigPath, err.Error())
	}
	config, err := ioutil.ReadFile(crcNetworkManagerConfigPath)
	if err != nil {
		return false, fmt.Errorf("Error opening file: %s: %s", crcNetworkManagerConfigPath, err.Error())
	}
	if !bytes.Equal(config, c) {
		return false, fmt.Errorf("Config file contains changes: %s", crcNetworkManagerConfigPath)
	}
	logging.Debug("NetworkManager configuration is good")
	return true, nil
}
func fixCrcNetworkManagerConfig() (bool, error) {
	logging.Debug("Fixing NetworkManager configuration")
	cmd := exec.Command("sudo", "tee", crcNetworkManagerConfigPath)
	cmd.Stdin = strings.NewReader(crcNetworkManagerConfig)
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("Failed to write NetworkManager config file: %s: %s: %v", crcNetworkManagerConfigPath, buf.String(), err)
	}

	logging.Debug("Reloading NetworkManager")
	sd := systemd.NewHostSystemdCommander()
	if _, err := sd.Reload("NetworkManager"); err != nil {
		return false, fmt.Errorf("Failed to restart NetworkManager: %v", err)
	}

	logging.Debug("NetworkManager configuration fixed")
	return true, nil
}
func checkNetworkManagerInstalled() (bool, error) {
	logging.Debug("Checking if 'nmcli' is available")
	path, err := exec.LookPath("nmcli")
	if err != nil {
		return false, errors.New("NetworkManager cli nmcli was not found in path")
	}
	logging.Debug("'nmcli' was found in ", path)
	return true, nil
}
func fixNetworkManagerInstalled() (bool, error) {
	return false, fmt.Errorf("NetworkManager is required and must be installed manually")
}
func CheckNetworkManagerIsRunning() (bool, error) {
	logging.Debug("Checking if NetworkManager.service is running")
	path, err := exec.LookPath("systemctl")
	if err != nil {
		return false, err
	}
	stdOut, stdErr, err := crcos.RunWithDefaultLocale(path, "is-active", "NetworkManager")
	if err != nil {
		return false, fmt.Errorf("%v : %s", err, stdErr)
	}
	if strings.TrimSpace(stdOut) != "active" {
		return false, errors.New("NetworkManager.service is not running")
	}
	logging.Debug("NetworkManager.service is already running")
	return true, nil
}
func fixNetworkManagerIsRunning() (bool, error) {
	return false, fmt.Errorf("NetworkManager is required. Please make sure it is installed and running manually")
}
