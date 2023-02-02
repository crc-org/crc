package preflight

import (
	"testing"

	"github.com/crc-org/crc/pkg/crc/config"
	"github.com/crc-org/crc/pkg/crc/constants"
	"github.com/crc-org/crc/pkg/crc/network"
	"github.com/crc-org/crc/pkg/crc/preset"
	"github.com/stretchr/testify/assert"
)

func TestCountConfigurationOptions(t *testing.T) {
	cfg := config.New(config.NewEmptyInMemoryStorage(), config.NewEmptyInMemorySecretStorage())
	RegisterSettings(cfg)
	assert.Len(t, cfg.AllConfigs(), 11)
}

func TestCountPreflights(t *testing.T) {
	assert.Len(t, getPreflightChecks(true, network.SystemNetworkingMode, constants.GetDefaultBundlePath(preset.OpenShift), preset.OpenShift), 16)
	assert.Len(t, getPreflightChecks(true, network.SystemNetworkingMode, constants.GetDefaultBundlePath(preset.OpenShift), preset.OpenShift), 16)

	assert.Len(t, getPreflightChecks(true, network.UserNetworkingMode, constants.GetDefaultBundlePath(preset.OpenShift), preset.OpenShift), 15)
	assert.Len(t, getPreflightChecks(true, network.UserNetworkingMode, constants.GetDefaultBundlePath(preset.OpenShift), preset.OpenShift), 15)
}
