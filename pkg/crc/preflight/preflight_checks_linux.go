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

	"github.com/code-ready/crc/pkg/crc/constants"
	. "github.com/code-ready/crc/pkg/os"
	units "github.com/docker/go-units"
)

const (
	driverBinaryDir      = "/usr/local/bin"
	kvmDriverBinaryPath  = driverBinaryDir + "/docker-machine-driver-kvm"
	kvmDriverDownloadURL = "https://github.com/dhiltgen/docker-machine-kvm/releases/download/v0.10.0/docker-machine-driver-kvm-centos7"
	defaultPoolSize      = "20 GB"
	nodeMac              = "52:fd:fc:07:21:82"
	nodeIp               = "192.168.126.11"
	domName              = "crc"
	poolTemplate         = `<pool type='dir'>
	<name>{{ .PoolName }}</name>
		<target>
			<path>{{ .PoolDir }}</path>
		</target>
</pool>`

	crcNetworkTemplate = `<network>
	<name>crc</name>
	<uuid>49eee855-d342-46c3-9ed3-b8d1758814cd</uuid>
	<forward mode='nat'>
	  <nat>
		<port start='1024' end='65535'/>
	  </nat>
	</forward>
	<bridge name='tt0' stp='on' delay='0'/>
	<mac address='52:54:00:fd:be:d0'/>
	<domain name='test1.tt.testing' localOnly='yes'/>
	<dns>
	  <srv service='etcd-server-ssl' protocol='tcp' domain='test1.tt.testing' target='etcd-0.test1.tt.testing' port='2380' weight='10'/>
	  <host ip='192.168.126.11'>
		<hostname>api.test1.tt.testing</hostname>
		<hostname>etcd-0.test1.tt.testing</hostname>
	  </host>
	</dns>
	<ip family='ipv4' address='192.168.126.1' prefix='24'>
	  <dhcp>
		<host mac='{{ .NodeMac }}' name='{{ .NodeDomName }}' ip='{{ .NodeIP }}'/>
	  </dhcp>
	</ip>
  </network>`
)

type poolXMLAvailableSpace struct {
	Available float64 `xml:"available,unit"`
}

type nodeNetworkConfig struct {
	NodeMac     string
	NodeDomName string
	NodeIP      string
}
type poolConfig struct {
	PoolName string
	PoolDir  string
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
	stdOut, stdErr, err := RunWithPrivilage("yum", "install", "-y", "libvirt", "libvirt-daemon-kvm", "qemu-kvm")
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
	stdOut, stdErr, err := RunWithPrivilage(path, "enable", "libvirtd")
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
	gids, err := currentUser.GroupIds()
	if err != nil {
		return false, err
	}

	libvirtGroup, err := user.LookupGroup("libvirt")
	if err != nil {
		return false, err
	}

	for _, gid := range gids {
		if gid == libvirtGroup.Gid {
			return true, nil
		}
	}
	return false, err
}

func fixUserPartOfLibvirtGroup() (bool, error) {
	// Add user to libvirt/libvirtd group based on distro
	currentUser, err := user.Current()
	if err != nil {
		return false, err
	}
	stdOut, stdErr, err := RunWithPrivilage("usermod", "-a", "-G", "libvirt", currentUser.Username)
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
	stdOut, stdErr, err := RunWithPrivilage(path, "start", "libvirtd")
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

func checkDockerMachineDriverKvmInstalled() (bool, error) {
	// Check if docker-machine-driver-kvm is available
	path, err := exec.LookPath("docker-machine-driver-kvm")
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
		return false, errors.New("docker-machine-driver-kvm do not have correct permissions")
	}
	return true, nil
}

func fixDockerMachineDriverInstalled() (bool, error) {
	// Download the driver binary in /tmp
	tempFilePath := filepath.Join(os.TempDir(), "docker-machine-driver-kvm")
	out, err := os.Create(tempFilePath)
	if err != nil {
		return false, err
	}
	defer out.Close()
	resp, err := http.Get(kvmDriverDownloadURL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return false, err
	}

	stdOut, stdErr, err := RunWithPrivilage("mkdir", "-p", driverBinaryDir)
	if err != nil {
		return false, fmt.Errorf("%s %v: %s", stdOut, err, stdErr)
	}
	stdOut, stdErr, err = RunWithPrivilage("cp", tempFilePath, kvmDriverBinaryPath)
	if err != nil {
		return false, fmt.Errorf("%s %v: %s", stdOut, err, stdErr)
	}
	stdOut, stdErr, err = RunWithPrivilage("chmod", "755", kvmDriverBinaryPath)
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
	config := poolConfig{
		PoolName: constants.PoolName,
		PoolDir:  constants.PoolDir,
	}

	t, err := template.New("poolxml").Parse(poolTemplate)
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
	config := nodeNetworkConfig{
		NodeMac:     constants.NodeMac,
		NodeDomName: constants.DomName,
		NodeIP:      constants.NodeIP,
	}

	t, err := template.New("netxml").Parse(crcNetworkTemplate)
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
