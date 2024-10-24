package bundle

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/preset"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func jsonForBundle(name string) string {
	return jsonForBundleWithVersion("1.0", name)
}

func jsonForBundleWithVersion(version, name string) string {
	return fmt.Sprintf(`{
  "version": "%s",
  "type": "snc",
  "name": "%s",
  "buildInfo": {
    "buildTime": "2020-10-26T04:48:26+00:00",
    "openshiftInstallerVersion": "./openshift-install v4.6.0\nbuilt from commit ebdbda57fc18d3b73e69f0f2cc499ddfca7e6593\nrelease image registry.svc.ci.openshift.org/origin/release:4.5",
    "sncVersion": "git4.1.14-137-g14e7"
  },
  "clusterInfo": {
    "openshiftVersion": "%s",
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
        "name": "%s",
        "type": "oc-executable",
        "size": "72728632",
        "sha256sum": "983f0883a6dffd601afa663d10161bfd8033fd6d45cf587a9cb22e9a681d6047"
      }
    ]
  },
  "driverInfo": {
    "name": "libvirt"
  }
}`, version, name, openshiftVersion(name), constants.OcExecutableName)
}

func openshiftVersion(name string) string {
	split := strings.Split(name, "_")
	return split[len(split)-1]
}

var parsedReference = CrcBundleInfo{
	Version: "1.0",
	Type:    "snc",
	Name:    "crc_libvirt_4.6.1",
	BuildInfo: BuildInfo{
		BuildTime:                 "2020-10-26T04:48:26+00:00",
		OpenshiftInstallerVersion: "./openshift-install v4.6.0\nbuilt from commit ebdbda57fc18d3b73e69f0f2cc499ddfca7e6593\nrelease image registry.svc.ci.openshift.org/origin/release:4.5",
		SncVersion:                "git4.1.14-137-g14e7",
	},
	ClusterInfo: ClusterInfo{
		OpenShiftVersion:  semver.MustParse("4.6.1"),
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
					Name:     constants.OcExecutableName,
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
	assert.NoError(t, json.Unmarshal([]byte(jsonForBundle("crc_libvirt_4.6.1")), &bundle))
	assert.Equal(t, parsedReference, bundle)
	bin, err := json.Marshal(bundle)
	assert.NoError(t, err)
	assert.JSONEq(t, string(bin), jsonForBundle("crc_libvirt_4.6.1"))
}

// check that the bundle name has the form "crc_libvirt_4.7.8.crcbundle" or "crc_libvirt_4.7.8_123456.crcbundle"
func checkBundleName(t *testing.T, bundleName string) {
	logging.Debugf("Checking bundle '%s", bundleName)
	baseName := GetBundleNameWithoutExtension(constants.GetDefaultBundle(preset.OpenShift))
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
	checkBundleName(t, constants.GetDefaultBundle(preset.OpenShift))
	customBundleName := GetCustomBundleName(constants.GetDefaultBundle(preset.OpenShift))
	checkBundleName(t, customBundleName)
	customBundleName = GetCustomBundleName(customBundleName)
	checkBundleName(t, customBundleName)
}

func TestGetBundleType(t *testing.T) {
	var bundle CrcBundleInfo
	bundle.Type = "okd"
	require.Equal(t, preset.OKD, bundle.GetBundleType())
	bundle.Type = "okd_custom"
	require.Equal(t, preset.OKD, bundle.GetBundleType())
	bundle.Type = "microshift"
	require.Equal(t, preset.Microshift, bundle.GetBundleType())
	bundle.Type = "microshift_custom"
	require.Equal(t, preset.Microshift, bundle.GetBundleType())
	bundle.Type = "openshift"
	require.Equal(t, preset.OpenShift, bundle.GetBundleType())
	bundle.Type = "openshift_custom"
	require.Equal(t, preset.OpenShift, bundle.GetBundleType())
	bundle.Type = "snc"
	require.Equal(t, preset.OpenShift, bundle.GetBundleType())
	bundle.Type = "snc_custom"
	require.Equal(t, preset.OpenShift, bundle.GetBundleType())
	bundle.Type = ""
	require.Equal(t, preset.OpenShift, bundle.GetBundleType())
}

func TestVerifiedHash(t *testing.T) {
	sha256sum, err := getVerifiedHash(testDataURI(t, "sha256sum_correct_4.13.0.txt.sig"), "crc_libvirt_4.13.0_amd64.crcbundle")
	require.NoError(t, err)
	require.Equal(t, "6aad57019aaab95b670378f569b3f4a16398da0358dd1057996453a8d6d92212", sha256sum)

	// sha256sum.txt is unsigned
	_, err = getVerifiedHash(testDataURI(t, "sha256sum.txt"), "crc_libvirt_4.13.0_amd64.crcbundle")
	require.Error(t, err)

	// sha256sum.txt.sig does not contain a sha256sum for fake.crcbundle
	_, err = getVerifiedHash(testDataURI(t, "sha256sum_correct_4.13.0.txt.sig"), "fake.crcbundle")
	require.Error(t, err)

	// the sha256sum file is signed with a GPG key which is not Red Hat's
	_, err = getVerifiedHash(testDataURI(t, "sha256sum_incorrect_4.13.0.txt.sig"), "crc_libvirt_4.13.0_amd64.crcbundle")
	require.ErrorContains(t, err, "signature made by unknown entity")
}

func testDataURI(t *testing.T, sha256sum string) string {
	absPath, err := filepath.Abs(filepath.Join("testdata", sha256sum))
	require.NoError(t, err)
	return fmt.Sprintf("file://%s", filepath.ToSlash(absPath))
}

func TestGetBundleInfoFromNameValid(t *testing.T) {
	valid := [][]string{
		// crc_preset_driver_version_arch_customSuffix.crcbundle
		{"crc_libvirt_4.16.7_amd64.crcbundle", preset.OpenShift.String(), "libvirt", "4.16.7", "amd64", ""},
		{"crc_libvirt_4.16.7_amd64_232.crcbundle", preset.OpenShift.String(), "libvirt", "4.16.7", "amd64", "232"},
		{"crc_microshift_libvirt_4.16.7_amd64.crcbundle", preset.Microshift.String(), "libvirt", "4.16.7", "amd64", ""},
		{"crc_microshift_libvirt_4.16.7_amd64_2345.crcbundle", preset.Microshift.String(), "libvirt", "4.16.7", "amd64", "2345"},
		{"crc_okd_vfkit_4.16.7_amd64.crcbundle", preset.OKD.String(), "vfkit", "4.16.7", "amd64", ""},
		{"crc_okd_vfkit_4.16.7_amd64_2342465234654.crcbundle", preset.OKD.String(), "vfkit", "4.16.7", "amd64", "2342465234654"},
		{"crc_hyperv_4.18.0_arm64.crcbundle", preset.OpenShift.String(), "hyperv", "4.18.0", "arm64", ""},
		{"crc_libvirt_4.18.0-ec.2_amd64.crcbundle", preset.OpenShift.String(), "libvirt", "4.18.0-ec.2", "amd64", ""},

		{"crc_microshift_hyperv_4.18.0_x86.crcbundle", preset.Microshift.String(), "hyperv", "4.18.0", "x86", ""},
		{"crc_microshift_hyperv_4.18.0_x86_1233.crcbundle", preset.Microshift.String(), "hyperv", "4.18.0", "x86", "1233"},
		{"crc_hyperv_4.18.0_x86.crcbundle", preset.OpenShift.String(), "hyperv", "4.18.0", "x86", ""},
		{"crc_hyperv_4.18.0_x86_123.crcbundle", preset.OpenShift.String(), "hyperv", "4.18.0", "x86", "123"},
		{"crc_ABCdrv_4.18.0_x86.crcbundle", preset.OpenShift.String(), "ABCdrv", "4.18.0", "x86", ""},
		{"crc_ABCdrv_4.18.0_x86_4566.crcbundle", preset.OpenShift.String(), "ABCdrv", "4.18.0", "x86", "4566"},
		{"crc_hyperv_4.18.1_64bit.crcbundle", preset.OpenShift.String(), "hyperv", "4.18.1", "64bit", ""},
		{"crc_hyperv_4.1.1-rc.0_64bit.crcbundle", preset.OpenShift.String(), "hyperv", "4.1.1-rc.0", "64bit", ""},

		{"crc_openshift_libvirt_4.16.7_amd64.crcbundle", preset.OpenShift.String(), "libvirt", "4.16.7", "amd64", ""},
		{"crc_openshift_libvirt_4.16.7_amd64_1.crcbundle", preset.OpenShift.String(), "libvirt", "4.16.7", "amd64", "1"},
		{"crc_libvirt_4.16.7_intel.crcbundle", preset.OpenShift.String(), "libvirt", "4.16.7", "intel", ""},
		{"crc_libvirt_4.16.7_intel_23.crcbundle", preset.OpenShift.String(), "libvirt", "4.16.7", "intel", "23"},
		{"crc_libvirt_4.16.7_64.crcbundle", preset.OpenShift.String(), "libvirt", "4.16.7", "64", ""},
		{"crc_libvirt_4.16.7_64_132.crcbundle", preset.OpenShift.String(), "libvirt", "4.16.7", "64", "132"},
		{"crc_microshift_libvirt_4.16.7_64.crcbundle", preset.Microshift.String(), "libvirt", "4.16.7", "64", ""},
		{"crc_microshift_libvirt_4.16.7_64_123.crcbundle", preset.Microshift.String(), "libvirt", "4.16.7", "64", "123"},
		{"crc_okd_libvirt_4.0.0_amd64.crcbundle", preset.OKD.String(), "libvirt", "4.0.0", "amd64", ""},
		{"crc_okd_libvirt_4.0.0_amd64_0123.crcbundle", preset.OKD.String(), "libvirt", "4.0.0", "amd64", "0123"},
	}

	for _, parts := range valid {
		bundleInfo, err := GetBundleInfoFromName(parts[0])
		assert.NoError(t, err)
		assert.Equal(t, bundleInfo.Preset.String(), parts[1])
		assert.Equal(t, bundleInfo.Driver, parts[2])
		assert.Equal(t, bundleInfo.Version, parts[3])
		assert.Equal(t, bundleInfo.Arch, parts[4])
		assert.Equal(t, bundleInfo.CustomBundleSuffix, parts[5])
	}
}

func TestGetBundleInfoFromNameInvalid(t *testing.T) {
	// missing version
	_, err := GetBundleInfoFromName("crc_libvirt_amd64.crcbundle")
	assert.Error(t, err)

	// wrong version format
	_, err = GetBundleInfoFromName("crc_hyperv_4.18_x86.crcbundle")
	assert.Error(t, err)

	// missing crc prefix
	_, err = GetBundleInfoFromName("libvirt_4.16.0_amd64.crcbundle")
	assert.Error(t, err)

	// missing arch
	_, err = GetBundleInfoFromName("crc_microshift_libvirt_4.7.0.crcbundle")
	assert.Error(t, err)

	// non numeric suffix
	_, err = GetBundleInfoFromName("crc_microshift_libvirt_4.7.0_x86_custom.crcbundle")
	assert.Error(t, err)

	// missing driver
	_, err = GetBundleInfoFromName("crc_4.16.2_amd64_123.crcbundle")
	assert.Error(t, err)

	// missing driver and version
	_, err = GetBundleInfoFromName("crc_amd64_123.crcbundle")
	assert.Error(t, err)

	// unknown preset
	_, err = GetBundleInfoFromName("crc_nanoshift_libvirt_4.16.7_amd64_232.crcbundle")
	assert.Error(t, err)
}
