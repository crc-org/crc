//+build darwin build

package darwin

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/constants/common"
	"github.com/code-ready/crc/pkg/crc/version"
)

const (
	crcTrayUrl = "https://github.com/code-ready/tray-macos/releases/download/v%s/crc-tray-macos.tar.gz"
)

var (
	DefaultBundle = fmt.Sprintf("crc_hyperkit_%s.crcbundle", version.GetBundleVersion())
	OcUrl         = fmt.Sprintf("%s/%s", common.OcUrlBase, "macosx/oc.tar.gz")
	PodmanUrl     = fmt.Sprintf("%s/%s", common.PodmanUrlBase, "podman-remote-latest-master-darwin-amd64.zip")
)

func GetCrcTrayUrl() string {
	return fmt.Sprintf(crcTrayUrl, version.GetCRCTrayVersion())
}
