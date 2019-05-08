package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
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
	// TODO: add flags and bind to configs
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
