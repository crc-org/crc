package bundle

import (
	"github.com/crc-org/crc/pkg/crc/preset"
	"github.com/crc-org/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.13/crc_vfkit_4.11.13_amd64.crcbundle",
				"5e86e132e9a2d441dbbfa696c4f3256ade2018de66391d932f560bd5afa1415a"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.13/crc_libvirt_4.11.13_amd64.crcbundle",
				"f5e3fb8be91dff8b573edcf915d0234314199a8547316ebf1b9bd467121c88db"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.13/crc_hyperv_4.11.13_amd64.crcbundle",
				"e622db3f985d42e29c1029f7f7d02997feac2a13930faa2ed6580139570f4d30"),
		},
	},
	"arm64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.11.13/crc_vfkit_4.11.13_arm64.crcbundle",
				"86ce28ad5acf79e9eb2bd189d9059056fb1bb9689326abaf1db83aa456427e2c"),
		},
	},
}
