package bundle

import (
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.3/crc_vfkit_4.11.3_amd64.crcbundle",
				"f912ca39c5243f0e089ba6a25273f4d50c8891fdf43f1f5333ff098022952472"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_vfkit_4.1.0_amd64.crcbundle",
				"81028e2373ba4d541ef926d2ba819b39b79d7c7993d7e66010fa0b2bec270670"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.3/crc_libvirt_4.11.3_amd64.crcbundle",
				"89a31a0571de49c536e6cbd53ec0d3fa0568362c6b53b6b78f85710a6e35fc27"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_libvirt_4.1.0_amd64.crcbundle",
				"59d00867bf8358a6a3d77abbd3a0606e4a0c9917f5b02ec6d0b08a08a0285ced"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.3/crc_hyperv_4.11.3_amd64.crcbundle",
				"acee6b1f4b58e5faf8c6b56c85b565775411ee8c89d986dcc928a17468789166"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_hyperv_4.1.0_amd64.crcbundle",
				"cfd4aacbc0b45459af5d8f1dc916db29b6b2010b24153d0050e6d8a9b515b106"),
		},
	},
	"arm64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.3/crc_vfkit_4.11.3_arm64.crcbundle",
				"290c5544ada1e0e10e5390d349754deabe48a49773c96e92a45359f38eb8f185"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_vfkit_4.1.0_arm64.crcbundle",
				"6674c016591ee56451741de754b14b754aaaf1a981e4a1fa922c238353988e9a"),
		},
	},
}
