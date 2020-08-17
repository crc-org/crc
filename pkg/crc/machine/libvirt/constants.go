//+build linux build

package libvirt

import "fmt"

const (
	// Defaults
	DefaultNetwork   = "crc"
	DefaultCacheMode = "default"
	DefaultIOMode    = "threads"

	// Static addresses
	MACAddress = "52:fd:fc:07:21:82"
	IPAddress  = "192.168.130.11"
)

const (
	MachineDriverCommand = "crc-driver-libvirt"
	MachineDriverVersion = "0.12.9"
)

var (
	MachineDriverDownloadURL = fmt.Sprintf("https://github.com/code-ready/machine-driver-libvirt/releases/download/%s/crc-driver-libvirt", MachineDriverVersion)
)
