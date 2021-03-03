package preflight

import (
	"testing"

	"github.com/code-ready/crc/pkg/crc/network"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/stretchr/testify/assert"
)

func TestCountConfigurationOptions(t *testing.T) {
	cfg := config.New(config.NewEmptyInMemoryStorage())
	RegisterSettings(cfg)
	assert.Len(t, cfg.AllConfigs(), 9)
}

func TestCountPreflights(t *testing.T) {
	assert.Len(t, getPreflightChecks(false, network.DefaultMode), 14)
	assert.Len(t, getPreflightChecks(true, network.DefaultMode), 20)

	assert.Len(t, getPreflightChecks(false, network.VSockMode), 13)
	assert.Len(t, getPreflightChecks(true, network.VSockMode), 19)
}
