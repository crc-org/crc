package bundle

import (
	"github.com/crc-org/crc/pkg/crc/preset"
	"github.com/crc-org/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.0/crc_vfkit_4.12.0_amd64.crcbundle",
				"d19c80e53f5c593908a09eb9b3f43ebd908db60ca2e54a01d87a8b09c208557f"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.0/crc_libvirt_4.12.0_amd64.crcbundle",
				"cbc75023e63fb33ce4a571ba2047c813acc04f68d6e51879e6a3b238913f54bb"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.0/crc_hyperv_4.12.0_amd64.crcbundle",
				"a780aed82eea3a023b67247598f57ecf16c3d6fd80375cc3e2f1b564d2b9bc71"),
		},
	},
	"arm64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.0/crc_vfkit_4.12.0_arm64.crcbundle",
				"59ec291480daf9e0a92978473b573f06fbb38ff5e1839ee7029aa77d2abea93a"),
		},
	},
}
