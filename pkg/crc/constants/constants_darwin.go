package constants

import (
	"fmt"
	"github.com/code-ready/crc/pkg/crc/version"
)

const (
	OcBinaryName = "oc"
	DefaultOcURL = "https://mirror.openshift.com/pub/openshift-v4/clients/oc/latest/macosx/oc.tar.gz"
)

func GetDefaultBundle() string {
	return fmt.Sprintf("crc_hyperkit_%s.tar.xz", version.GetBundleVersion())
}
