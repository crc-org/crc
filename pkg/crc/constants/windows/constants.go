//+build windows build

package windows

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/constants/common"
	"github.com/code-ready/crc/pkg/crc/version"
)

var (
	DefaultBundle = fmt.Sprintf("crc_hyperv_%s.crcbundle", version.GetBundleVersion())
	OcUrl         = fmt.Sprintf("%s/%s", common.OcUrlBase, "windows/oc.zip")
	PodmanUrl     = fmt.Sprintf("%s/%s", common.PodmanUrlBase, "podman-remote-latest-master-windows-amd64.zip")
)
