package bundle

import (
	"github.com/crc-org/crc/pkg/crc/preset"
	"github.com/crc-org/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.5/crc_vfkit_4.12.5_amd64.crcbundle",
				"7a691040952b138b200d8bd7dd57ba352ba99f0c4006c976b3660bf7e2d5ccc0"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.5/crc_libvirt_4.12.5_amd64.crcbundle",
				"12797290dda62930f65e74fd12b8d634ef4d265f50ec708563f37b91ad68edc6"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.5/crc_hyperv_4.12.5_amd64.crcbundle",
				"f49945f40c9f219845d2086b66fcc29cbca3b8490bf6319d2a40f4e3e7bd1ad8"),
		},
	},
	"arm64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.5/crc_vfkit_4.12.5_arm64.crcbundle",
				"f89fc8fa21c762e8e483575950bef5b2aa62e575df6ed0fb76051258ef6352e5"),
		},
	},
}
