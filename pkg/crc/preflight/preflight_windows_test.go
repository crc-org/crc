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
	assert.Len(t, getPreflightChecks(false, network.DefaultMode), 16)
	assert.Len(t, getPreflightChecks(true, network.DefaultMode), 19)

	assert.Len(t, getPreflightChecks(false, network.VSockMode), 17)
	assert.Len(t, getPreflightChecks(true, network.VSockMode), 20)
}
