package bundle

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": download.NewRemoteFile("", ""),
		"linux": download.NewRemoteFile(
			fmt.Sprintf("https://storage.googleapis.com/crc-bundle-github-ci/%s/%s", version.GetBundleVersion(), constants.GetDefaultBundle()),
			"603d87599f4c129683d3a8a11bcd41f240a529b2dd81075755eebf3d2b8cacc6",
		),
		"windows": download.NewRemoteFile("", ""),
	},
}
