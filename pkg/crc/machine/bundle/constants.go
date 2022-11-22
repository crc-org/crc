package bundle

import (
	"github.com/crc-org/crc/pkg/crc/preset"
	"github.com/crc-org/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.13/crc_vfkit_4.11.13_amd64.crcbundle",
				"bec15ed8f86bd8951940f4e23cbe918594c53accd2dd67fe45b66576bd8de98f"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.13/crc_libvirt_4.11.13_amd64.crcbundle",
				"a9ffe4453be67969e501b97d6c8477907deaf268a4a967045450fca5cbb26b41"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.13/crc_hyperv_4.11.13_amd64.crcbundle",
				"494e770c9a6e64f54148463932dcb16392176c6fc3e625fab46be5b7bb473963"),
		},
	},
	"arm64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.13/crc_vfkit_4.11.13_arm64.crcbundle",
				"dbb6a2c8fbd09b0b26470938a15d97743a49656dfbc4919a128554ff3b06d284"),
		},
	},
}
