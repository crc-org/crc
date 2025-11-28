package macadam

import (
	"testing"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/stretchr/testify/assert"
)

func TestUseMacadam(t *testing.T) {
	config := UseMacadam()
	assert.NotNil(t, config.Runner)
	assert.Equal(t, constants.MacadamPath(), config.MacadamExecutablePath)
	assert.NotNil(t, config.Env)
	assert.Equal(t, 1, len(config.Env))
}

func TestSetEnv(t *testing.T) {
	config := UseMacadam()
	configWithEnv := config.SetEnv("TEST_VAR", "test_value")

	assert.Equal(t, 2, len(configWithEnv.Env))
	assert.Equal(t, "test_value", configWithEnv.Env["TEST_VAR"])

	// Original config should be unchanged
	assert.Equal(t, 1, len(config.Env))
}

func TestWithEnv(t *testing.T) {
	config := UseMacadam()
	env := map[string]string{
		"VAR1": "value1",
		"VAR2": "value2",
	}

	configWithEnv := config.WithEnv(env)

	assert.Equal(t, 2, len(configWithEnv.Env))
	assert.Equal(t, "value1", configWithEnv.Env["VAR1"])
	assert.Equal(t, "value2", configWithEnv.Env["VAR2"])
}

func TestSetEnvChaining(t *testing.T) {
	config := UseMacadam()
	configWithEnv := config.SetEnv("VAR1", "value1").SetEnv("VAR2", "value2")

	assert.Equal(t, 3, len(configWithEnv.Env))
	assert.Equal(t, "value1", configWithEnv.Env["VAR1"])
	assert.Equal(t, "value2", configWithEnv.Env["VAR2"])
}

func TestVMOptions(t *testing.T) {
	opts := VMOptions{
		DiskImagePath:   "/path/to/disk.qcow2",
		DiskSize:        uint64(31),
		Memory:          uint64(11264),
		Name:            "crc-ng",
		Username:        "core",
		SSHIdentityPath: "/path/to/id_rsa",
		CPUs:            uint64(6),
		CloudInitPath:   "/path/to/cloud-init.yaml",
	}

	assert.Equal(t, "/path/to/disk.qcow2", opts.DiskImagePath)
	assert.Equal(t, uint64(31), opts.DiskSize)
	assert.Equal(t, uint64(11264), opts.Memory)
	assert.Equal(t, "crc-ng", opts.Name)
	assert.Equal(t, "core", opts.Username)
	assert.Equal(t, "/path/to/id_rsa", opts.SSHIdentityPath)
	assert.Equal(t, uint64(6), opts.CPUs)
	assert.Equal(t, "/path/to/cloud-init.yaml", opts.CloudInitPath)
}
