package config

import (
	"fmt"
	"reflect"

	"github.com/crc-org/crc/pkg/crc/preset"
	"github.com/crc-org/crc/pkg/crc/validation"
	"github.com/spf13/cast"
)

const (
	configPropDoesntExistMsg = "Configuration property '%s' does not exist"
	invalidProp              = "Value '%v' for configuration property '%s' is invalid, reason: %s"
	invalidType              = "Type %T for configuration property '%s' is invalid"
)

type Config struct {
	storage        RawStorage
	secretStorage  RawStorage
	settingsByName map[string]Setting
}

func New(storage, secretStorage RawStorage) *Config {
	return &Config{
		storage:        storage,
		secretStorage:  secretStorage,
		settingsByName: make(map[string]Setting),
	}
}

// AllConfigs returns all the known configs
// A known config is one which was registered through AddSetting
// - config with a default value
// - config with a value set
// - config with no value set
func (c *Config) AllConfigs() map[string]SettingValue {
	var allConfigs = make(map[string]SettingValue)
	for key := range c.settingsByName {
		allConfigs[key] = c.Get(key)
	}
	return allConfigs
}

func (c *Config) AllSettings() []Setting {
	var settings []Setting
	for _, setting := range c.settingsByName {
		settings = append(settings, setting)
	}
	return settings
}

// AddSetting returns a filled struct of ConfigSetting
// takes the config name and default value as arguments
func (c *Config) AddSetting(name string, defValue interface{}, validationFn ValidationFnType, callbackFn SetFn, help string) {
	c.settingsByName[name] = Setting{
		Name:         name,
		defaultValue: defValue,
		validationFn: validationFn,
		callbackFn:   callbackFn,
		isSecret:     isUnderlyingTypeSecret(defValue),
		Help:         help,
	}
}

// Set sets the value for a given config key
func (c *Config) Set(key string, value interface{}) (string, error) {
	setting, ok := c.settingsByName[key]
	if !ok {
		return "", fmt.Errorf(configPropDoesntExistMsg, key)
	}

	ok, expectedValue := c.settingsByName[key].validationFn(value)
	if !ok {
		return "", fmt.Errorf(invalidProp, value, key, expectedValue)
	}

	var castValue interface{}
	var err error
	switch setting.defaultValue.(type) {
	case int:
		castValue, err = cast.ToIntE(value)
		if err != nil {
			return "", fmt.Errorf(invalidProp, value, key, err)
		}
	case string, Secret:
		castValue = cast.ToString(value)
	case bool:
		castValue, err = cast.ToBoolE(value)
		if err != nil {
			return "", fmt.Errorf(invalidProp, value, key, err)
		}
	case preset.Preset:
		castValue = cast.ToString(value)
	default:
		return "", fmt.Errorf(invalidType, value, key)
	}

	// Preset is mapped with `memory`, `cpus` and `bundle` and
	// we want to make sure if cpu or memory is less for a preset
	// then default is set automatic.
	if setting.Name == Preset {
		preset := preset.ParsePreset(value.(string))
		mem := c.Get(Memory)
		if err := validation.ValidateMemory(mem.AsInt(), preset); err != nil {
			if _, err := c.Unset(Memory); err != nil {
				return "", err
			}
		}
		cpu := c.Get(CPUs)
		if err := validation.ValidateCPUs(cpu.AsInt(), preset); err != nil {
			if _, err := c.Unset(CPUs); err != nil {
				return "", err
			}
		}
	}

	// Make sure if user try to set same value which
	// is default then just unset the value which
	// anyway make it default and don't update it
	// ~/.crc/crc.json (viper config) file.
	if setting.defaultValue == castValue {
		if _, err := c.Unset(key); err != nil {
			return "", err
		}
		return "", nil
	}

	if setting.isSecret {
		if err := c.secretStorage.Set(key, castValue); err != nil {
			return "", err
		}
	} else {
		if err := c.storage.Set(key, castValue); err != nil {
			return "", err
		}
	}

	return c.settingsByName[key].callbackFn(key, castValue), nil
}

// Unset unsets a given config key
func (c *Config) Unset(key string) (string, error) {
	setting, ok := c.settingsByName[key]
	if !ok {
		return "", fmt.Errorf(configPropDoesntExistMsg, key)
	}
	if setting.isSecret {
		if err := c.secretStorage.Unset(key); err != nil {
			return "", err
		}
	} else {
		if err := c.storage.Unset(key); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("Successfully unset configuration property '%s'", key), nil
}

func (c *Config) Get(key string) SettingValue {
	setting, ok := c.settingsByName[key]
	if !ok {
		return SettingValue{
			Invalid: true,
		}
	}
	var value interface{}
	if setting.isSecret {
		value = c.secretStorage.Get(key)
	} else {
		value = c.storage.Get(key)
	}
	if value == nil {
		value = setting.defaultValue
	}
	var err error
	switch setting.defaultValue.(type) {
	case int:
		value, err = cast.ToIntE(value)
		if err != nil {
			return SettingValue{
				Invalid: true,
			}
		}
	case string:
		value = cast.ToString(value)
	case bool:
		value, err = cast.ToBoolE(value)
		if err != nil {
			return SettingValue{
				Invalid: true,
			}
		}
	case preset.Preset:
		value, err = preset.ParsePresetE(cast.ToString(value))
		if err != nil {
			return SettingValue{
				Invalid: true,
			}
		}
	case Secret:
		value, err = toSecret(value)
		if err != nil {
			return SettingValue{
				Invalid: true,
			}
		}
	default:
		return SettingValue{
			Invalid: true,
		}
	}
	return SettingValue{
		Value:     value,
		IsDefault: reflect.DeepEqual(setting.defaultValue, value),
		IsSecret:  isUnderlyingTypeSecret(value),
	}
}

func toSecret(value interface{}) (Secret, error) {
	if isUnderlyingTypeSecret(value) {
		return value.(Secret), nil
	}
	v, err := cast.ToStringE(value)
	if err != nil {
		return Secret(""), err
	}
	return Secret(v), nil
}

func isUnderlyingTypeSecret(value interface{}) bool {
	if _, ok := value.(Secret); ok {
		return true
	}
	return false
}
