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
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.4/crc_podman_hyperkit_3.4.4.crcbundle",
				"c413375524a774149cca563b523f6316fa3af2a0a669a923cce3e707ff5251c7"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile(
				"https://storage.googleapis.com/crc-bundle-github-ci/4.9.12/crc_libvirt_4.9.12.crcbundle",
				"715fa19aebb36a1b3817554f9b8f6f67a9b5c53045e28b9716388186da2dc03b",
			),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.4/crc_podman_libvirt_3.4.4.crcbundle",
				"7df37d2292dfe29d58d9dbcbdaabe6b9cc64f339bc370a88169682e1ba16c191"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/4.9.12/crc_hyperv_4.9.12.crcbundle",
				"d5dd89ee500dc1ef9d5af2059f8fd24dc55d07dd5463b57d8de45f29702b9e28"),
			preset.Podman: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/3.4.4/crc_podman_hyperv_3.4.4.crcbundle",
				"fec1dd6f57cb0ffd74da28648438925d49f517e53088f5c7f13b96a88c7130f1"),
		},
	},
}
