package preflight

import (
	"testing"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/stretchr/testify/assert"
)

func TestCountConfigurationOptions(t *testing.T) {
	cfg := config.New(config.NewEmptyInMemoryStorage())
	RegisterSettings(cfg)
	assert.Len(t, cfg.AllConfigs(), 12)
}

func TestCountPreflights(t *testing.T) {
	assert.Len(t, getPreflightChecks(false, network.SystemNetworkingMode, constants.GetDefaultBundlePath(preset.OpenShift), preset.OpenShift), 18)
	assert.Len(t, getPreflightChecks(true, network.SystemNetworkingMode, constants.GetDefaultBundlePath(preset.OpenShift), preset.OpenShift), 18)

	assert.Len(t, getPreflightChecks(false, network.UserNetworkingMode, constants.GetDefaultBundlePath(preset.OpenShift), preset.OpenShift), 17)
	assert.Len(t, getPreflightChecks(true, network.UserNetworkingMode, constants.GetDefaultBundlePath(preset.OpenShift), preset.OpenShift), 17)
}
