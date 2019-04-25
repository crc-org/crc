package preflight

import (
	"bytes"
	"encoding/xml"
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

	"github.com/code-ready/crc/pkg/crc/machine/libvirt"

	crcos "github.com/code-ready/crc/pkg/os"
	units "github.com/docker/go-units"
)

const (
	driverBinaryDir          = "/usr/local/bin"
	libvirtDriverCommand     = "crc-driver-libvirt"
	libvirtDriverDownloadURL = "https://github.com/code-ready/machine-driver-libvirt/releases/download/0.9.1/crc-driver-libvirt"
	defaultPoolSize          = "20 GB"
)

var (
	libvirtDriverBinaryPath = filepath.Join(driverBinaryDir, libvirtDriverCommand)
)

type poolXMLAvailableSpace struct {
	Available float64 `xml:"available,unit"`
}

func checkVirtualizationEnabled() (bool, error) {
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
	return true, nil
}

func fixVirtualizationEnabled() (bool, error) {
	return false, errors.New("You need to enable virtualization in BIOS")
}

func checkKvmEnabled() (bool, error) {
	// Check if /dev/kvm exists
	if _, err := os.Stat("/dev/kvm"); os.IsNotExist(err) {
		return false, errors.New("kvm kernel module is not loaded")
	}
	return true, nil
}

func fixKvmEnabled() (bool, error) {
	cmd := exec.Command("modprobe", "kvm")
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("%v : %s", err, buf.String())
	}
	return true, nil
}

func checkLibvirtInstalled() (bool, error) {
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
	return true, nil
}

func fixLibvirtInstalled() (bool, error) {
	stdOut, stdErr, err := crcos.RunWithPrivilege("yum", "install", "-y", "libvirt", "libvirt-daemon-kvm", "qemu-kvm")
	if err != nil {
		return false, fmt.Errorf("Could not install required packages: %s %v: %s", stdOut, err, stdErr)
	}
	return true, nil
}

func checkLibvirtEnabled() (bool, error) {
	// check if libvirt service is enabled
	path, err := exec.LookPath("systemctl")
	if err != nil {
		return false, fmt.Errorf("systemctl not found on path: %s", err.Error())
	}
	cmd := exec.Command(path, "is-enabled", "libvirtd")
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	stdOut, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("%v : %s", err, buf.String())
	}
	if strings.TrimSpace(string(stdOut)) != "enabled" {
		return false, errors.New("Libvirt is not enabled")
	}
	return true, nil
}

func fixLibvirtEnabled() (bool, error) {
	// Start libvirt service
	path, err := exec.LookPath("systemctl")
	if err != nil {
		return false, err
	}
	stdOut, stdErr, err := crcos.RunWithPrivilege(path, "enable", "libvirtd")
	if err != nil {
		return false, fmt.Errorf("%s, %v : %s", stdOut, err, stdErr)
	}
	return true, nil
}

func checkUserPartOfLibvirtGroup() (bool, error) {
	// check if user is part of libvirt group
	currentUser, err := user.Current()
	if err != nil {
		return false, err
	}
	path, err := exec.LookPath("groups")
	if err != nil {
		return false, err
	}
	cmd := exec.Command(path, currentUser.Username)
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	stdOut, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("%+v : %s", err, buf.String())
	}
	if strings.Contains(string(stdOut), "libvirt") {
		return true, nil
	}
	return false, err
}

func fixUserPartOfLibvirtGroup() (bool, error) {
	// Add user to libvirt/libvirtd group based on distro
	currentUser, err := user.Current()
	if err != nil {
		return false, err
	}
	stdOut, stdErr, err := crcos.RunWithPrivilege("usermod", "-a", "-G", "libvirt", currentUser.Username)
	if err != nil {
		return false, fmt.Errorf("%s %v : %s", stdOut, err, stdErr)
	}
	return true, nil
}

func checkLibvirtServiceRunning() (bool, error) {
	path, err := exec.LookPath("systemctl")
	if err != nil {
		return false, err
	}
	cmd := exec.Command(path, "is-active", "libvirtd")
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	stdOut, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("%v : %s", err, buf.String())
	}
	if strings.TrimSpace(string(stdOut)) != "active" {
		return false, errors.New("Libvirt service is not running")
	}
	return true, nil
}

func fixLibvirtServiceRunning() (bool, error) {
	path, err := exec.LookPath("systemctl")
	if err != nil {
		return false, err
	}
	stdOut, stdErr, err := crcos.RunWithPrivilege(path, "start", "libvirtd")
	if err != nil {
		return false, fmt.Errorf("%s %v : %s", stdOut, err, stdErr)
	}
	return true, nil
}

func checkIpForwardingEnabled() (bool, error) {
	cmd := exec.Command("cat", "/proc/sys/net/ipv4/ip_forward")
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	stdOut, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("%v : %s", err, buf.String())
	}
	if strings.TrimSpace(string(stdOut)) != "1" {
		return false, errors.New("IP forwarding is disabled")
	}
	return true, nil
}

