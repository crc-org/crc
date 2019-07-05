package constants

import (
	"fmt"
	"github.com/code-ready/crc/pkg/crc/version"
)

const (
	OcBinaryName = "oc.exe"
	DefaultOcURL = "https://mirror.openshift.com/pub/openshift-v4/clients/oc/latest/windows/oc.zip"
)

func GetDefaultBundle() string {
	// TODO: we will change once we have correct bundle for windows
	return fmt.Sprintf("crc_hyperv_%s.tar.xz", version.GetBundleVersion())
}
