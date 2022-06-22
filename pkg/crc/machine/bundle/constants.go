package bundle

import (
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.18/crc_vfkit_4.10.18_amd64.crcbundle",
				"47a32d7e230201b2b90e41862b99fa94e5844b741faa4d321f3a3c10d8bfd5a3"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.0.2/crc_podman_vfkit_4.0.2_amd64.crcbundle",
				"6a0cb740e02a7f109da4c2b72a48d72da081469b5eba7692a2dc583727fd8e01"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.18/crc_libvirt_4.10.18_amd64.crcbundle",
				"2bb55cf19f4bfce359dd8861fcaf11062d61d57ab1fae7d6a16e0303f7c3d5c9"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.0.2/crc_podman_libvirt_4.0.2_amd64.crcbundle",
				"e8148d317f4aff523f00f2f9a7d520e70eb0c55d466e43847172ddb5d97e3ebf"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.18/crc_hyperv_4.10.18_amd64.crcbundle",
				"de45a0ab0dd16e6a11c839c5d3aa3b6dbf2686a6e16eec5e07de2d8d2ddd6fc1"),
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
