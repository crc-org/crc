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
	CPUs       = "cpus"
	NameServer = "nameservers"
)

func newTestConfig(configFile, envPrefix string) (*Config, error) {
	storage, err := NewViperStorage(configFile, envPrefix)
	if err != nil {
		return nil, err
	}
	config := New(storage)
	config.AddSetting(CPUs, 4, ValidateCPUs, RequiresRestartMsg, "")
	config.AddSetting(NameServer, "", ValidateIPAddress, SuccessfullyApplied, "")
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

	_, err = config.Set(CPUs, 5)
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     5,
		IsDefault: false,
	}, config.Get(CPUs))

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

	_, err = config.Unset(CPUs)
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     4,
		IsDefault: true,
	}, config.Get(CPUs))

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

	_, err = config.Set(CPUs, 5)
	require.NoError(t, err)

	config, err = newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     5,
		IsDefault: false,
	}, config.Get(CPUs))
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
	}, config.Get(CPUs))

	_, err = config.Set(CPUs, 4)
	assert.NoError(t, err)

	bin, err := ioutil.ReadFile(configFile)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"cpus":4}`, string(bin))

	assert.Equal(t, SettingValue{
		Value:     4,
		IsDefault: true,
	}, config.Get(CPUs))

	config, err = newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     4,
		IsDefault: true,
	}, config.Get(CPUs))
}

func TestViperConfigBindFlagSet(t *testing.T) {
	dir, err := ioutil.TempDir("", "cfg")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	configFile := filepath.Join(dir, "crc.json")

	storage, err := NewViperStorage(configFile, "CRC")
	require.NoError(t, err)
	config := New(storage)
	config.AddSetting(CPUs, 4, ValidateCPUs, RequiresRestartMsg, "")
	config.AddSetting(NameServer, "", ValidateIPAddress, SuccessfullyApplied, "")

	flagSet := pflag.NewFlagSet("start", pflag.ExitOnError)
	flagSet.IntP(CPUs, "c", 4, "")
	flagSet.StringP(NameServer, "n", "", "")
	flagSet.StringP("extra", "e", "", "")

	_ = storage.BindFlagSet(flagSet)

	assert.Equal(t, SettingValue{
		Value:     4,
		IsDefault: true,
	}, config.Get(CPUs))
	assert.Equal(t, SettingValue{
		Value:     "",
		IsDefault: true,
	}, config.Get(NameServer))

	assert.NoError(t, flagSet.Set(CPUs, "5"))

	assert.Equal(t, SettingValue{
		Value:     5,
		IsDefault: false,
	}, config.Get(CPUs))

	_, err = config.Set(CPUs, "6")
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

	_, err = config.Set(CPUs, "5")
	require.NoError(t, err)

	config, err = newTestConfig(configFile, "CRC")
	require.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     5,
		IsDefault: false,
	}, config.Get(CPUs))

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

	_, err = config.Set(CPUs, "helloworld")
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

	assert.True(t, config.Get(CPUs).Invalid)
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

	_, err = config1.Set(CPUs, 5)
	require.NoError(t, err)

	bin, err := ioutil.ReadFile(configFile)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"cpus":5}`, string(bin))

	assert.Equal(t, SettingValue{
		Value:     5,
		IsDefault: false,
	}, config2.Get(CPUs))
	assert.Equal(t, SettingValue{
		Value:     5,
		IsDefault: false,
	}, config1.Get(CPUs))
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

	_, err = config1.Set(CPUs, 5)
	require.NoError(t, err)

	_, err = config2.Set(NameServer, "1.1.1.1")
	require.NoError(t, err)

	bin, err := ioutil.ReadFile(configFile)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"cpus":5, "nameservers":"1.1.1.1"}`, string(bin))

	assert.Equal(t, 5, config2.Get(CPUs).Value)
	assert.Equal(t, 5, config1.Get(CPUs).Value)

	assert.Equal(t, "1.1.1.1", config2.Get(NameServer).Value)
	assert.Equal(t, "1.1.1.1", config1.Get(NameServer).Value)
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

	_, err = config1.Set(CPUs, 5)
	require.NoError(t, err)

	_, err = config2.Unset(CPUs)
	require.NoError(t, err)

	bin, err := ioutil.ReadFile(configFile)
	assert.NoError(t, err)
	assert.JSONEq(t, `{}`, string(bin))

	assert.Equal(t, 4, config2.Get(CPUs).Value)
	assert.Equal(t, 4, config1.Get(CPUs).Value)
}
