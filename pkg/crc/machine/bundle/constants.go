package bundle

import (
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.12/crc_vfkit_4.10.12_amd64.crcbundle",
				"99a07b0ac77867365571a40b79d5f8f9fd79acb9ac071def7fe2e99d603ca9a1"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/3.4.4/crc_podman_vfkit_3.4.4_amd64.crcbundle",
				"91836b50c1b8e6b2141ed7560d3b38ae1e24e2d63bc70fe52547aa894afb09d6"),
		},

		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.12/crc_libvirt_4.10.12_amd64.crcbundle",
				"3081028fbb9da369e5e2aa730d3b65f4f689f7a5b4e6f48952391f696a616e48"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/3.4.4/crc_podman_libvirt_3.4.4_amd64.crcbundle",
				"8e8d150dfbefcec93639df53e6181377d1bfc0df3c8719a6fedf0929779f6f63"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.10.12/crc_hyperv_4.10.12_amd64.crcbundle",
				"6d17978c8dd678ffefcd46d5aaf8d2057517c018460f1a5ece4e18538bb4bb46"),
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/3.4.4/crc_podman_hyperv_3.4.4_amd64.crcbundle",
				"e2a55c818b8d2f071f4cc0d26aba11f0a852aebdb7588a08eb96bd41faea0ef4"),
		},
	},
	"arm64": {
		"darwin": {
			preset.Podman: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/podman/4.0.2/crc_podman_vfkit_4.0.2_arm64.crcbundle",
				"2c03258fdc75805d9f57e6f1b4d32d82e16a2bddadb3207888e36d9b06ad4008"),
		},
	},
}
