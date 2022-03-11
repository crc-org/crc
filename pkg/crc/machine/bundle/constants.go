package bundle

import (
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.3/crc_hyperkit_4.10.3_amd64.crcbundle",
				"57c8adae49beeb83d7a180aaba962256c2833863e1f59492295841a3e6e3f016"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/3.4.4/crc_podman_hyperkit_3.4.4_amd64.crcbundle",
				"12c0b610e8e3a0d446ac106e2b6f4345ee6fc45304ed05d529933a1261a6dd04"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.3/crc_libvirt_4.10.3_amd64.crcbundle",
				"d15f0171de51f5fe0e15e30e52ea6c03a785d7e70eed484384e83f2bf86d919d"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/3.4.4/crc_podman_libvirt_3.4.4_amd64.crcbundle",
				"8e8d150dfbefcec93639df53e6181377d1bfc0df3c8719a6fedf0929779f6f63"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.3/crc_hyperv_4.10.3_amd64.crcbundle",
				"6361a803c97fa67d702056caf0e52151522c6186ad6048354f4f512b95fef9f8"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/3.4.4/crc_podman_hyperv_3.4.4_amd64.crcbundle",
				"e2a55c818b8d2f071f4cc0d26aba11f0a852aebdb7588a08eb96bd41faea0ef4"),
		},
	},
}
