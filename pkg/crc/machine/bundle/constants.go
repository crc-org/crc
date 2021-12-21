package bundle

import (
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/4.9.10/crc_hyperkit_4.9.10.crcbundle",
				"c7a4bc040a1a7347eb2ab6947f157f586dd500878f91c481c89be1ab0f55c69a"),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.2/crc_podman_hyperkit_3.4.2.crcbundle",
				"763078544559b6981bbd7ba5ee946ace30e9cbd9bf37e88913eb9239ed62ad83"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile(
				"https://storage.googleapis.com/crc-bundle-github-ci/4.9.10/crc_libvirt_4.9.10.crcbundle",
				"9c50de98600eb4275f1fbe0515d9cedfd9e16509b024192c20c8facae328dc3e",
			),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.2/crc_podman_libvirt_3.4.2.crcbundle",
				"45a8ce71a69d39cf6857ecc1bbc00a0704cc67e827db20f4563c0dd510a27b6f"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/4.9.10/crc_hyperv_4.9.10.crcbundle",
				"926fe58e9f364da720c1ac8c301dae293586063a9e08c3f30716af4d81627acc"),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.2/crc_podman_hyperv_3.4.2.crcbundle",
				"5542a66975a594346073887f1ea64416144d1f93fa6aa625e556344349c7e30e"),
		},
	},
}
