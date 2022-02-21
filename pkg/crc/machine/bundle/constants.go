package bundle

import (
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.9/crc_vfkit_4.10.9_amd64.crcbundle",
				"720655995a76de4af537d819dbb03d9b88450c601fbfbd4d535812e67f29bb19"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/3.4.4/crc_podman_vfkit_3.4.4_amd64.crcbundle",
				"91836b50c1b8e6b2141ed7560d3b38ae1e24e2d63bc70fe52547aa894afb09d6"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.9/crc_libvirt_4.10.9_amd64.crcbundle",
				"a317681f7e8027eed8aa08a423f6065ef9848bdc617e02cd783d79ff089cbcf9"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/3.4.4/crc_podman_libvirt_3.4.4_amd64.crcbundle",
				"8e8d150dfbefcec93639df53e6181377d1bfc0df3c8719a6fedf0929779f6f63"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.9/crc_hyperv_4.10.9_amd64.crcbundle",
				"b64968be7736d9766e6920d92c5802a7d85247c22bc67cc7cc3bd6887ea967d8"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/3.4.4/crc_podman_hyperv_3.4.4_amd64.crcbundle",
				"e2a55c818b8d2f071f4cc0d26aba11f0a852aebdb7588a08eb96bd41faea0ef4"),
		},
	},
}
