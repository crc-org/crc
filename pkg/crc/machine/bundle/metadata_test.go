package bundle

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func jsonForBundle(name string) string {
	return fmt.Sprintf(`{
  "version": "1.0",
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
    "kubeConfig": "kubeconfig",
    "kubeadminPasswordFile": "kubeadmin-password"
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
    ]
  },
  "driverInfo": {
    "name": "libvirt"
  }
}`, name, openshiftVersion(name))
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
		OpenShiftVersion:      "4.6.1",
		ClusterName:           "crc",
		BaseDomain:            "testing",
		AppsDomain:            "apps-crc.testing",
		SSHPrivateKeyFile:     "id_ecdsa_crc",
		KubeConfig:            "kubeconfig",
		KubeadminPasswordFile: "kubeadmin-password",
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
				Name:     "crc.qcow2",
				Format:   "qcow2",
				Size:     "9",
				Checksum: "245a0e5acd4f09000a9a5f37d731082ed1cf3fdcad1b5320cbe9b153c9fd82a4",
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
