//+build linux build

package linux

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/constants/common"
	"github.com/code-ready/crc/pkg/crc/version"
)

var (
	DefaultBundle = fmt.Sprintf("crc_libvirt_%s.crcbundle", version.GetBundleVersion())
	OcUrl         = fmt.Sprintf("%s/%s", common.OcUrlBase, "linux/oc.tar.gz")
	PodmanUrl     = fmt.Sprintf("%s/%s", common.PodmanUrlBase, "podman-remote-latest-master-linux---amd64.zip")
)
