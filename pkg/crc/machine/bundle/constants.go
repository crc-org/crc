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
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_vfkit_4.1.0_amd64.crcbundle",
				"81028e2373ba4d541ef926d2ba819b39b79d7c7993d7e66010fa0b2bec270670"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.18/crc_libvirt_4.10.18_amd64.crcbundle",
				"2bb55cf19f4bfce359dd8861fcaf11062d61d57ab1fae7d6a16e0303f7c3d5c9"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_libvirt_4.1.0_amd64.crcbundle",
				"59d00867bf8358a6a3d77abbd3a0606e4a0c9917f5b02ec6d0b08a08a0285ced"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.18/crc_hyperv_4.10.18_amd64.crcbundle",
				"de45a0ab0dd16e6a11c839c5d3aa3b6dbf2686a6e16eec5e07de2d8d2ddd6fc1"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_hyperv_4.1.0_amd64.crcbundle",
				"cfd4aacbc0b45459af5d8f1dc916db29b6b2010b24153d0050e6d8a9b515b106"),
		},
	},
	"arm64": {
		"darwin": {
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_vfkit_4.1.0_arm64.crcbundle",
				"6674c016591ee56451741de754b14b754aaaf1a981e4a1fa922c238353988e9a"),
		},
	},
}
