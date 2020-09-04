package config

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/spf13/cast"
	"github.com/spf13/pflag"
)

const (
	configPropDoesntExistMsg = "Configuration property '%s' does not exist"
)

var defaultConfig *Config

type Config struct {
	storage     RawStorage
	allSettings map[string]*Setting
}

func New(storage RawStorage) *Config {
	return &Config{
		storage:     storage,
		allSettings: make(map[string]*Setting),
	}
}

// InitViper initializes viper
func InitViper() error {
	storage, err := NewViperStorage(constants.ConfigPath, constants.CrcEnvPrefix)
	if err != nil {
		return err
	}
	defaultConfig = New(storage)
	return nil
}

// AllConfigs returns all the known configs
// A known config is one which was registered through AddSetting
// - config with a default value
// - config with a value set
// - config with no value set
func AllConfigs() map[string]SettingValue {
	return defaultConfig.AllConfigs()
}

func (c *Config) AllConfigs() map[string]SettingValue {
	var allConfigs = make(map[string]SettingValue)
	for key := range c.allSettings {
		allConfigs[key] = c.Get(key)
	}
	return allConfigs
}

// BindFlagset binds a flagset to their respective config properties
func BindFlagSet(flagSet *pflag.FlagSet) error {
	return defaultConfig.BindFlagSet(flagSet)
}

func (c *Config) BindFlagSet(flagSet *pflag.FlagSet) error {
	if v, ok := c.storage.(*ViperStorage); ok {
		return v.BindFlagSet(flagSet)
	}
	return errors.New("not implemented")
}

// AddSetting returns a filled struct of ConfigSetting
// takes the config name and default value as arguments
func AddSetting(name string, defValue interface{}, validationFn ValidationFnType, callbackFn SetFn) *Setting {
	return defaultConfig.AddSetting(name, defValue, validationFn, callbackFn)
}

func (c *Config) AddSetting(name string, defValue interface{}, validationFn ValidationFnType, callbackFn SetFn) *Setting {
	s := Setting{Name: name, defaultValue: defValue, validationFn: validationFn, callbackFn: callbackFn}
	c.allSettings[name] = &s
	return &s
}

// Set sets the value for a given config key
func Set(key string, value interface{}) (string, error) {
	return defaultConfig.Set(key, value)
}

func (c *Config) Set(key string, value interface{}) (string, error) {
	setting, ok := c.allSettings[key]
	if !ok {
		return "", fmt.Errorf(configPropDoesntExistMsg, key)
	}

	var castValue interface{}
	switch setting.defaultValue.(type) {
	case int:
		castValue = cast.ToInt(value)
	case string:
		castValue = cast.ToString(value)
	case bool:
		castValue = cast.ToBool(value)
	}

	ok, expectedValue := c.allSettings[key].validationFn(castValue)
	if !ok {
		return "", fmt.Errorf("Value '%s' for configuration property '%s' is invalid, reason: %s", castValue, key, expectedValue)
	}

	if err := c.storage.Set(key, castValue); err != nil {
		return "", err
	}

	return c.allSettings[key].callbackFn(key, castValue), nil
}

// Unset unsets a given config key
func Unset(key string) (string, error) {
	return defaultConfig.Unset(key)
}

func (c *Config) Unset(key string) (string, error) {
	_, ok := c.allSettings[key]
	if !ok {
		return "", fmt.Errorf(configPropDoesntExistMsg, key)
	}

	if err := c.storage.Unset(key); err != nil {
		return "", err
	}

	return fmt.Sprintf("Successfully unset configuration property '%s'", key), nil
}

func Get(key string) SettingValue {
	return defaultConfig.Get(key)
}

func (c *Config) Get(key string) SettingValue {
	setting, ok := c.allSettings[key]
	if !ok {
		return SettingValue{
			Invalid: true,
		}
	}
	value := c.storage.Get(key)
	if value == nil {
		value = setting.defaultValue
	}
	switch setting.defaultValue.(type) {
	case int:
		value = cast.ToInt(value)
	case string:
		value = cast.ToString(value)
	case bool:
		value = cast.ToBool(value)
	default:
		return SettingValue{
			Invalid: true,
		}
	}
	return SettingValue{
		Value:     value,
		IsDefault: reflect.DeepEqual(setting.defaultValue, value),
	}
}
