package bundle

import (
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/4.9.18/crc_hyperkit_4.9.18.crcbundle",
				"d354e7c9bfb88abfb822151bf721d1df45a83b4ac0330f3f5c2bb930466ae86d"),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.4-1/crc_podman_hyperkit_3.4.4_amd64.crcbundle",
				"1195daa0bccd6293609673f424c788b7ad153a44829d55168460370fb88920e1"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/4.9.18/crc_libvirt_4.9.18.crcbundle",
				"120d435caad8349c7d7fa5f836f982efc0c0433fc18ccf0247cd0f30b63f66b9"),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.4-1/crc_podman_libvirt_3.4.4_amd64.crcbundle",
				"7db6a63da95a8a9e336fb77f35d63f336c67930b26e7dbdad04a1b214298c511"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/4.9.18/crc_hyperv_4.9.18.crcbundle",
				"009b2e33ec795a6cdb47083425966f37f22f3d37790fedc7fdb927f52cbfb94c"),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.4-1/crc_podman_hyperv_3.4.4_amd64.crcbundle",
				"21329103dfe706b1c772b9c9ff2f78888d2e8783a5e9138f46fc78efa6a4da00"),
		},
	},
}
