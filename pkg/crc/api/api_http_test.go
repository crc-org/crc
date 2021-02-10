package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/machine/fakemachine"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/stretchr/testify/assert"
)

func TestHTTPApi(t *testing.T) {
	machine := fakemachine.NewClient()
	config := setupNewInMemoryConfig()

	ts := httptest.NewServer(NewMux(config, machine))

	client := http.DefaultClient
	res, err := client.Get(fmt.Sprintf("%s%s", ts.URL, "/version"))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	res, err = client.Post(fmt.Sprintf("%s%s", ts.URL, "/status"), "application/json", strings.NewReader("test"))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)

	res, err = client.Post(fmt.Sprintf("%s%s", ts.URL, "/config/get"), "application/json", strings.NewReader("{\"properties\":[\"cpus\"]}"))
	assert.NoError(t, err)
	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "{\"Error\":\"\",\"Configs\":{\"cpus\":4}}", string(body))

	res, _ = client.Get(fmt.Sprintf("%s%s", ts.URL, "/config/get"))
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func setupNewInMemoryConfig() config.Storage {
	storage := config.NewEmptyInMemoryStorage()
	cfg := config.New(&skipPreflights{
		storage: storage,
	})
	cmdConfig.RegisterSettings(cfg)
	preflight.RegisterSettings(cfg)

	return cfg
}

type skipPreflights struct {
	storage config.RawStorage
}

func (s *skipPreflights) Get(key string) interface{} {
	if strings.HasPrefix(key, "skip-") {
		return "true"
	}
	return s.storage.Get(key)
}

func (s *skipPreflights) Set(key string, value interface{}) error {
	return s.storage.Set(key, value)
}

func (s *skipPreflights) Unset(key string) error {
	return s.storage.Unset(key)
}