func fixIpForwardingEnabled() (bool, error) {
	path, err := exec.LookPath("sh")
	if err != nil {
		return false, err
	}
	cmd := exec.Command(path, "-c", "echo 'net.ipv4.ip_forward = 1' | sudo tee /etc/sysctl.d/99-ipforward.conf")
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		return false, fmt.Errorf("%v : %s", err, buf.String())
	}
	return true, nil
}

func checkMachineDriverLibvirtInstalled() (bool, error) {
	// Check if docker-machine-driver-kvm is available
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
	return true, nil
}

func fixMachineDriverLibvirtInstalled() (bool, error) {
	// Download the driver binary in /tmp
	tempFilePath := filepath.Join(os.TempDir(), libvirtDriverCommand)
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
	return true, nil
}

func checkDefaultPoolAvailable() (bool, error) {
	// Check if the default pool by libvirt is available
	cmd := exec.Command("virsh", "--connect", "qemu:///system", "pool-list")
	//cmd.Env = cmdUtil.ReplaceEnv(os.Environ(), "LC_ALL", "C")
	//cmd.Env = cmdUtil.ReplaceEnv(cmd.Env, "LANG", "C")

	out, err := cmd.Output()
	if err != nil {
		return false, nil
	}

	stdOut := string(out)
	outputSlice := strings.Split(stdOut, "\n")
	for _, l := range outputSlice {
		l = strings.TrimSpace(l)
		match, err := regexp.MatchString("^default\\s", l)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}
	return true, nil
}

func fixDefaultPoolAvailable() (bool, error) {
	config := libvirt.PoolConfig{
		PoolName: libvirt.PoolName,
		Dir:      libvirt.PoolDir,
	}

	t, err := template.New("poolxml").Parse(libvirt.StoragePoolTemplate)
	if err != nil {
		return false, err
	}
	var poolXMLDef strings.Builder
	err = t.Execute(&poolXMLDef, config)
	if err != nil {
		return false, err
	}

	cmd := exec.Command("virsh", "--connect", "qemu:///system", "pool-define", "/dev/stdin")
	cmd.Stdin = strings.NewReader(poolXMLDef.String())
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		return false, fmt.Errorf("%v, %s", err, buf.String())
	}
	return true, nil
}

func checkDefaultPoolHasSufficientSpace() (bool, error) {
	var freeSpace poolXMLAvailableSpace
	// Check free space in default pool
	cmd := exec.Command("virsh", "--connect", "qemu:///system", "pool-dumpxml", "default")
	stdOut, err := cmd.Output()
	if err != nil {
		return false, err
	}
	err = xml.Unmarshal(stdOut, &freeSpace)
	if err != nil {
		return false, err
	}

	size, err := units.FromHumanSize(defaultPoolSize)
	if err != nil {
		return false, nil
	}

	if freeSpace.Available >= float64(size) {
		return true, nil
	}
	return false, fmt.Errorf("Not enough space in default pool, available free space: %s", units.HumanSize(freeSpace.Available))
}

func fixDefaultPoolHasSufficientSpace() (bool, error) {
	return false, errors.New("Increase the size of /var, or free up space by deleting")
}

func checkLibvirtCrcNetworkAvailable() (bool, error) {
	cmd := exec.Command("virsh", "--connect", "qemu:///system", "net-list")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	stdOut := string(out)
	outputSlice := strings.Split(stdOut, "\n")
	for _, stdOut = range outputSlice {
		stdOut = strings.TrimSpace(stdOut)
		match, err := regexp.MatchString("^crc\\s", stdOut)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}
	return false, errors.New("Libvirt network crc not found")
}

func fixLibvirtCrcNetworkAvailable() (bool, error) {
	config := libvirt.NetworkConfig{
		DomainName: libvirt.DefaultDomainName,
		MAC:        libvirt.MACAddress,
		IP:         libvirt.IPAddress,
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
	cmd := exec.Command("virsh", "--connect", "qemu:///system", "net-define", "/dev/stdin")
	cmd.Stdin = strings.NewReader(netXMLDef.String())
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		return false, fmt.Errorf("%v : %s", err, buf.String())
	}
	return true, nil
}

func checkLibvirtCrcNetworkActive() (bool, error) {
	cmd := exec.Command("virsh", "--connect", "qemu:///system", "net-list")
	//cmd.Env = cmdUtil.ReplaceEnv(os.Environ(), "LC_ALL", "C")
	//cmd.Env = cmdUtil.ReplaceEnv(cmd.Env, "LANG", "C")
	stdOutStdError, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	stdOut := string(stdOutStdError)
	outputSlice := strings.Split(stdOut, "\n")

	for _, stdOut = range outputSlice {
		stdOut = strings.TrimSpace(stdOut)
		match, err := regexp.MatchString("^crc\\s", stdOut)
		if err != nil {
			return false, err
		}
		if match && strings.Contains(stdOut, "active") {
			return true, nil
		}
	}
	return false, errors.New("Libvirt crc network is not active")
}

func fixLibvirtCrcNetworkActive() (bool, error) {
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
	return true, nil
}
