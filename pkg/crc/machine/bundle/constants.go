package bundle

import (
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/4.9.5/crc_hyperkit_4.9.5.crcbundle",
				"25dca789468f5b59be039cd26948f800498b9eaf5a26e540b8d9c99bfc40d0fc"),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.1/crc_podman_hyperkit_3.4.1.crcbundle",
				"91c661d3b9417bbf0114790c2352d314f3047f1093f65c30b1eb3a8f29c34583"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile(
				"https://storage.googleapis.com/crc-bundle-github-ci/4.9.5/crc_libvirt_4.9.5.crcbundle",
				"603d87599f4c129683d3a8a11bcd41f240a529b2dd81075755eebf3d2b8cacc6",
			),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.1/crc_podman_libvirt_3.4.1.crcbundle",
				"be090392382b84ae340ea3593acbe359f997c84a229e1c695dfe711726fff2aa"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/4.9.5/crc_hyperv_4.9.5.crcbundle",
				"38ca501ceb4c3a2a58db942022548a2dd9d05af6558b416e680ba81463fab3ef"),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.1/crc_podman_hyperv_3.4.1.crcbundle",
				"6edd0cbf0c7e1a1211bc3d4239ff62d5c8acdf436e13f05fd8495cbb450eea0a"),
		},
	},
}
