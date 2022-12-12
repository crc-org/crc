package bundle

import (
	"github.com/crc-org/crc/pkg/crc/preset"
	"github.com/crc-org/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.18/crc_vfkit_4.11.18_amd64.crcbundle",
				"dd5e570bd8eac02ef3ff41e38a477fc26e4f86cfa74256f81f5c401ac131193c"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.18/crc_libvirt_4.11.18_amd64.crcbundle",
				"4e0380ad83dfd2c32f8675f79643f2f23aa726b4fad3825d5618ce1f0ed2f6d9"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.18/crc_hyperv_4.11.18_amd64.crcbundle",
				"e2ee24a2128ee8054561aedebb825650369534f9ad962d066dd23a5654f0089b"),
		},
	},
	"arm64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.18/crc_vfkit_4.11.18_arm64.crcbundle",
				"5a3564d28af6aeeacae7c8a2a08441206f976a287acef317961bcea9d390af1c"),
		},
	},
}
