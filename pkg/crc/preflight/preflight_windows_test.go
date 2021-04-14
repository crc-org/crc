package preflight

import (
	"testing"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/stretchr/testify/assert"
)

func TestCountConfigurationOptions(t *testing.T) {
	cfg := config.New(config.NewEmptyInMemoryStorage())
	RegisterSettings(cfg)
	assert.Len(t, cfg.AllConfigs(), 12)
}

func TestCountPreflights(t *testing.T) {
	assert.Len(t, getPreflightChecks(false, false, network.DefaultMode), 17)
	assert.Len(t, getPreflightChecks(true, true, network.DefaultMode), 20)

	assert.Len(t, getPreflightChecks(false, false, network.VSockMode), 18)
	assert.Len(t, getPreflightChecks(true, true, network.VSockMode), 21)
}
