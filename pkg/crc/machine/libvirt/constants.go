//go:build linux || build
// +build linux build

package libvirt

import (
	"fmt"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
)

const (
	// Defaults
	DefaultNetwork     = "crc"
	DefaultStoragePool = "crc"

	// Static addresses
	MACAddress = "52:fd:fc:07:21:82"
	IPAddress  = "192.168.130.11"
)

const (
	machineDriverCommand = "crc-driver-libvirt"
	MachineDriverVersion = "0.13.7"
)

var (
	MachineDriverDownloadURL = fmt.Sprintf("https://github.com/crc-org/machine-driver-libvirt/releases/download/%s/%s", MachineDriverVersion, machineDriverCommand)
)

func MachineDriverPath() string {
	return constants.ResolveHelperPath(machineDriverCommand)
}
