package bundle

import (
	"github.com/crc-org/crc/pkg/crc/preset"
	"github.com/crc-org/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.13/crc_vfkit_4.12.13_amd64.crcbundle",
				"7f4bbdb80a748ffeb3331edbb604681372af581efde10797c51a3cf3abe3d627"),
			preset.Microshift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/microshift/4.12.13/crc_microshift_vfkit_4.12.13_amd64.crcbundle",
				"ac1fd99e68c41c85e3454e65f379e4f2f5b3f94b1b83164a4d2f585ac6f0e00b"),
		},
		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.13/crc_libvirt_4.12.13_amd64.crcbundle",
				"a5d500a61460f353882e08176e06802cafdc7c21401da25452f551deb13d5e34"),
			preset.Microshift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/microshift/4.12.13/crc_microshift_libvirt_4.12.13_amd64.crcbundle",
				"b8e084b323f1cb5e644c78afe79fe18abdd5dd3eea17560270373f9f22a6e86c"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.13/crc_hyperv_4.12.13_amd64.crcbundle",
				"418a7a6742d8861aa01fbf7b74d5355b4fe05bec290bd3c368569a99c3690556"),
			preset.Microshift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/microshift/4.12.13/crc_microshift_hyperv_4.12.13_amd64.crcbundle",
				"ebcc47729ae0ecc1e6e1a2dd2fa9fc95f29296e38ae981d135c5f8228b059ac0"),
		},
	},
	"arm64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.13/crc_vfkit_4.12.13_arm64.crcbundle",
				"91c488573e413e238b3703730918fbf5e21218597593687e9f2337cc4ff72a81"),
			preset.Microshift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/microshift/4.12.13/crc_microshift_vfkit_4.12.13_arm64.crcbundle",
				"f855c912cf5495e91a6ab5bcc16bbd373bae9964adb31085276886ae5f9221c3"),
		},
	},
}
