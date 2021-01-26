package cluster

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/stretchr/testify/assert"
)

const (
	secret1 = `{"auths":{"quay.io":{"auth":"secret1"}}}` // #nosec G101
	secret2 = `{"auths":{"quay.io":{"auth":"secret2"}}}` // #nosec G101
)

func TestLoadPullSecret(t *testing.T) {
	dir, err := ioutil.TempDir("", "pull-secret")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	cfg := config.New(config.NewEmptyInMemoryStorage())
	cmdConfig.RegisterSettings(cfg)

	loader := &nonInteractivePullSecretLoader{
		config: cfg,
		path:   filepath.Join(dir, "file1"),
	}

	_, err = loader.Value()
	assert.Error(t, err)

	assert.NoError(t, ioutil.WriteFile(filepath.Join(dir, "file2"), []byte(secret2), 0600))
	_, err = cfg.Set(cmdConfig.PullSecretFile, filepath.Join(dir, "file2"))
	assert.NoError(t, err)

	val, err := loader.Value()
	assert.NoError(t, err)
	assert.Equal(t, secret2, val)

	assert.NoError(t, ioutil.WriteFile(filepath.Join(dir, "file2"), []byte(secret1), 0600))

	val, err = loader.Value()
	assert.NoError(t, err)
	assert.Equal(t, secret1, val)
}
