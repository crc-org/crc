package bundle

import (
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.1/crc_vfkit_4.11.1_amd64.crcbundle",
				"b13ee71cc819b01e4d303e4f25defa2cf33f2dc4106d1bd1f904b80f74a32442"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_vfkit_4.1.0_amd64.crcbundle",
				"81028e2373ba4d541ef926d2ba819b39b79d7c7993d7e66010fa0b2bec270670"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.1/crc_libvirt_4.11.1_amd64.crcbundle",
				"ac4f9e6a1f62ae71a191764e95d46e6141994906659ee9d028f2d620bd9c9792"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_libvirt_4.1.0_amd64.crcbundle",
				"59d00867bf8358a6a3d77abbd3a0606e4a0c9917f5b02ec6d0b08a08a0285ced"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.1/crc_hyperv_4.11.1_amd64.crcbundle",
				"a92d698ca6c1b885bbd42dd11afbf825461ee802a461702ba6e0627ccf0f6b9e"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_hyperv_4.1.0_amd64.crcbundle",
				"cfd4aacbc0b45459af5d8f1dc916db29b6b2010b24153d0050e6d8a9b515b106"),
		},
	},
	"arm64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.1/crc_vfkit_4.11.1_arm64.crcbundle",
				"2ef2f532aa45f9f17385705f2fdd5f0c56ce6650f5e29203e7eb28277f4ad623"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_vfkit_4.1.0_arm64.crcbundle",
				"6674c016591ee56451741de754b14b754aaaf1a981e4a1fa922c238353988e9a"),
		},
	},
}
