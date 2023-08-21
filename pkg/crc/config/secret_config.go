package config

import (
	"errors"
	"fmt"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
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
	secretService   string
	storeAccessible bool
}

func NewSecretStorage() *SecretStorage {
	return &SecretStorage{
		secretService:   secretServiceName,
		storeAccessible: keyringAccessible(),
	}
}

func (c *SecretStorage) Get(key string) interface{} {
	if !c.storeAccessible {
		return nil
	}
	secret, err := keyring.Get(c.secretService, key)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return secret
}

func (c *SecretStorage) Set(key string, value interface{}) error {
	if !c.storeAccessible {
		return ErrSecretsNotAccessible
	}
	secret, err := cast.ToStringE(value)
	if err != nil {
		return fmt.Errorf("Failed to cast secret value to string: %w", err)
	}
	return keyring.Set(c.secretService, key, secret)
}

func (c *SecretStorage) Unset(key string) error {
	if !c.storeAccessible {
		return ErrSecretsNotAccessible
	}
	err := keyring.Delete(c.secretService, key)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

func keyringAccessible() bool {
	err := keyring.Set("crc-test", "foo", "bar")
	if err == nil {
		_ = keyring.Delete("crc-test", "foo")
		return true
	}
	logging.Debugf("Keyring is not accessible: %v", err)
	return false
}

func NewEmptyInMemorySecretStorage() *SecretStorage {
	keyring.MockInit()
	return &SecretStorage{
		secretService:   secretServiceName,
		storeAccessible: true,
	}
}
