//go:build linux || build
// +build linux build

package libvirt

import (
	"fmt"
	"path/filepath"

	"github.com/crc-org/crc/pkg/crc/constants"
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
	MachineDriverCommand = "crc-driver-libvirt"
	MachineDriverVersion = "0.13.5"
)

var (
	MachineDriverDownloadURL = fmt.Sprintf("https://github.com/crc-org/machine-driver-libvirt/releases/download/%s/%s", MachineDriverVersion, MachineDriverCommand)
)

func MachineDriverPath() string {
	return filepath.Join(constants.BinDir(), MachineDriverCommand)
}
