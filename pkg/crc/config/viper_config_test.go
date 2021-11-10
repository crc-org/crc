package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	cpus       = "cpus"
	nameServer = "nameservers"
)

func newTestConfig(configFile, envPrefix string) (*Config, error) {
	validateCPUs := func(value interface{}) (bool, string) {
		return ValidateCPUs(value, true)
	}
	storage, err := NewViperStorage(configFile, envPrefix)
	if err != nil {
		return nil, err
	}
	config := New(storage)
	config.AddSetting(cpus, 4, validateCPUs, RequiresRestartMsg, "")
	config.AddSetting(nameServer, "", ValidateIPAddress, SuccessfullyApplied, "")
	return config, nil
}

func TestViperConfigUnknown(t *testing.T) {
	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	configFile := filepath.Join(dir, "crc.json")

	config, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	assert.Equal(t, SettingValue{
		Invalid: true,
	}, config.Get("foo"))
}

func TestViperConfigSetAndGet(t *testing.T) {
	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	configFile := filepath.Join(dir, "crc.json")

	config, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	_, err = config.Set(cpus, 5)
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     5,
		IsDefault: false,
	}, config.Get(cpus))

	bin, err := ioutil.ReadFile(configFile)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"cpus":5}`, string(bin))
}

func TestViperConfigUnsetAndGet(t *testing.T) {
	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	configFile := filepath.Join(dir, "crc.json")
	assert.NoError(t, ioutil.WriteFile(configFile, []byte("{\"cpus\": 5}"), 0600))

	config, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	_, err = config.Unset(cpus)
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     4,
		IsDefault: true,
	}, config.Get(cpus))

	bin, err := ioutil.ReadFile(configFile)
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(bin))
}

func TestViperConfigSetReloadAndGet(t *testing.T) {
	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
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
	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	configFile := filepath.Join(dir, "crc.json")

	config, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     4,
		IsDefault: true,
	}, config.Get(cpus))

	_, err = config.Set(cpus, 4)
	assert.NoError(t, err)

	bin, err := ioutil.ReadFile(configFile)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"cpus":4}`, string(bin))

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
	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	configFile := filepath.Join(dir, "crc.json")

	validateCPUs := func(value interface{}) (bool, string) {
		return ValidateCPUs(value, true)
	}
	storage, err := NewViperStorage(configFile, "CRC")
	require.NoError(t, err)
	config := New(storage)
	config.AddSetting(cpus, 4, validateCPUs, RequiresRestartMsg, "")
	config.AddSetting(nameServer, "", ValidateIPAddress, SuccessfullyApplied, "")

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

	bin, err := ioutil.ReadFile(configFile)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"cpus":6}`, string(bin))
}

func TestViperConfigCastSet(t *testing.T) {
	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
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

	bin, err := ioutil.ReadFile(configFile)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"cpus": 5}`, string(bin))
}

func TestCannotSetWithWrongType(t *testing.T) {
	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	configFile := filepath.Join(dir, "crc.json")

	config, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	_, err = config.Set(cpus, "helloworld")
	assert.EqualError(t, err, "Value 'helloworld' for configuration property 'cpus' is invalid, reason: requires integer value >= 4")
}

func TestCannotGetWithWrongType(t *testing.T) {
	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	configFile := filepath.Join(dir, "crc.json")
	assert.NoError(t, ioutil.WriteFile(configFile, []byte("{\"cpus\": \"hello\"}"), 0600))

	config, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	assert.True(t, config.Get(cpus).Invalid)
}

func TestTwoInstancesSharingSameConfiguration(t *testing.T) {
	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	configFile := filepath.Join(dir, "crc.json")

	config1, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	config2, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	_, err = config1.Set(cpus, 5)
	require.NoError(t, err)

	bin, err := ioutil.ReadFile(configFile)
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
	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	configFile := filepath.Join(dir, "crc.json")

	config1, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	config2, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	_, err = config1.Set(cpus, 5)
	require.NoError(t, err)

	_, err = config2.Set(nameServer, "1.1.1.1")
	require.NoError(t, err)

	bin, err := ioutil.ReadFile(configFile)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"cpus":5, "nameservers":"1.1.1.1"}`, string(bin))

	assert.Equal(t, 5, config2.Get(cpus).Value)
	assert.Equal(t, 5, config1.Get(cpus).Value)

	assert.Equal(t, "1.1.1.1", config2.Get(nameServer).Value)
	assert.Equal(t, "1.1.1.1", config1.Get(nameServer).Value)
}

func TestTwoInstancesSetAndUnsetSameConfiguration(t *testing.T) {
	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	configFile := filepath.Join(dir, "crc.json")

	config1, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	config2, err := newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	_, err = config1.Set(cpus, 5)
	require.NoError(t, err)

	_, err = config2.Unset(cpus)
	require.NoError(t, err)

	bin, err := ioutil.ReadFile(configFile)
	assert.NoError(t, err)
	assert.JSONEq(t, `{}`, string(bin))

	assert.Equal(t, 4, config2.Get(cpus).Value)
	assert.Equal(t, 4, config1.Get(cpus).Value)
}
