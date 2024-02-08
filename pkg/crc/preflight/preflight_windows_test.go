package preflight

import (
	"testing"

	"github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/network"
	"github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/stretchr/testify/assert"
)

func TestCountConfigurationOptions(t *testing.T) {
	cfg := config.New(config.NewEmptyInMemoryStorage(), config.NewEmptyInMemorySecretStorage())
	RegisterSettings(cfg)
	assert.Len(t, cfg.AllConfigs(), 16)
}

func TestCountPreflights(t *testing.T) {
	assert.Len(t, getPreflightChecks(false, network.SystemNetworkingMode, constants.GetDefaultBundlePath(preset.OpenShift), preset.OpenShift, false), 24)
	assert.Len(t, getPreflightChecks(true, network.SystemNetworkingMode, constants.GetDefaultBundlePath(preset.OpenShift), preset.OpenShift, false), 24)

	assert.Len(t, getPreflightChecks(false, network.UserNetworkingMode, constants.GetDefaultBundlePath(preset.OpenShift), preset.OpenShift, false), 23)
	assert.Len(t, getPreflightChecks(true, network.UserNetworkingMode, constants.GetDefaultBundlePath(preset.OpenShift), preset.OpenShift, false), 23)
}
