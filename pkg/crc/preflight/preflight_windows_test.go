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
	assert.Len(t, cfg.AllConfigs(), 22)
}

func TestCountPreflights(t *testing.T) {
	assert.Len(t, getPreflightChecks(false, network.DefaultMode), 14)
	assert.Len(t, getPreflightChecks(true, network.DefaultMode), 16)
}
