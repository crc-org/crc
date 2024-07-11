//go:build linux || build
// +build linux build

package libvirt

import (
	"fmt"
	"runtime"

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
	MachineDriverVersion = "0.13.8"
)

var (
	machineDriverCommand     = fmt.Sprintf("crc-driver-libvirt-%s", runtime.GOARCH)
	MachineDriverDownloadURL = fmt.Sprintf("https://github.com/crc-org/machine-driver-libvirt/releases/download/%s/%s", MachineDriverVersion, machineDriverCommand)
)

func MachineDriverPath() string {
	return constants.ResolveHelperPath(machineDriverCommand)
}
