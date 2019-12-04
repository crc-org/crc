package constants

import (
	"fmt"
	"github.com/code-ready/crc/pkg/crc/version"
)

const (
	OcBinaryName = "oc"
)

func GetDefaultBundle() string {
	return fmt.Sprintf("crc_libvirt_%s.crcbundle", version.GetBundleVersion())
}
