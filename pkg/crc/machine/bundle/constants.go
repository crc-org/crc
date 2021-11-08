package bundle

import (
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": download.NewRemoteFile("", ""),
		"linux": download.NewRemoteFile(
			"https://storage.googleapis.com/crc-bundle-github-ci/4.9.5/crc_libvirt_4.9.5.crcbundle",
			"603d87599f4c129683d3a8a11bcd41f240a529b2dd81075755eebf3d2b8cacc6",
		),
		"windows": download.NewRemoteFile("", ""),
	},
}
