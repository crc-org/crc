package bundle

import (
	"github.com/crc-org/crc/pkg/crc/preset"
	"github.com/crc-org/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.1/crc_vfkit_4.12.1_amd64.crcbundle",
				"b73f2740cfa13d5abbf131907dd7be8ad9a48c3c14b174b23b3970247c474a4d"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.1/crc_libvirt_4.12.1_amd64.crcbundle",
				"e4cdb4ba53590150a02db0c471264120481a84b84578252551264d451d4ca4f7"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.1/crc_hyperv_4.12.1_amd64.crcbundle",
				"c5e53fd81f1ad7b280d80fdca1b8ccba9048faeec1b97115fee64f6913fc01fc"),
		},
	},
	"arm64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.1/crc_vfkit_4.12.1_arm64.crcbundle",
				"594bb86484b92325566b2ac3fd05b5c79b24e24b18c74d4e4e4498385359e6b9"),
		},
	},
}
