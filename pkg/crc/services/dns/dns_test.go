package dns

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/crc-org/crc/v2/pkg/crc/machine/bundle"
	"github.com/crc-org/crc/v2/pkg/crc/services"
	"github.com/stretchr/testify/assert"
)

func TestGetApplicableHostnames(t *testing.T) {
	// Given
	bundleMetadata := services.ServicePostStartConfig{
		BundleMetadata: bundle.CrcBundleInfo{
			ClusterInfo: bundle.ClusterInfo{
				OpenShiftVersion:  semver.MustParse("4.6.1"),
				ClusterName:       "crc",
				BaseDomain:        "testing",
				AppsDomain:        "apps.crc.testing",
				SSHPrivateKeyFile: "id_ecdsa_crc",
				KubeConfig:        "kubeconfig",
			},
		},
	}
	// When
	hostnames := getApplicableHostnames(bundleMetadata)
	// Then
	assert.Equal(t, []string{
		"api.crc.testing",
		"host.crc.testing",
		"oauth-openshift.apps.crc.testing",
		"console-openshift-console.apps.crc.testing",
		"downloads-openshift-console.apps.crc.testing",
		"canary-openshift-ingress-canary.apps.crc.testing",
		"default-route-openshift-image-registry.apps.crc.testing",
	}, hostnames)
}
