package bundle

import (
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.6/crc_hyperkit_4.10.6_amd64.crcbundle",
				"aad9bad8b18b8d9d8fed70f7b8e83061eb7b6fe84871973b82aafb4b53e9d510"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/3.4.4/crc_podman_hyperkit_3.4.4_amd64.crcbundle",
				"12c0b610e8e3a0d446ac106e2b6f4345ee6fc45304ed05d529933a1261a6dd04"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.6/crc_libvirt_4.10.6_amd64.crcbundle",
				"2292e1a93799130086575938c139fd98e6aa0cd2403591d454860831e9c1edf7"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/3.4.4/crc_podman_libvirt_3.4.4_amd64.crcbundle",
				"8e8d150dfbefcec93639df53e6181377d1bfc0df3c8719a6fedf0929779f6f63"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.6/crc_hyperv_4.10.6_amd64.crcbundle",
				"9212e416a4455ddac2e48aa50d9c5da07bfc18f77e09805a07aa1452605e9839"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/3.4.4/crc_podman_hyperv_3.4.4_amd64.crcbundle",
				"e2a55c818b8d2f071f4cc0d26aba11f0a852aebdb7588a08eb96bd41faea0ef4"),
		},
	},
}
