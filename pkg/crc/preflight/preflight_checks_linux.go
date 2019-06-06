package preflight

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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
	"github.com/code-ready/crc/pkg/crc/oc"
	"github.com/code-ready/crc/pkg/crc/systemd"

	crcos "github.com/code-ready/crc/pkg/os"
)

const (
	driverBinaryDir             = "/usr/local/bin"
	libvirtDriverCommand        = "crc-driver-libvirt"
	libvirtDriverVersion        = "0.12.1"
	crcDnsmasqConfigFile        = "crc.conf"
	crcNetworkManagerConfigFile = "crc-nm-dnsmasq.conf"
)

var (
	libvirtDriverBinaryPath = filepath.Join(driverBinaryDir, libvirtDriverCommand)
	crcDnsmasqConfigPath    = filepath.Join(string(filepath.Separator), "etc", "NetworkManager", "dnsmasq.d", crcDnsmasqConfigFile)
	crcDnsmasqConfig        = `server=/crc.testing/192.168.130.1
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
	fi, _ := os.Stat(path)
	if fi.Mode()&os.ModeSymlink != 0 {
		path, err = os.Readlink(path)
		if err != nil {
			return false, errors.New("Libvirt cli virsh was not found in path")
		}
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
	logging.DebugF("Checking if %s is installed", libvirtDriverCommand)
	// Check if crc-driver-libvirt is available
	path, err := exec.LookPath(libvirtDriverCommand)
	if err != nil {
		return false, err
	}
	fi, _ := os.Stat(path)
	// follow symlinks
	if fi.Mode()&os.ModeSymlink != 0 {
		path, err = os.Readlink(path)
		if err != nil {
			return false, err
		}
	}
	// Check if permissions are correct
	if fi.Mode()&0011 == 0 {
		return false, errors.New("crc-driver-libvirt does not have correct permissions")
	}
	// Check the version of driver if it matches to supported one
	stdOut, stdErr, _ := crcos.RunWithDefaultLocale(path, "version")
	if !strings.Contains(stdOut, libvirtDriverVersion) {
		return false, fmt.Errorf("crc-driver-libvirt does not have right version \n Required: %s \n Got: %s use 'crc setup' command.\n %v\n", libvirtDriverVersion, stdOut, stdErr)
	}
	logging.DebugF("%s is already installed in %s", libvirtDriverCommand, path)
	return true, nil
}

func fixMachineDriverLibvirtInstalled() (bool, error) {
	logging.DebugF("Installing %s", libvirtDriverCommand)
	// Download the driver binary in /tmp
	tempFilePath := filepath.Join(os.TempDir(), libvirtDriverCommand)
	logging.DebugF("Downloading %s in %s", libvirtDriverDownloadURL, tempFilePath)
	out, err := os.Create(tempFilePath)
	if err != nil {
		return false, err
	}
	defer out.Close()
	resp, err := http.Get(libvirtDriverDownloadURL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return false, err
	}

	logging.DebugF("Copying %s in %s", libvirtDriverCommand, libvirtDriverBinaryPath)
	stdOut, stdErr, err := crcos.RunWithPrivilege("mkdir", "-p", driverBinaryDir)
	if err != nil {
		return false, fmt.Errorf("%s %v: %s", stdOut, err, stdErr)
	}
	stdOut, stdErr, err = crcos.RunWithPrivilege("cp", tempFilePath, libvirtDriverBinaryPath)
	if err != nil {
		return false, fmt.Errorf("%s %v: %s", stdOut, err, stdErr)
	}
	stdOut, stdErr, err = crcos.RunWithPrivilege("chmod", "755", libvirtDriverBinaryPath)
	if err != nil {
		return false, fmt.Errorf("%s %v: %s", stdOut, err, stdErr)
	}
	logging.DebugF("%s is installed in %s", libvirtDriverCommand, libvirtDriverBinaryPath)
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
		match, err := regexp.MatchString("^crc\\s", stdOut)
		if err != nil {
			return false, err
		}
		if match {
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
	cmd.Run()
	cmd = exec.Command("virsh", "--connect", "qemu:///system", "net-undefine", libvirt.DefaultNetwork)
	cmd.Run()
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
		match, err := regexp.MatchString("^crc\\s", stdOut)
		if err != nil {
			return false, err
		}
		if match && strings.Contains(stdOut, "active") {
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
