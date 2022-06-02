package bundle

import (
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.12/crc_vfkit_4.10.12_amd64.crcbundle",
				"99a07b0ac77867365571a40b79d5f8f9fd79acb9ac071def7fe2e99d603ca9a1"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.0.2/crc_podman_vfkit_4.0.2_amd64.crcbundle",
				"6a0cb740e02a7f109da4c2b72a48d72da081469b5eba7692a2dc583727fd8e01"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.12/crc_libvirt_4.10.12_amd64.crcbundle",
				"3081028fbb9da369e5e2aa730d3b65f4f689f7a5b4e6f48952391f696a616e48"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.0.2/crc_podman_libvirt_4.0.2_amd64.crcbundle",
				"e8148d317f4aff523f00f2f9a7d520e70eb0c55d466e43847172ddb5d97e3ebf"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.12/crc_hyperv_4.10.12_amd64.crcbundle",
				"6d17978c8dd678ffefcd46d5aaf8d2057517c018460f1a5ece4e18538bb4bb46"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.0.2/crc_podman_hyperv_4.0.2_amd64.crcbundle",
				"aeaeece60435fc9dd5c10125662416e8976bc567c65f4ca191f542e2d64f9bfd"),
		},
	},
	"arm64": {
		"darwin": {
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.0.2/crc_podman_vfkit_4.0.2_arm64.crcbundle",
				"2c03258fdc75805d9f57e6f1b4d32d82e16a2bddadb3207888e36d9b06ad4008"),
		},
	},
}
