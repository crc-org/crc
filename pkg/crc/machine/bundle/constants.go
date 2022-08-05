package bundle

import (
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.22/crc_vfkit_4.10.22_amd64.crcbundle",
				"e776618e6529882f8ac5ba17c40143239822d00900e44ff2af3a3fdb25b2f46a"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_vfkit_4.1.0_amd64.crcbundle",
				"81028e2373ba4d541ef926d2ba819b39b79d7c7993d7e66010fa0b2bec270670"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.22/crc_libvirt_4.10.22_amd64.crcbundle",
				"e9d49650626fef8f18755e85df368e9fab466bd11b1f750a61276df98afe0414"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_libvirt_4.1.0_amd64.crcbundle",
				"59d00867bf8358a6a3d77abbd3a0606e4a0c9917f5b02ec6d0b08a08a0285ced"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.22/crc_hyperv_4.10.22_amd64.crcbundle",
				"67cc8816786c37c674224d41597b4dd4831f539ce3e902d6348a9be1b8d12c2e"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_hyperv_4.1.0_amd64.crcbundle",
				"cfd4aacbc0b45459af5d8f1dc916db29b6b2010b24153d0050e6d8a9b515b106"),
		},
	},
	"arm64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/crc_vfkit_4.10.23_arm64.crcbundle",
				"1cd0f693dffc212ae94bca6e3b033897c3fd8546f1fb2bb2425bc8033ea9cee1"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.1.0/crc_podman_vfkit_4.1.0_arm64.crcbundle",
				"6674c016591ee56451741de754b14b754aaaf1a981e4a1fa922c238353988e9a"),
		},
	},
}
