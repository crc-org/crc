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

func TestURIOpenShiftVersion(t *testing.T) {
	openshiftVersion := version.GetBundleVersion(preset.OpenShift)
	require.NotEqual(t, openshiftVersion, "0.0.0-unset", fmt.Sprintf("OpenShift version is unset (%s), build flags are incorrect", openshiftVersion))

	uriPrefix := fmt.Sprintf("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift/%s", openshiftVersion)
	for _, osInfo := range bundleLocations {
		for _, presetInfo := range osInfo {
			for _, remoteFile := range presetInfo {
				assert.True(t, strings.HasPrefix(remoteFile.URI, uriPrefix), fmt.Sprintf("URI %s does not match OpenShift version %s", remoteFile.URI, openshiftVersion))
			}
		}
	}
}
