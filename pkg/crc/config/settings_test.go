package config

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	crcpreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/crc/version"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// override for ValidateMemory in validations.go to disable the physical memory check
func validateMemoryNoPhysicalCheck(value interface{}, preset crcpreset.Preset) (bool, string) {
	v, err := cast.ToIntE(value)
	if err != nil {
		return false, fmt.Sprintf("requires integer value in MiB >= %d", constants.GetDefaultMemory(preset))
	}
	if v < constants.GetDefaultMemory(preset) {
		return false, fmt.Sprintf("requires memory in MiB >= %d", constants.GetDefaultMemory(preset))
	}
	return true, ""
}

func newInMemoryConfig() (*Config, error) {
	validateMemory = validateMemoryNoPhysicalCheck
	cfg := New(NewEmptyInMemoryStorage(), NewEmptyInMemorySecretStorage())

	RegisterSettings(cfg)

	return cfg, nil
}

// Check that with the default preset, we cannot set less CPUs than the defaultCPUs
// but that it is allowed with a different preset with less requirements (podman)
func TestCPUsValidate(t *testing.T) {
	cfg, err := newInMemoryConfig()
	require.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     version.GetDefaultPreset().String(),
		Invalid:   false,
		IsDefault: true,
		IsSecret:  false,
	}, cfg.Get(Preset))

	defaultCPUs := constants.GetDefaultCPUs(version.GetDefaultPreset())
	_, err = cfg.Set(CPUs, defaultCPUs-1)
	require.Error(t, err)
	assert.Equal(t, SettingValue{
		Value:     defaultCPUs,
		Invalid:   false,
		IsDefault: true,
		IsSecret:  false,
	}, cfg.Get(CPUs))

	_, err = cfg.Set(Preset, crcpreset.Podman)
	require.NoError(t, err)
	_, err = cfg.Set(CPUs, defaultCPUs-1)
	require.NoError(t, err)
	assert.Equal(t, SettingValue{
		Value:     defaultCPUs - 1,
		Invalid:   false,
		IsDefault: false,
		IsSecret:  false,
	}, cfg.Get(CPUs))
}

// Check that when changing preset, invalid memory values are reset to their
// default value
func TestSetPreset(t *testing.T) {
	cfg, err := newInMemoryConfig()
	require.NoError(t, err)

	_, err = cfg.Set(Preset, crcpreset.Podman)
	require.NoError(t, err)
	_, err = cfg.Set(Memory, 10000)
	require.NoError(t, err)
	assert.Equal(t, SettingValue{
		Value:     10000,
		Invalid:   false,
		IsDefault: false,
		IsSecret:  false,
	}, cfg.Get(Memory))
	_, err = cfg.Set(CPUs, 3)
	require.NoError(t, err)
	assert.Equal(t, SettingValue{
		Value:     3,
		Invalid:   false,
		IsDefault: false,
		IsSecret:  false,
	}, cfg.Get(CPUs))

	// Changing the preset to 'openshift' should reset the CPUs config value
	_, err = cfg.Set(Preset, crcpreset.OpenShift)
	require.NoError(t, err)
	assert.Equal(t, SettingValue{
		Value:     10000,
		Invalid:   false,
		IsDefault: false,
		IsSecret:  false,
	}, cfg.Get(Memory))
	assert.Equal(t, SettingValue{
		Value:     4,
		Invalid:   false,
		IsDefault: true,
		IsSecret:  false,
	}, cfg.Get(CPUs))
}

// Check that when unsetting preset, invalid memory values are reset to their
// default value
func TestUnsetPreset(t *testing.T) {
	cfg, err := newInMemoryConfig()
	require.NoError(t, err)

	_, err = cfg.Set(Preset, crcpreset.Podman)
	require.NoError(t, err)
	_, err = cfg.Set(Memory, 10000)
	require.NoError(t, err)
	assert.Equal(t, SettingValue{
		Value:     10000,
		Invalid:   false,
		IsDefault: false,
		IsSecret:  false,
	}, cfg.Get(Memory))
	_, err = cfg.Set(CPUs, 3)
	require.NoError(t, err)
	assert.Equal(t, SettingValue{
		Value:     3,
		Invalid:   false,
		IsDefault: false,
		IsSecret:  false,
	}, cfg.Get(CPUs))

	// Unsetting the preset should reset the CPUs config value
	_, err = cfg.Unset(Preset)
	require.NoError(t, err)
	assert.Equal(t, SettingValue{
		Value:     crcpreset.OpenShift.String(),
		Invalid:   false,
		IsDefault: true,
		IsSecret:  false,
	}, cfg.Get(Preset))
	assert.Equal(t, SettingValue{
		Value:     10000,
		Invalid:   false,
		IsDefault: false,
		IsSecret:  false,
	}, cfg.Get(Memory))
	assert.Equal(t, SettingValue{
		Value:     4,
		Invalid:   false,
		IsDefault: true,
		IsSecret:  false,
	}, cfg.Get(CPUs))
}

func TestPath(t *testing.T) {
	cfg, err := newInMemoryConfig()
	require.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     Path(""),
		Invalid:   false,
		IsDefault: true,
		IsSecret:  false,
	}, cfg.Get(ProxyCAFile))

	_, err = cfg.Set(ProxyCAFile, "testdata/foo.crt")
	require.NoError(t, err)
	expectedPath, err := filepath.Abs("testdata/foo.crt")
	require.NoError(t, err)
	assert.Equal(t, SettingValue{
		Value:     Path(expectedPath),
		Invalid:   false,
		IsDefault: false,
		IsSecret:  false,
	}, cfg.Get(ProxyCAFile))
}
