package bundle

import (
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.9/crc_hyperkit_4.10.9_amd64.crcbundle",
				"feb48ea07f0344b684a7c812786659c6f8fd23142d9d57acea31ab3aac157936"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/3.4.4/crc_podman_hyperkit_3.4.4_amd64.crcbundle",
				"12c0b610e8e3a0d446ac106e2b6f4345ee6fc45304ed05d529933a1261a6dd04"),
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
