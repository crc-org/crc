package bundle

import (
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.0/crc_vfkit_4.11.0_amd64.crcbundle",
				"981db5a96f7ce1fa4fdb01c4e5d674afe2961393293c47001b364eef357fd8c0"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_vfkit_4.1.0_amd64.crcbundle",
				"81028e2373ba4d541ef926d2ba819b39b79d7c7993d7e66010fa0b2bec270670"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.0/crc_libvirt_4.11.0_amd64.crcbundle",
				"454c30ff5ef3dd8039d9dee962a7d8ab52daae48b68535fb8a627ba889116797"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_libvirt_4.1.0_amd64.crcbundle",
				"59d00867bf8358a6a3d77abbd3a0606e4a0c9917f5b02ec6d0b08a08a0285ced"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.0/crc_hyperv_4.11.0_amd64.crcbundle",
				"47ccf04b522de844d231f39d7ed0c9ed2aaa1beb88fca8ab1419d337fcac3812"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_hyperv_4.1.0_amd64.crcbundle",
				"cfd4aacbc0b45459af5d8f1dc916db29b6b2010b24153d0050e6d8a9b515b106"),
		},
	},
	"arm64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.0/crc_vfkit_4.11.0_arm64.crcbundle",
				"e3abbf8accd3263d9be8ede49e673d7726522d83478d7aa81f867231efd45eae"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_vfkit_4.1.0_arm64.crcbundle",
				"6674c016591ee56451741de754b14b754aaaf1a981e4a1fa922c238353988e9a"),
		},
	},
}
