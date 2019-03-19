package libvirt

import (
	"github.com/code-ready/crc/pkg/crc/constants"
)

const (
	// Defaults
	DefaultNetwork    = "crc"
	DefaultCacheMode  = "default"
	DefaultIOMode     = "threads"
	DefaultDomainName = constants.DefaultName

	// Static reesources
	PoolName = "crc"
	PoolDir  = "/var/lib/libvirt/images"

	// Static addresses
	MACAddress = "52:fd:fc:07:21:82"
	IPAddress  = "192.168.126.11"
)
