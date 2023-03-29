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
			preset.Microshift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/microshift/4.12.9/crc_microshift_vfkit_4.12.9_amd64.crcbundle",
				"8edf2df61e6b310f633bb06a658c340a7e94d1f304c6358e7020f43daac9c3dc"),
		},
		"linux": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.9/crc_libvirt_4.12.9_amd64.crcbundle",
				"f57ab331ad092d8cb1f354b4308046c5ffd15bd143b19f841cb64b0fda89db67"),
			preset.Microshift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/microshift/4.12.9/crc_microshift_libvirt_4.12.9_amd64.crcbundle",
				"8e2ef526bfe7973642e75cdb167d9122ae8b549eb2bb2d70432e607e1240bf3b"),
		},
		"windows": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.9/crc_hyperv_4.12.9_amd64.crcbundle",
				"a8267d09eac58e3c7f0db093f3cba83091390e5ac623b0a0282f6f55102b7681"),
			preset.Microshift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/microshift/4.12.9/crc_microshift_hyperv_4.12.9_amd64.crcbundle",
				"f5ac6a032b50e340122d790cfec7bebcda7bb74d26fbd9479a5d001b2809beb7"),
		},
	},
	"arm64": {
		"darwin": {
			preset.OpenShift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/4.12.9/crc_vfkit_4.12.9_arm64.crcbundle",
				"412d20e4969e872c24b14e55cbaa892848a1657b95a20f4af8ad4629ffdf73ab"),
			preset.Microshift: download.NewRemoteFile("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/microshift/4.12.9/crc_microshift_vfkit_4.12.9_arm64.crcbundle",
				"13fd3074632f016ae6a98bc09eb069020d7a817b3d76f7b62be0b4bc6dee4a9e"),
		},
	},
}
