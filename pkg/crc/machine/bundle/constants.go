package bundle

import (
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.14/crc_vfkit_4.10.14_amd64.crcbundle",
				"f42e71805a38c28d70318801291ec660e9c33b3c7b1f677a59a9137f8337ccb2"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.0.2/crc_podman_vfkit_4.0.2_amd64.crcbundle",
				"6a0cb740e02a7f109da4c2b72a48d72da081469b5eba7692a2dc583727fd8e01"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.14/crc_libvirt_4.10.14_amd64.crcbundle",
				"49613b2b04d661283b7ed362307728586f8beaab52063fa8b27d2ff0a5cb7014"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.0.2/crc_podman_libvirt_4.0.2_amd64.crcbundle",
				"e8148d317f4aff523f00f2f9a7d520e70eb0c55d466e43847172ddb5d97e3ebf"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.14/crc_hyperv_4.10.14_amd64.crcbundle",
				"7f9c867874ea607e08ad8c2c0e9013f011dfb8df911a1d1b4934e3ff785be2a7"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.0.2/crc_podman_hyperv_4.0.2_amd64.crcbundle",
				"aeaeece60435fc9dd5c10125662416e8976bc567c65f4ca191f542e2d64f9bfd"),
		},
	},
	"arm64": {
		"darwin": {
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.0.2/crc_podman_vfkit_4.0.2_arm64.crcbundle",
				"2c03258fdc75805d9f57e6f1b4d32d82e16a2bddadb3207888e36d9b06ad4008"),
		},
	},
}
