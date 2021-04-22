package api

import (
	"strings"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/preflight"
)

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

type mockLogger struct {
}

func (*mockLogger) Messages() []string {
	return []string{"message 1", "message 2", "message 3"}
}
