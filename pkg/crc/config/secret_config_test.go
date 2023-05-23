package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	password = "password"
	secret   = "secret"
)

func newTestConfigSecret() (*Config, error) {
	cfg := New(NewEmptyInMemoryStorage(), NewEmptyInMemorySecretStorage())

	cfg.AddSetting(password, Secret(""), validateString, SuccessfullyApplied, "")
	cfg.AddSetting(secret, Secret("apples"), validateString, SuccessfullyApplied, "")
	return cfg, nil
}

func TestGetSecret(t *testing.T) {
	cfg, err := newTestConfigSecret()
	require.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     Secret("apples"),
		IsDefault: true,
		IsSecret:  true,
	}, cfg.Get(secret))
}

func TestSecretConfigUnknown(t *testing.T) {
	cfg, err := newTestConfigSecret()
	require.NoError(t, err)

	assert.Equal(t, SettingValue{
		Invalid: true,
	}, cfg.Get("baz"))
}

func TestSecretConfigSetAndGet(t *testing.T) {
	cfg, err := newTestConfigSecret()
	require.NoError(t, err)

	_, err = cfg.Set(password, "pass123")
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     Secret("pass123"),
		IsDefault: false,
		IsSecret:  true,
	}, cfg.Get(password))
}

func TestSecretConfigUnsetAndGet(t *testing.T) {
	cfg, err := newTestConfigSecret()
	require.NoError(t, err)

	_, err = cfg.Unset(secret)
	assert.NoError(t, err)

	assert.Equal(t, SettingValue{
		Value:     Secret("apples"),
		IsDefault: true,
		IsSecret:  true,
	}, cfg.Get(secret))
}
