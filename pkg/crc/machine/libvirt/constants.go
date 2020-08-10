//+build linux build

package libvirt

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
	MachineDriverVersion = "0.12.8"
)

var (
	MachineDriverDownloadURL = "file://./out/linux-amd64/crc-driver-libvirt"
)
