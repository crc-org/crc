package bundle

import (
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/4.9.12/crc_hyperkit_4.9.12.crcbundle",
				"73167e735337b24efa18654667d8bff7cb2b7eb2d0bd2c6ffae4f1df4d8f6ced"),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.2/crc_podman_hyperkit_3.4.2.crcbundle",
				"763078544559b6981bbd7ba5ee946ace30e9cbd9bf37e88913eb9239ed62ad83"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile(
				"https://storage.googleapis.com/crc-bundle-github-ci/4.9.12/crc_libvirt_4.9.12.crcbundle",
				"715fa19aebb36a1b3817554f9b8f6f67a9b5c53045e28b9716388186da2dc03b",
			),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.2/crc_podman_libvirt_3.4.2.crcbundle",
				"45a8ce71a69d39cf6857ecc1bbc00a0704cc67e827db20f4563c0dd510a27b6f"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/4.9.12/crc_hyperv_4.9.12.crcbundle",
				"d5dd89ee500dc1ef9d5af2059f8fd24dc55d07dd5463b57d8de45f29702b9e28"),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.2/crc_podman_hyperv_3.4.2.crcbundle",
				"5542a66975a594346073887f1ea64416144d1f93fa6aa625e556344349c7e30e"),
		},
	},
}
