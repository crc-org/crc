package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/spf13/cast"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	configPropDoesntExistMsg = "Configuration property '%s' does not exist"
)

var defaultConfig *Config

type Config struct {
	globalViper *viper.Viper
	// allSettings holds all the config settings
	allSettings map[string]*Setting
	configFile  string
}

// ensureConfigFileExists creates the viper config file if it does not exists
func ensureConfigFileExists(configFile string) error {
	_, err := os.Stat(configFile)
	if err != nil {
		f, err := os.Create(configFile)
		if err == nil {
			_, err = f.WriteString("{}")
			f.Close()
		}
		return err
	}
	return nil
}

func New(configFile, envPrefix string) (*Config, error) {
	if err := ensureConfigFileExists(configFile); err != nil {
		return nil, err
	}
	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("json")
	v.SetEnvPrefix(envPrefix)
	// Replaces '-' in flags with '_' in env variables
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()
	v.SetTypeByDefaultValue(true)
	err := v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("Error reading configuration file '%s': %v", configFile, err)
	}
	v.WatchConfig()
	return &Config{
		globalViper: v,
		allSettings: make(map[string]*Setting),
		configFile:  configFile,
	}, nil
}

// InitViper initializes viper
func InitViper() error {
	var err error
	defaultConfig, err = New(constants.ConfigPath, constants.CrcEnvPrefix)
	return err
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
	return c.globalViper.BindPFlags(flagSet)
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

// Set sets the value for a give config key
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

	c.globalViper.Set(key, castValue)

	return c.allSettings[key].callbackFn(key, castValue), c.globalViper.WriteConfig()
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

	settings := c.globalViper.AllSettings()
	delete(settings, key)
	bin, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}
	if err = c.globalViper.ReadConfig(bytes.NewReader(bin)); err != nil {
		return "", err
	}

	return fmt.Sprintf("Successfully unset configuration property '%s'", key), c.globalViper.WriteConfig()
}

func Get(key string) SettingValue {
	return defaultConfig.Get(key)
}

func (c *Config) Get(key string) SettingValue {
	setting, ok := c.allSettings[key]
	if !ok || c.globalViper == nil {
		return SettingValue{
			Invalid: true,
		}
	}
	value := c.globalViper.Get(key)
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
