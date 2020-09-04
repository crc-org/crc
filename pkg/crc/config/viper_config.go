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

type SetFn func(string, interface{}) string

type Setting struct {
	Name         string
	defaultValue interface{}
	validationFn ValidationFnType
	callbackFn   SetFn
}

const (
	configPropDoesntExistMsg = "Configuration property '%s' does not exist"
)

var (
	globalViper *viper.Viper
	// changedConfigs holds the config keys/values which have a non
	// default value (either because they are set in the config file, or
	// because they were changed at runtime)
	changedConfigs map[string]interface{}
	// allSettings holds all the config settings
	allSettings = make(map[string]*Setting)
)

func syncViperState(viper *viper.Viper) error {
	encodedConfig, err := json.MarshalIndent(changedConfigs, "", " ")
	if err != nil {
		return fmt.Errorf("Error encoding configuration to JSON: %v", err)
	}
	err = viper.ReadConfig(bytes.NewBuffer(encodedConfig))
	if err != nil {
		return fmt.Errorf("Error reading configuration file '%s': %v", constants.ConfigFile, err)
	}
	return nil
}

type SettingValue struct {
	Value     interface{}
	Invalid   bool
	IsDefault bool
}

func (v SettingValue) AsBool() bool {
	return cast.ToBool(v.Value)
}

func (v SettingValue) AsString() string {
	return cast.ToString(v.Value)
}

func (v SettingValue) AsInt() int {
	return cast.ToInt(v.Value)
}

// ensureConfigFileExists creates the viper config file if it does not exists
func ensureConfigFileExists() error {
	_, err := os.Stat(constants.ConfigPath)
	if err != nil {
		f, err := os.Create(constants.ConfigPath)
		if err == nil {
			_, err = f.WriteString("{}")
			f.Close()
		}
		return err
	}
	return nil
}

// InitViper initializes viper
func InitViper() error {
	if err := ensureConfigFileExists(); err != nil {
		return err
	}
	v := viper.New()
	v.SetConfigFile(constants.ConfigPath)
	v.SetConfigType("json")
	v.SetEnvPrefix(constants.CrcEnvPrefix)
	// Replaces '-' in flags with '_' in env variables
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()
	v.SetTypeByDefaultValue(true)
	err := v.ReadInConfig()
	if err != nil {
		return fmt.Errorf("Error reading configuration file '%s': %v", constants.ConfigFile, err)
	}
	v.WatchConfig()
	globalViper = v
	return v.Unmarshal(&changedConfigs)
}

// WriteConfig write config to file
func WriteConfig() error {
	// We recreate a new viper instance, as globalViper.WriteConfig()
	// writes both default values and set values back to disk while we only
	// want the latter to be written
	v := viper.New()
	v.SetConfigFile(constants.ConfigPath)
	v.SetConfigType("json")
	err := syncViperState(v)
	if err != nil {
		return err
	}
	return v.WriteConfig()
}

// AllConfigs returns all the known configs
// A known config is one which was registered through AddSetting
// - config with a default value
// - config with a value set
// - config with no value set
func AllConfigs() map[string]SettingValue {
	var allConfigs = make(map[string]SettingValue)
	for key := range allSettings {
		allConfigs[key] = Get(key)
	}
	return allConfigs
}

// BindFlagset binds a flagset to their respective config properties
func BindFlagSet(flagSet *pflag.FlagSet) error {
	return globalViper.BindPFlags(flagSet)
}

// AddSetting returns a filled struct of ConfigSetting
// takes the config name and default value as arguments
func AddSetting(name string, defValue interface{}, validationFn ValidationFnType, callbackFn SetFn) *Setting {
	s := Setting{Name: name, defaultValue: defValue, validationFn: validationFn, callbackFn: callbackFn}
	allSettings[name] = &s
	globalViper.SetDefault(name, defValue)
	return &s
}

// Set sets the value for a give config key
func Set(key string, value interface{}) (string, error) {
	_, ok := allSettings[key]
	if !ok {
		return "", fmt.Errorf(configPropDoesntExistMsg, key)
	}

	ok, expectedValue := allSettings[key].validationFn(value)
	if !ok {
		return "", fmt.Errorf("Value '%s' for configuration property '%s' is invalid, reason: %s", value, key, expectedValue)
	}

	globalViper.Set(key, value)
	changedConfigs[key] = value

	callbackMsg := allSettings[key].callbackFn(key, value)
	if callbackMsg != "" {
		return callbackMsg, nil
	}

	return "", nil
}

// Unset unsets a given config key
func Unset(key string) (string, error) {
	_, ok := allSettings[key]
	if !ok {
		return "", fmt.Errorf(configPropDoesntExistMsg, key)
	}

	delete(changedConfigs, key)
	if err := syncViperState(globalViper); err != nil {
		return "", fmt.Errorf("Error unsetting configuration property '%s': %v", key, err)
	}

	return fmt.Sprintf("Successfully unset configuration property '%s'", key), nil
}

func Get(key string) SettingValue {
	setting, ok := allSettings[key]
	if !ok || globalViper == nil {
		return SettingValue{
			Invalid: true,
		}
	}
	value := globalViper.Get(key)
	if value == nil {
		value = setting.defaultValue
	}
	return SettingValue{
		Value:     value,
		IsDefault: reflect.DeepEqual(setting.defaultValue, value),
	}
}
