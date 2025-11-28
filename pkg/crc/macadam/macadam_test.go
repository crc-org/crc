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

func TestConfigVMOptions(t *testing.T) {
	opts := ConfigVMOptions{
		Name:     "crc",
		CPUs:     uint64(4),
		Memory:   uint64(8192),
		DiskSize: uint64(30),
	}

	assert.Equal(t, "crc", opts.Name)
	assert.Equal(t, uint64(4), opts.CPUs)
	assert.Equal(t, uint64(8192), opts.Memory)
	assert.Equal(t, uint64(30), opts.DiskSize)
}
