package preflight

import (
	"testing"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/network"
	crcos "github.com/code-ready/crc/pkg/os/linux"
	"github.com/stretchr/testify/assert"
)

func TestCountConfigurationOptions(t *testing.T) {
	cfg := config.New(config.NewEmptyInMemoryStorage())
	RegisterSettings(cfg)
	options := len(cfg.AllConfigs())
	assert.True(t, options == 38 || options == 30)
}

func TestCountPreflights(t *testing.T) {
	assert.Len(t, getPreflightChecksForDistro(crcos.RHEL, network.DefaultMode), 19)
	assert.Len(t, getPreflightChecksForDistro(crcos.RHEL, network.VSockMode), 16)

	assert.Len(t, getPreflightChecksForDistro("unexpected", network.DefaultMode), 19)
	assert.Len(t, getPreflightChecksForDistro("unexpected", network.VSockMode), 16)

	assert.Len(t, getPreflightChecksForDistro(crcos.Ubuntu, network.DefaultMode), 15)
	assert.Len(t, getPreflightChecksForDistro(crcos.Ubuntu, network.VSockMode), 16)
}
