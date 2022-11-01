package api

import (
	"strings"

	"github.com/crc-org/crc/pkg/crc/config"
	"github.com/crc-org/crc/pkg/crc/preflight"
)

func setupNewInMemoryConfig() *config.Config {
	storage := config.NewEmptyInMemoryStorage()
	cfg := config.New(&skipPreflights{
		storage: storage,
	}, config.NewEmptyInMemorySecretStorage())
	cfg.AddSetting("a&a", "foo", nil, nil, "test special string")
	cfg.AddSetting("b&&&b", "bar", nil, nil, "test special string")
	config.RegisterSettings(cfg)
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

type mockTelemetry struct {
	actions []string
}

func (m *mockTelemetry) UploadAction(action, _, _ string) error {
	m.actions = append(m.actions, action)
	return nil
}
