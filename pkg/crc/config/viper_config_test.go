package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	cpus       = "cpus"
	nameServer = "nameservers"
)

func newTestConfig(configFile, envPrefix string) (*Config, error) {
	validCPUs := func(value interface{}) (bool, string) {
		return validateCPUs(value, preset.OpenShift)
	}
	storage, err := NewViperStorage(configFile, envPrefix)
	if err != nil {
		return nil, err
	}

	secretStorage := NewEmptyInMemorySecretStorage()
	config := New(storage, secretStorage)
	config.AddSetting(cpus, 4, validCPUs, RequiresRestartMsg, "")
	config.AddSetting(nameServer, "", validateIPAddress, SuccessfullyApplied, "")
	return config, nil
}

func TestViperConfigUnknown(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "crc.json")

	config, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	assert.Equal(t, SettingValue{
		Invalid: true,
	}, config.Get("foo"))
}

func TestViperConfigSetAndGet(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "crc.json")

	config, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	_, err = config.Set(cpus, 5)
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     5,
		IsDefault: false,
	}, config.Get(cpus))

	bin, err := os.ReadFile(configFile)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"cpus":5}`, string(bin))
}

func TestViperConfigUnsetAndGet(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "crc.json")
	assert.NoError(t, os.WriteFile(configFile, []byte("{\"cpus\": 5}"), 0600))

	config, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	_, err = config.Unset(cpus)
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     4,
		IsDefault: true,
	}, config.Get(cpus))

	bin, err := os.ReadFile(configFile)
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(bin))
}

func TestViperConfigSetReloadAndGet(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "crc.json")

	config, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	_, err = config.Set(cpus, 5)
	require.NoError(t, err)

	config, err = newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     5,
		IsDefault: false,
	}, config.Get(cpus))
}

func TestViperConfigLoadDefaultValue(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "crc.json")

	config, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     4,
		IsDefault: true,
	}, config.Get(cpus))
	_, err = config.Set(cpus, 4)
	assert.NoError(t, err)

	bin, err := os.ReadFile(configFile)
	assert.NoError(t, err)
	// Setting default value will not update
	// write to the config so expected would be {}
	assert.JSONEq(t, `{}`, string(bin))

	assert.Equal(t, SettingValue{
		Value:     4,
		IsDefault: true,
	}, config.Get(cpus))

	config, err = newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     4,
		IsDefault: true,
	}, config.Get(cpus))
}

func TestViperConfigBindFlagSet(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "crc.json")

	validCPUs := func(value interface{}) (bool, string) {
		return validateCPUs(value, preset.OpenShift)
	}
	storage, err := NewViperStorage(configFile, "CRC")
	require.NoError(t, err)
	config := New(storage, NewEmptyInMemorySecretStorage())
	config.AddSetting(cpus, 4, validCPUs, RequiresRestartMsg, "")
	config.AddSetting(nameServer, "", validateIPAddress, SuccessfullyApplied, "")

	flagSet := pflag.NewFlagSet("start", pflag.ExitOnError)
	flagSet.IntP(cpus, "c", 4, "")
	flagSet.StringP(nameServer, "n", "", "")
	flagSet.StringP("extra", "e", "", "")

	_ = storage.BindFlagSet(flagSet)

	assert.Equal(t, SettingValue{
		Value:     4,
		IsDefault: true,
	}, config.Get(cpus))
	assert.Equal(t, SettingValue{
		Value:     "",
		IsDefault: true,
	}, config.Get(nameServer))

	assert.NoError(t, flagSet.Set(cpus, "5"))

	assert.Equal(t, SettingValue{
		Value:     5,
		IsDefault: false,
	}, config.Get(cpus))

	_, err = config.Set(cpus, "6")
	assert.NoError(t, err)

	bin, err := os.ReadFile(configFile)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"cpus":6}`, string(bin))
}

func TestViperConfigCastSet(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "crc.json")

	config, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	_, err = config.Set(cpus, "5")
	require.NoError(t, err)

	config, err = newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     5,
		IsDefault: false,
	}, config.Get(cpus))

	bin, err := os.ReadFile(configFile)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"cpus": 5}`, string(bin))
}

func TestCannotSetWithWrongType(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "crc.json")

	config, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	_, err = config.Set(cpus, "helloworld")
	assert.EqualError(t, err, "Value 'helloworld' for configuration property 'cpus' is invalid, reason: requires integer value >= 4")
}

func TestCannotGetWithWrongType(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "crc.json")
	assert.NoError(t, os.WriteFile(configFile, []byte("{\"cpus\": \"hello\"}"), 0600))

	config, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	assert.True(t, config.Get(cpus).Invalid)
}

func TestTwoInstancesSharingSameConfiguration(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "crc.json")

	config1, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	config2, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	_, err = config1.Set(cpus, 5)
	require.NoError(t, err)

	bin, err := os.ReadFile(configFile)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"cpus":5}`, string(bin))

	assert.Equal(t, SettingValue{
		Value:     5,
		IsDefault: false,
	}, config2.Get(cpus))
	assert.Equal(t, SettingValue{
		Value:     5,
		IsDefault: false,
	}, config1.Get(cpus))
}

func TestTwoInstancesWriteSameConfiguration(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "crc.json")

	config1, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	config2, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	_, err = config1.Set(cpus, 5)
	require.NoError(t, err)

	_, err = config2.Set(nameServer, "1.1.1.1")
	require.NoError(t, err)

	bin, err := os.ReadFile(configFile)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"cpus":5, "nameservers":"1.1.1.1"}`, string(bin))

	assert.Equal(t, 5, config2.Get(cpus).Value)
	assert.Equal(t, 5, config1.Get(cpus).Value)

	assert.Equal(t, "1.1.1.1", config2.Get(nameServer).Value)
	assert.Equal(t, "1.1.1.1", config1.Get(nameServer).Value)
}

func TestTwoInstancesSetAndUnsetSameConfiguration(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "crc.json")

	config1, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	config2, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	_, err = config1.Set(cpus, 5)
	require.NoError(t, err)

	_, err = config2.Unset(cpus)
	require.NoError(t, err)

	bin, err := os.ReadFile(configFile)
	assert.NoError(t, err)
	assert.JSONEq(t, `{}`, string(bin))

	assert.Equal(t, 4, config2.Get(cpus).Value)
	assert.Equal(t, 4, config1.Get(cpus).Value)
}

func TestNotifier(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "crc.json")

	config, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	var notified = false
	err = config.RegisterNotifier(CPUs, func(config *Config, key string, value interface{}) {
		notified = true
		require.Equal(t, key, CPUs)
		assert.Equal(t, SettingValue{
			Value:     value,
			IsDefault: false,
		}, config.Get(cpus))

	})
	require.NoError(t, err)
	_, err = config.Set(cpus, 5)
	assert.NoError(t, err)
	require.True(t, notified)
}

func TestCallbacks(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "crc.json")

	config, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	callback, err := config.Set(cpus, 5)
	require.NoError(t, err)
	require.NotEmpty(t, callback)

	callback, err = config.Unset(cpus)
	require.NoError(t, err)
	require.NotEmpty(t, callback)

	callback, err = config.Set(cpus, 4)
	require.NoError(t, err)
	require.NotEmpty(t, callback)

	callback, err = config.Unset(cpus)
	require.NoError(t, err)
	require.NotEmpty(t, callback)
}
