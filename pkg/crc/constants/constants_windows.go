package constants

import (
	"fmt"
	"github.com/code-ready/crc/pkg/crc/version"
)

const (
	OcBinaryName = "oc.exe"
)

func GetDefaultBundle() string {
	// TODO: we will change once we have correct bundle for windows
	return fmt.Sprintf("crc_hyperv_%s.crcbundle", version.GetBundleVersion())
}
