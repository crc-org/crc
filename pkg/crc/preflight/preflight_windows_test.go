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
	assert.Len(t, cfg.AllConfigs(), 24)
}

func TestCountPreflights(t *testing.T) {
	assert.Len(t, getPreflightChecks(false, network.DefaultMode), 13)
	assert.Len(t, getPreflightChecks(true, network.DefaultMode), 15)

	assert.Len(t, getPreflightChecks(false, network.VSockMode), 14)
	assert.Len(t, getPreflightChecks(true, network.VSockMode), 16)
}
