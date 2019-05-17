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

var (
	globalViper *viper.Viper
	ViperConfig map[string]interface{}
)

// GetBool returns the value of a boolean config key
func GetBool(key string) bool {
	return globalViper.GetBool(key)
}

// Set sets the value for a give config key
func Set(key string, value interface{}) {
	globalViper.Set(key, value)
}

// Unset unsets a given config key
func Unset(key string) error {
	delete(ViperConfig, key)
	encodedConfig, err := json.MarshalIndent(ViperConfig, "", " ")
	if err != nil {
		return errors.NewF("Error encoding config to JSON: %v", err)
	}
	err = globalViper.ReadConfig(bytes.NewBuffer(encodedConfig))
	if err != nil {
		return errors.NewF("Error reading in new config: %s : %v", constants.ConfigFile, err)
	}
	return nil
}

// GetString return the value of a key in string
func GetString(key string) string {
	return globalViper.GetString(key)
}

// EnsureConfigFileExists creates the viper config file it does not exists
func EnsureConfigFileExists() error {
	_, err := os.Stat(constants.ConfigPath)
	if err != nil {
		f, err := os.Create(constants.ConfigPath)
		defer f.Close()
		f.WriteString("{}")
		return err
	}
	return err
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
	return v.Unmarshal(&ViperConfig)
}

// SetDefault sets the default for a config
func SetDefault(key string, value interface{}) {
	globalViper.SetDefault(key, value)
}

// WriteConfig write config to file
func WriteConfig() error {
	return globalViper.WriteConfig()
}

// AllConfigs returns all the configs
func AllConfigs() map[string]interface{} {
	return globalViper.AllSettings()
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
