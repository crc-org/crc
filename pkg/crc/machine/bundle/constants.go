package bundle

import (
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.3/crc_vfkit_4.11.3_amd64.crcbundle",
				"f912ca39c5243f0e089ba6a25273f4d50c8891fdf43f1f5333ff098022952472"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.3/crc_libvirt_4.11.3_amd64.crcbundle",
				"89a31a0571de49c536e6cbd53ec0d3fa0568362c6b53b6b78f85710a6e35fc27"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.3/crc_hyperv_4.11.3_amd64.crcbundle",
				"acee6b1f4b58e5faf8c6b56c85b565775411ee8c89d986dcc928a17468789166"),
		},
	},
	"arm64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.3/crc_vfkit_4.11.3_arm64.crcbundle",
				"290c5544ada1e0e10e5390d349754deabe48a49773c96e92a45359f38eb8f185"),
		},
	},
}
