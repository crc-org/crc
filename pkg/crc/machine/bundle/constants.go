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
			preset.Microshift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/microshift/4.12.6/crc_microshift_vfkit_4.12.6_amd64.crcbundle",
				"ca1c72a9bc6a28b4b808899604eb4826d439d99f727c3ff0bc13fa0adaa4c43f"),
		},
		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.5/crc_libvirt_4.12.5_amd64.crcbundle",
				"12797290dda62930f65e74fd12b8d634ef4d265f50ec708563f37b91ad68edc6"),
			preset.Microshift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/microshift/4.12.6/crc_microshift_libvirt_4.12.6_amd64.crcbundle",
				"ec97707b1ae4fa981a619d595e2d7d970c5a8f8ab81e00cae588a718ded89a98"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.5/crc_hyperv_4.12.5_amd64.crcbundle",
				"f49945f40c9f219845d2086b66fcc29cbca3b8490bf6319d2a40f4e3e7bd1ad8"),
			preset.Microshift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/microshift/4.12.6/crc_microshift_hyperv_4.12.6_amd64.crcbundle",
				"19a1dee43daca0d86a741ad75ce8797b715e01d8175fa63172f938801c54c9c2"),
		},
	},
	"arm64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.5/crc_vfkit_4.12.5_arm64.crcbundle",
				"f89fc8fa21c762e8e483575950bef5b2aa62e575df6ed0fb76051258ef6352e5"),
			preset.Microshift: download.NewRemoteFile("https://storage.googleapis.com/crc-bundle-github-ci/microshift/4.12.6/crc_microshift_vfkit_4.12.6_arm64.crcbundle",
				"c1b1cc4ea86542995c8692977731e415d076adb3b27df8e8b49749fdb09af639"),
		},
	},
}
