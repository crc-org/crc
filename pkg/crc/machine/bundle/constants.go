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
			preset.Microshift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/microshift/4.12.8/crc_microshift_vfkit_4.12.8_amd64.crcbundle",
				"8f2d144f282c755db56ae354731d5a94e337e4ddb76b44486e2c64f93a230e2a"),
		},
		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.5/crc_libvirt_4.12.5_amd64.crcbundle",
				"12797290dda62930f65e74fd12b8d634ef4d265f50ec708563f37b91ad68edc6"),
			preset.Microshift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/microshift/4.12.8/crc_microshift_libvirt_4.12.8_amd64.crcbundle",
				"b2cd31888817ed426dfc26eb431ce6514ecf6f269526492d1b3f34567a8ed60a"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.5/crc_hyperv_4.12.5_amd64.crcbundle",
				"f49945f40c9f219845d2086b66fcc29cbca3b8490bf6319d2a40f4e3e7bd1ad8"),
			preset.Microshift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/microshift/4.12.8/crc_microshift_hyperv_4.12.8_amd64.crcbundle",
				"2b5d46d03c475beab851470b0aff2955f3e79e46a5e43ac9c5a1fa8a8d9f8c3e"),
		},
	},
	"arm64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.5/crc_vfkit_4.12.5_arm64.crcbundle",
				"f89fc8fa21c762e8e483575950bef5b2aa62e575df6ed0fb76051258ef6352e5"),
			preset.Microshift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/microshift/4.12.8/crc_microshift_vfkit_4.12.8_arm64.crcbundle",
				"448e46774636ae3591365417dff90405c6286b4cb747305c75cfd67864e3e0ad"),
		},
	},
}
