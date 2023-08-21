package cluster

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/stretchr/testify/assert"
	"github.com/zalando/go-keyring"
)

const (
	secret1 = `{"auths":{"quay.io":{"auth":"secret1"}}}` // #nosec G101
	secret2 = `{"auths":{"quay.io":{"auth":"secret2"}}}` // #nosec G101
	secret3 = `{"auths":{"quay.io":{"auth":"secret3"}}}` // #nosec G101
	secret4 = `{`
)

func TestLoadPullSecret(t *testing.T) {
	keyring.MockInit()

	dir := t.TempDir()

	cfg := config.New(config.NewEmptyInMemoryStorage(), config.NewEmptyInMemorySecretStorage())
	config.RegisterSettings(cfg)

	loader := &nonInteractivePullSecretLoader{
		config: cfg,
		path:   filepath.Join(dir, "file1"),
	}

	_, err := loader.Value()
	assert.Error(t, err)

	assert.NoError(t, StoreInKeyring(secret3))

	val, err := loader.Value()
	assert.NoError(t, err)
	assert.Equal(t, secret3, val)

	assert.NoError(t, os.WriteFile(filepath.Join(dir, "file2"), []byte(secret2), 0600))
	_, err = cfg.Set(config.PullSecretFile, filepath.Join(dir, "file2"))
	assert.NoError(t, err)

	val, err = loader.Value()
	assert.NoError(t, err)
	assert.Equal(t, secret2, val)

	assert.NoError(t, os.WriteFile(filepath.Join(dir, "file2"), []byte(secret1), 0600))

	val, err = loader.Value()
	assert.NoError(t, err)
	assert.Equal(t, secret1, val)

	assert.Error(t, StoreInKeyring(secret4))
}
