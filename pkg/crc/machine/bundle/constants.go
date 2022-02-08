package bundle

import (
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/4.9.15/crc_hyperkit_4.9.15.crcbundle",
				"9e90bcbce5837cfc7c13f94d269fe043ba38c3228b57ba4748cc52d47ededf86"),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.4/crc_podman_hyperkit_3.4.4.crcbundle",
				"c413375524a774149cca563b523f6316fa3af2a0a669a923cce3e707ff5251c7"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile(
				"https://storage.googleapis.com/crc-bundle-github-ci/4.9.15/crc_libvirt_4.9.15.crcbundle",
				"c0af00cbb4eee11913606382dbd5355bd526b850921d5ce8b5be9625b0f2b364",
			),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.4/crc_podman_libvirt_3.4.4.crcbundle",
				"7df37d2292dfe29d58d9dbcbdaabe6b9cc64f339bc370a88169682e1ba16c191"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/4.9.15/crc_hyperv_4.9.15.crcbundle",
				"67dee61a3de54e7c5e35a67fa1552a4a38e867459d67c716556be920cd78114e"),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.4/crc_podman_hyperv_3.4.4.crcbundle",
				"fec1dd6f57cb0ffd74da28648438925d49f517e53088f5c7f13b96a88c7130f1"),
		},
	},
}
