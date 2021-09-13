package cluster

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/stretchr/testify/assert"
	"github.com/zalando/go-keyring"
)

const (
	secret1          = `{"auths":{"quay.io":{"auth":"secret1"}}}` // #nosec G101
	secret2          = `{`
	crcPullSecretEnv = "CRC_TEST_PULL_SECRET_FILE" // #nosec G101
)

var (
	cfg    *config.Config
	loader *nonInteractivePullSecretLoader
	dir    string
)

func TestLoadPullSecret(t *testing.T) {
	_, err := loader.Value()
	assert.Error(t, err)

	assert.Error(t, StoreInKeyring(secret1))
	assert.Error(t, StoreInKeyring(secret2))
}

func TestStoreInKeyringAndLoad(t *testing.T) {
	if os.Getenv(crcPullSecretEnv) == "" {
		t.Skip("Skipping since CRC_TEST_PULL_SECRET_FILE env is not set")
	}
	secret, err := getPullSecretContent()
	assert.NoError(t, err)
	assert.NoError(t, StoreInKeyring(secret))

	val, err := loader.Value()
	assert.NoError(t, err)
	assert.Equal(t, secret, val)

}

func TestLoadFromConfig(t *testing.T) {
	if os.Getenv(crcPullSecretEnv) == "" {
		t.Skip(fmt.Sprintf("Skipping since %s env is not set", crcPullSecretEnv))
	}
	secret, err := getPullSecretContent()
	assert.NoError(t, err)
	assert.NoError(t, ioutil.WriteFile(filepath.Join(dir, "file2"), []byte(secret), 0600))
	_, err = cfg.Set(config.PullSecretFile, filepath.Join(dir, "file2"))
	assert.NoError(t, err)

	val, err := loader.Value()
	assert.NoError(t, err)
	assert.Equal(t, secret, val)
}

func TestMain(m *testing.M) {
	keyring.MockInit()

	dir, err := ioutil.TempDir("", "pull-secret")
	assert.NoError(&testing.T{}, err)

	cfg = config.New(config.NewEmptyInMemoryStorage())
	config.RegisterSettings(cfg)

	loader = &nonInteractivePullSecretLoader{
		config: cfg,
		path:   filepath.Join(dir, "file1"),
	}

	code := m.Run()

	defer func() {
		os.RemoveAll(dir)
		os.Exit(code)
	}()
}

func getPullSecretContent() (string, error) {
	secret, err := ioutil.ReadFile(os.Getenv(crcPullSecretEnv))
	return strings.TrimSpace(string(secret)), err
}
