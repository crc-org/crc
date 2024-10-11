package machine

import (
	"os"
	"path/filepath"
	"testing"

	crcConfig "github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/crc-org/crc/v2/pkg/crc/machine/state"
	crcOs "github.com/crc-org/crc/v2/pkg/os"
	"github.com/stretchr/testify/assert"
)

func TestClient_WhenStopInvokedWithNonExistentVM_ThenThrowError(t *testing.T) {
	// Given
	dir := t.TempDir()
	oldKubeConfigEnvVarValue := os.Getenv("KUBECONFIG")
	kubeConfigPath := filepath.Join(dir, "kubeconfig")
	err := crcOs.CopyFile(filepath.Join("testdata", "kubeconfig.in"), kubeConfigPath)
	assert.NoError(t, err)
	err = os.Setenv("KUBECONFIG", kubeConfigPath)
	assert.NoError(t, err)
	crcConfigStorage := crcConfig.New(crcConfig.NewEmptyInMemoryStorage(), crcConfig.NewEmptyInMemorySecretStorage())
	client := NewClient("i-dont-exist", false, crcConfigStorage)

	// When
	clusterState, stopErr := client.Stop()

	// Then
	assert.EqualError(t, stopErr, "Instance is already stopped")
	assert.Equal(t, clusterState, state.Error)
	err = os.Setenv("KUBECONFIG", oldKubeConfigEnvVarValue)
	assert.NoError(t, err)
}

var testArguments = map[string]struct {
	inputKubeConfigPath    string
	expectedKubeConfigPath string
}{
	"When KubeConfig contains crc context, then cleanup KubeConfig": {
		"kubeconfig.in", "kubeconfig.out",
	},
	"When KubeConfig does not contain crc context, then KubeConfig remains unchanged": {
		"kubeconfig.out", "kubeconfig.out",
	},
}

func TestClient_WhenStopInvoked_ThenKubeConfigUpdatedIfRequired(t *testing.T) {
	for name, test := range testArguments {
		t.Run(name, func(t *testing.T) {
			// Given
			dir := t.TempDir()
			oldKubeConfigEnvVarValue := os.Getenv("KUBECONFIG")
			kubeConfigPath := filepath.Join(dir, "kubeconfig")
			err := crcOs.CopyFile(filepath.Join("testdata", test.inputKubeConfigPath), kubeConfigPath)
			assert.NoError(t, err)
			err = os.Setenv("KUBECONFIG", kubeConfigPath)
			assert.NoError(t, err)
			crcConfigStorage := crcConfig.New(crcConfig.NewEmptyInMemoryStorage(), crcConfig.NewEmptyInMemorySecretStorage())
			client := NewClient("test-client", false, crcConfigStorage)

			// When
			clusterState, _ := client.Stop()

			// Then
			actualKubeConfigFile, err := os.ReadFile(kubeConfigPath)
			assert.NoError(t, err)
			expectedKubeConfigPath, err := os.ReadFile(filepath.Join("testdata", test.expectedKubeConfigPath))
			assert.NoError(t, err)
			assert.YAMLEq(t, string(expectedKubeConfigPath), string(actualKubeConfigFile))
			assert.Equal(t, clusterState, state.Error)
			err = os.Setenv("KUBECONFIG", oldKubeConfigEnvVarValue)
			assert.NoError(t, err)
		})
	}
}
