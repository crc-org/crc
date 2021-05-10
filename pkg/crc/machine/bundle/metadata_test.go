package bundle

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const reference = `{
  "version": "1.0",
  "type": "snc",
  "buildInfo": {
    "buildTime": "2020-10-26T04:48:26+00:00",
    "openshiftInstallerVersion": "./openshift-install v4.6.0\nbuilt from commit ebdbda57fc18d3b73e69f0f2cc499ddfca7e6593\nrelease image registry.svc.ci.openshift.org/origin/release:4.5",
    "sncVersion": "git4.1.14-137-g14e7"
  },
  "clusterInfo": {
    "openshiftVersion": "4.6.1",
    "clusterName": "crc",
    "baseDomain": "testing",
    "appsDomain": "apps-crc.testing",
    "sshPrivateKeyFile": "id_ecdsa_crc",
    "kubeConfig": "kubeconfig"
  },
  "nodes": [
    {
      "kind": [
        "master",
        "worker"
      ],
      "hostname": "crc-h66l2-master-0",
      "diskImage": "crc.qcow2",
      "internalIP": "192.168.126.11"
    }
  ],
  "storage": {
    "diskImages": [
      {
        "name": "crc.qcow2",
        "format": "qcow2",
	"size": "9",
        "sha256sum": "245a0e5acd4f09000a9a5f37d731082ed1cf3fdcad1b5320cbe9b153c9fd82a4"
      }
    ],
    "fileList": [
      {
        "name": "oc",
        "type": "oc-executable",
        "size": "72728632",
        "sha256sum": "983f0883a6dffd601afa663d10161bfd8033fd6d45cf587a9cb22e9a681d6047"
      }
    ]
  },
  "driverInfo": {
    "name": "libvirt"
  }
}`

var parsedReference = CrcBundleInfo{
	Version: "1.0",
	Type:    "snc",
	BuildInfo: BuildInfo{
		BuildTime:                 "2020-10-26T04:48:26+00:00",
		OpenshiftInstallerVersion: "./openshift-install v4.6.0\nbuilt from commit ebdbda57fc18d3b73e69f0f2cc499ddfca7e6593\nrelease image registry.svc.ci.openshift.org/origin/release:4.5",
		SncVersion:                "git4.1.14-137-g14e7",
	},
	ClusterInfo: ClusterInfo{
		OpenShiftVersion:  "4.6.1",
		ClusterName:       "crc",
		BaseDomain:        "testing",
		AppsDomain:        "apps-crc.testing",
		SSHPrivateKeyFile: "id_ecdsa_crc",
		KubeConfig:        "kubeconfig",
	}, Nodes: []Node{
		{
			Kind:       []string{"master", "worker"},
			Hostname:   "crc-h66l2-master-0",
			DiskImage:  "crc.qcow2",
			InternalIP: "192.168.126.11",
		},
	},
	Storage: Storage{
		DiskImages: []DiskImage{
			{
				File: File{
					Name:     "crc.qcow2",
					Size:     "9",
					Checksum: "245a0e5acd4f09000a9a5f37d731082ed1cf3fdcad1b5320cbe9b153c9fd82a4",
				},
				Format: "qcow2",
			},
		},
		Files: []FileListItem{
			{
				File: File{
					Name:     "oc",
					Size:     "72728632",
					Checksum: "983f0883a6dffd601afa663d10161bfd8033fd6d45cf587a9cb22e9a681d6047",
				},
				Type: "oc-executable",
			},
		},
	},
	DriverInfo: DriverInfo{
		Name: "libvirt",
	},
	cachedPath: "",
}

func TestUnmarshalMarshal(t *testing.T) {
	var bundle CrcBundleInfo
	assert.NoError(t, json.Unmarshal([]byte(reference), &bundle))
	assert.Equal(t, parsedReference, bundle)
	bin, err := json.Marshal(bundle)
	assert.NoError(t, err)
	assert.JSONEq(t, string(bin), reference)
}

// check that the bundle name has the form "crc_libvirt_4.7.8.crcbundle" or "crc_libvirt_4.7.8_123456.crcbundle"
func checkBundleName(t *testing.T, bundleName string) {
	logging.Debugf("Checking bundle '%s", bundleName)
	baseName := GetBundleNameWithoutExtension(constants.GetDefaultBundle())
	require.True(t, strings.HasPrefix(bundleName, baseName), "%s should start with %s", bundleName, baseName)
	bundleName = bundleName[len(baseName):]
	require.True(t, strings.HasSuffix(bundleName, ".crcbundle"), "%s should have a '.crcbundle' extension", bundleName)
	bundleName = bundleName[:len(bundleName)-len(".crcbundle")]
	if bundleName == "" {
		return
	}
	require.GreaterOrEqual(t, len(bundleName), 2)
	require.Equal(t, rune(bundleName[0]), '_')
	for _, char := range bundleName[1:] {
		require.True(t, unicode.IsDigit(char), "%c should be a digit", char)
	}
}

func TestCustomBundleName(t *testing.T) {
	checkBundleName(t, constants.GetDefaultBundle())
	customBundleName := GetCustomBundleName(constants.GetDefaultBundle())
	checkBundleName(t, customBundleName)
	customBundleName = GetCustomBundleName(customBundleName)
	checkBundleName(t, customBundleName)
}
