package preflight

import (
	"testing"

	"github.com/code-ready/crc/pkg/crc/config"
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
	assert.Len(t, getPreflightChecksForDistro(crcos.RHEL, false), 20)
	assert.Len(t, getPreflightChecksForDistro(crcos.RHEL, true), 20)

	assert.Len(t, getPreflightChecksForDistro("unexpected", false), 20)
	assert.Len(t, getPreflightChecksForDistro("unexpected", true), 20)

	assert.Len(t, getPreflightChecksForDistro(crcos.Ubuntu, false), 16)
	assert.Len(t, getPreflightChecksForDistro(crcos.Ubuntu, true), 16)
}
