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
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.1/crc_podman_hyperkit_3.4.1.crcbundle",
				"0b449988906fff01c7fa018c7c163335e95604a1f1feec5828c954bfeb7f47ba"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile(
				"https://storage.googleapis.com/crc-bundle-github-ci/4.9.10/crc_libvirt_4.9.10.crcbundle",
				"9c50de98600eb4275f1fbe0515d9cedfd9e16509b024192c20c8facae328dc3e",
			),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.1/crc_podman_libvirt_3.4.1.crcbundle",
				"e1ecc16473b1bbb9e4e781c6eb5b004fc6e93500adf0e6d5286ae0442253d281"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/4.9.10/crc_hyperv_4.9.10.crcbundle",
				"926fe58e9f364da720c1ac8c301dae293586063a9e08c3f30716af4d81627acc"),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.1/crc_podman_hyperv_3.4.1.crcbundle",
				"9e0b6ac883f942f2e9e6efce60c9664e7247bad985dd44eeb2dbb7f5bb4d6ddd"),
		},
	},
}
