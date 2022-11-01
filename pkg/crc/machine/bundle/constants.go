package bundle

import (
	"github.com/crc-org/crc/pkg/crc/preset"
	"github.com/crc-org/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.7/crc_vfkit_4.11.7_amd64.crcbundle",
				"c79a8de58b1c9e5dc9515e3d07da6516d28bf400c5c528d72940b008aaf81460"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.7/crc_libvirt_4.11.7_amd64.crcbundle",
				"b6361b96d2c5c59e84279e3fc21a99bad9c680c3a3fe52b4302848c5f195dbb0"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.7/crc_hyperv_4.11.7_amd64.crcbundle",
				"6cd9d53309dd491139f3b8f1d00e11ccfc347e96726a8df19ff94ad6d951871a"),
		},
	},
	"arm64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.7/crc_vfkit_4.11.7_arm64.crcbundle",
				"4f8f088fcfd709417865e034c3c5d0854510d0737ce8be4117679e6f7b011f0a"),
		},
	},
}
