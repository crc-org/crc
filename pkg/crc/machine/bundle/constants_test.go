package bundle

import (
	"fmt"
	"strings"
	"testing"

	"github.com/crc-org/crc/pkg/crc/preset"
	"github.com/crc-org/crc/pkg/crc/version"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBundleVersionURI(t *testing.T) {
	checkBundleVersionURI(t, preset.OpenShift)
	checkBundleVersionURI(t, preset.Microshift)
}

func checkBundleVersionURI(t *testing.T, p preset.Preset) {
	bundleVersion := version.GetBundleVersion(p)
	require.NotEqual(t, bundleVersion, "0.0.0-unset", fmt.Sprintf("%s version is unset (%s), build flags are incorrect", p, bundleVersion))

	uriPrefix := fmt.Sprintf("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/%s/%s", p, bundleVersion)
	for _, osInfo := range bundleLocations {
		for _, presetInfo := range osInfo {
			for preset, remoteFile := range presetInfo {
				if preset == p {
					assert.True(t, strings.HasPrefix(remoteFile.URI, uriPrefix), fmt.Sprintf("URI %s does not match %s version %s", remoteFile.URI, p, bundleVersion))
				}
			}
		}
	}
}
