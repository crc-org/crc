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

var (
	rhel crcos.OsRelease = crcos.OsRelease{
		ID:        crcos.RHEL,
		VersionID: "8.2",
	}

	ubuntu crcos.OsRelease = crcos.OsRelease{
		ID:        crcos.Ubuntu,
		VersionID: "20.04",
	}

	unexpected crcos.OsRelease = crcos.OsRelease{
		ID:        "unexpected",
		VersionID: "1234",
	}
)

func TestCountPreflights(t *testing.T) {
	assert.Len(t, getPreflightChecksForDistro(&rhel, network.DefaultMode), 20)
	assert.Len(t, getPreflightChecksForDistro(&rhel, network.VSockMode), 17)

	assert.Len(t, getPreflightChecksForDistro(&unexpected, network.DefaultMode), 20)
	assert.Len(t, getPreflightChecksForDistro(&unexpected, network.VSockMode), 17)

	assert.Len(t, getPreflightChecksForDistro(&ubuntu, network.DefaultMode), 16)
	assert.Len(t, getPreflightChecksForDistro(&ubuntu, network.VSockMode), 17)
}
