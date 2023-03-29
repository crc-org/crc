package bundle

import (
	"github.com/crc-org/crc/pkg/crc/preset"
	"github.com/crc-org/crc/pkg/download"
)

var bundleLocations = map[string]bundlesDownloadInfo{
	"amd64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.9/crc_vfkit_4.12.9_amd64.crcbundle",
				"7e83b6a4c4da6766b6b4981655d4bb38fd8f9da36ef1a5d16d017cd07d6ee7e9"),
			preset.Microshift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/microshift/4.12.5/crc_microshift_vfkit_4.12.5_amd64.crcbundle",
				"18fa7827c5a15318cac75feae2aba40da5ae2c83703301276190b6212e6d93a8"),
		},
		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.9/crc_libvirt_4.12.9_amd64.crcbundle",
				"f57ab331ad092d8cb1f354b4308046c5ffd15bd143b19f841cb64b0fda89db67"),
			preset.Microshift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/microshift/4.12.5/crc_microshift_libvirt_4.12.5_amd64.crcbundle",
				"aa79ee3c88d5855f864096512e78627b153d896b8aa07782799a6cfdfef77322"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.9/crc_hyperv_4.12.9_amd64.crcbundle",
				"a8267d09eac58e3c7f0db093f3cba83091390e5ac623b0a0282f6f55102b7681"),
			preset.Microshift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/microshift/4.12.5/crc_microshift_hyperv_4.12.5_amd64.crcbundle",
				"ed8eba68baba622011ceb9fd08a985097777c53ad941b3493b398c95512542c1"),
		},
	},
	"arm64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.9/crc_vfkit_4.12.9_arm64.crcbundle",
				"412d20e4969e872c24b14e55cbaa892848a1657b95a20f4af8ad4629ffdf73ab"),
			preset.Microshift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/microshift/4.12.5/crc_microshift_vfkit_4.12.5_arm64.crcbundle",
				"97fc903ce29ca4af99c8f834d4dc5b55366244e53375179063bd12111e251a73"),
		},
	},
}
