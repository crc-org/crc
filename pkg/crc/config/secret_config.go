package config

import (
	"errors"
	"fmt"

	"github.com/crc-org/crc/pkg/crc/logging"
	"github.com/spf13/cast"
	"github.com/zalando/go-keyring"
)

const secretServiceName = "crc"

var ErrSecretsNotAccessible = errors.New("secret store is not accessible")

type Secret string

func (s Secret) String() string {
	return string(s)
}

type SecretStorage struct {
	secretService string
}

func NewSecretStorage() *SecretStorage {
	return &SecretStorage{
		secretService: secretServiceName,
	}
}

func (c *SecretStorage) Get(key string) interface{} {
	secret, err := keyring.Get(c.secretService, key)
	if err != nil || errors.Is(err, keyring.ErrNotFound) {
		logging.Debugf("error while getting config from secret store: %v", err)
		return nil
	}
	return secret
}

func (c *SecretStorage) Set(key string, value interface{}) error {
	secret, err := cast.ToStringE(value)
	if err != nil {
		return fmt.Errorf("Failed to cast secret value to string: %w", err)
	}
	return keyring.Set(c.secretService, key, secret)
}

func (c *SecretStorage) Unset(key string) error {
	err := keyring.Delete(c.secretService, key)
	if err != nil || errors.Is(err, keyring.ErrNotFound) {
		logging.Debugf("error while getting config from secret store: %v", err)
		return nil
	}
	return err
}

func NewEmptyInMemorySecretStorage() *SecretStorage {
	// keyring.MockInit()
	return &SecretStorage{
		secretService: secretServiceName,
	}
}
