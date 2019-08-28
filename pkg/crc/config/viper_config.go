package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type setting struct {
	Name          string
	DefaultValue  interface{}
	ValidationFns []ValidationFnType
}

var (
	globalViper *viper.Viper
	// changedConfigs holds the config keys/values which have a non
	// default value (either because they are set in the config file, or
	// because they were changed at runtime)
	changedConfigs map[string]interface{}
	// allSettings holds all the config settings
	allSettings = make(map[string]*setting)
)

// GetBool returns the value of a boolean config key
func GetBool(key string) bool {
	return globalViper.GetBool(key)
}

func set(key string, value interface{}) {
	globalViper.Set(key, value)
	changedConfigs[key] = value
}

func syncViperState(viper *viper.Viper) error {
	encodedConfig, err := json.MarshalIndent(changedConfigs, "", " ")
	if err != nil {
		return errors.Newf("Error encoding config to JSON: %v", err)
	}
	err = viper.ReadConfig(bytes.NewBuffer(encodedConfig))
	if err != nil {
		return errors.Newf("Error reading in new config: %s : %v", constants.ConfigFile, err)
	}
	return nil
}

func unset(key string) error {
	delete(changedConfigs, key)
	return syncViperState(globalViper)
}

// GetString return the value of a key in string
func GetString(key string) string {
	return globalViper.GetString(key)
}

// GetInt return the value of a key in int
func GetInt(key string) int {
	return globalViper.GetInt(key)
}

// EnsureConfigFileExists creates the viper config file if it does not exists
func EnsureConfigFileExists() error {
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
		return fmt.Errorf("Error Reading config file: %s : %v", constants.ConfigFile, err)
	}
	globalViper = v
	return v.Unmarshal(&changedConfigs)
}

// setDefault sets the default for a config
func setDefault(key string, value interface{}) {
	globalViper.SetDefault(key, value)
}

func SetDefaults() {
	for _, setting := range allSettings {
		setDefault(setting.Name, setting.DefaultValue)
	}
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

// AllConfigs returns all the configs
// 'all the configs' means
// - config keys with a default value
// - config keys with a value set
// This does not include config keys with no default value, and no value set
func AllConfigs() map[string]interface{} {
	return globalViper.AllSettings()
}

// AllConfigKeys returns all the known config keys
// A known config key is one which was registered through AddSetting
// - config keys with a default value
// - config keys with a value set
// - config keys with no value set
// This is different from AllConfigs behaviour, and is there to maintain backwards compatibility
// while this is refactored
func AllConfigKeys() []string {
	var keys []string
	for key := range allSettings {
		keys = append(keys, key)
	}
	return keys
}

// IsSet returns true if the config property is set
func IsSet(key string) bool {
	ss := AllConfigs()
	_, ok := ss[key]
	return ok
}

// BindFlags binds flags to config properties
func BindFlag(key string, flag *pflag.Flag) error {
	return globalViper.BindPFlag(key, flag)
}

// BindFlagset binds a flagset to their repective config properties
func BindFlagSet(flagSet *pflag.FlagSet) error {
	return globalViper.BindPFlags(flagSet)
}

// CreateSetting returns a filled struct of ConfigSetting
// takes the config name and default value as arguments
func AddSetting(name string, defValue interface{}, validationFn []ValidationFnType) *setting {
	s := setting{Name: name, DefaultValue: defValue, ValidationFns: validationFn}
	allSettings[name] = &s
	return &s
}

func runValidations(validations []ValidationFnType, value interface{}) (bool, string) {
	for _, fn := range validations {
		ok, expectedValue := fn(value)
		if !ok {
			return false, expectedValue
		}
	}
	return true, ""
}

// Set sets the value for a give config key
func Set(key string, value interface{}) error {
	_, ok := allSettings[key]
	if !ok {
		return fmt.Errorf("Config property '%s' does not exist", key)
	}

	ok, expectedValue := runValidations(allSettings[key].ValidationFns, value)
	if !ok {
		return fmt.Errorf("Config value is invalid: %s, Expected: %s\n", value, expectedValue)
	}

	set(key, value)

	return nil
}

// Unset unsets a given config key
func Unset(key string) error {
	_, ok := allSettings[key]
	if !ok {
		return fmt.Errorf("Config property does not exist: %s", key)
	}

	if !IsSet(key) {
		return fmt.Errorf("Config property is not set: %s", key)
	}
	if err := unset(key); err != nil {
		return fmt.Errorf("Error unsetting config property: %s : %v", key, err)
	}

	return nil
}

func Get(key string) (interface{}, error) {
	v, ok := changedConfigs[key]
	if !ok {
		return nil, fmt.Errorf("Config property '%s' does not exist", key)
	}

	return v, nil
}
