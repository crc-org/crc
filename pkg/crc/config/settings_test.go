package config

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/containers/common/pkg/strongunits"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	crcpreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/crc/version"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// override for ValidateMemory in validations.go to disable the physical memory check
func validateMemoryNoPhysicalCheck(value interface{}, preset crcpreset.Preset) (bool, string) {
	valueAsInt, err := cast.ToUintE(value)
	if err != nil {
		return false, fmt.Sprintf("requires integer value in MiB >= %d", constants.GetDefaultMemory(preset))
	}
	memory := strongunits.MiB(valueAsInt)
	if memory < constants.GetDefaultMemory(preset) {
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
// but that it is allowed with a different preset with less requirements (microshift)
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

	_, err = cfg.Set(Preset, crcpreset.Microshift)
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

	_, err = cfg.Set(Preset, crcpreset.Microshift)
	require.NoError(t, err)
	_, err = cfg.Set(Memory, 10800)
	require.NoError(t, err)
	assert.Equal(t, SettingValue{
		Value:     uint(10800),
		Invalid:   false,
		IsDefault: false,
		IsSecret:  false,
	}, cfg.Get(Memory))
	_, err = cfg.Set(CPUs, 3)
	require.NoError(t, err)
	assert.Equal(t, SettingValue{
		Value:     uint(3),
		Invalid:   false,
		IsDefault: false,
		IsSecret:  false,
	}, cfg.Get(CPUs))

	// Changing the preset to 'openshift' should reset the CPUs config value
	_, err = cfg.Set(Preset, crcpreset.OpenShift)
	require.NoError(t, err)
	assert.Equal(t, SettingValue{
		Value:     uint(10800),
		Invalid:   false,
		IsDefault: false,
		IsSecret:  false,
	}, cfg.Get(Memory))
	assert.Equal(t, SettingValue{
		Value:     uint(4),
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

	_, err = cfg.Set(Preset, crcpreset.Microshift)
	require.NoError(t, err)
	_, err = cfg.Set(Memory, 10800)
	require.NoError(t, err)
	assert.Equal(t, SettingValue{
		Value:     uint(10800),
		Invalid:   false,
		IsDefault: false,
		IsSecret:  false,
	}, cfg.Get(Memory))
	_, err = cfg.Set(CPUs, 3)
	require.NoError(t, err)
	assert.Equal(t, SettingValue{
		Value:     uint(3),
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
		Value:     uint(10800),
		Invalid:   false,
		IsDefault: false,
		IsSecret:  false,
	}, cfg.Get(Memory))
	assert.Equal(t, SettingValue{
		Value:     uint(4),
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

func TestWhenInvalidKeySetThenErrorIsThrown(t *testing.T) {
	// Given
	cfg, err := newInMemoryConfig()
	require.NoError(t, err)

	// When + Then
	_, err = cfg.Set("i-dont-exist", "i-should-not-be-set")
	assert.Error(t, err, "Configuration property 'i-dont-exist' does not exist")
}

var configDefaultValuesTestArguments = []struct {
	key          string
	defaultValue interface{}
}{
	{
		KubeAdminPassword, "",
	},
	{
		DeveloperPassword, "developer",
	},
	{
		CPUs, uint(4),
	},
	{
		Memory, uint(10752),
	},
	{
		DiskSize, 31,
	},
	{
		NameServer, "",
	},
	{
		PullSecretFile, "",
	},
	{
		DisableUpdateCheck, false,
	},
	{
		ExperimentalFeatures, false,
	},
	{
		EmergencyLogin, false,
	},
	{
		PersistentVolumeSize, 15,
	},
	{
		HostNetworkAccess, false,
	},
	{
		HTTPProxy, "",
	},
	{
		HTTPSProxy, "",
	},
	{
		NoProxy, "",
	},
	{
		ProxyCAFile, Path(""),
	},
	{
		EnableClusterMonitoring, false,
	},
	{
		ConsentTelemetry, "",
	},
	{
		IngressHTTPPort, 80,
	},
	{
		IngressHTTPSPort, 443,
	},
	{
		EnableBundleQuayFallback, false,
	},
	{
		Preset, "openshift",
	},
}

func TestDefaultKeyValuesSetInConfig(t *testing.T) {
	for _, tt := range configDefaultValuesTestArguments {
		t.Run(tt.key, func(t *testing.T) {
			// Given
			cfg, err := newInMemoryConfig()
			require.NoError(t, err)

			// When + Then
			assert.Equal(t, SettingValue{
				Value:     tt.defaultValue,
				Invalid:   false,
				IsDefault: true,
			}, cfg.Get(tt.key))
		})
	}
}

var configProvidedValuesTestArguments = []struct {
	key           string
	providedValue interface{}
}{
	{
		KubeAdminPassword, "kubeadmin-secret-password",
	},
	{
		DeveloperPassword, "developer-secret-password",
	},
	{
		CPUs, uint(8),
	},
	{
		Memory, uint(21504),
	},
	{
		DiskSize, 62,
	},
	{
		NameServer, "127.0.0.1",
	},
	{
		DisableUpdateCheck, true,
	},
	{
		ExperimentalFeatures, true,
	},
	{
		EmergencyLogin, true,
	},
	{
		PersistentVolumeSize, 20,
	},
	{
		HTTPProxy, "http://proxy-via-http-proxy-property:3128",
	},
	{
		HTTPSProxy, "https://proxy-via-http-proxy-property:3128",
	},
	{
		NoProxy, "http://no-proxy-property:3128",
	},
	{
		EnableClusterMonitoring, true,
	},
	{
		ConsentTelemetry, "yes",
	},
	{
		IngressHTTPPort, 8080,
	},
	{
		IngressHTTPSPort, 6443,
	},
	{
		EnableBundleQuayFallback, true,
	},
	{
		Preset, "microshift",
	},
}

func TestSetProvidedValuesOverrideDefaultValuesInConfig(t *testing.T) {
	for _, tt := range configProvidedValuesTestArguments {
		t.Run(tt.key, func(t *testing.T) {

			// When + Then

			// Given
			cfg, err := newInMemoryConfig()
			require.NoError(t, err)

			// When
			_, err = cfg.Set(tt.key, tt.providedValue)
			require.NoError(t, err)

			// Then
			assert.Equal(t, SettingValue{
				Value:     tt.providedValue,
				Invalid:   false,
				IsDefault: false,
			}, cfg.Get(tt.key))
		})
	}
}
