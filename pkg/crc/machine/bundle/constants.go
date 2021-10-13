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
			"8eda3326dcf6e5476acbdd2705ff6a6bfb5f1dd7768512b885a48a6ac8cfe752",
		),
		"windows": download.NewRemoteFile("", ""),
	},
}
