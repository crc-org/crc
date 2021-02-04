package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type ViperStorage struct {
	viper      *viper.Viper
	configFile string
}

func NewViperStorage(configFile, envPrefix string) (*ViperStorage, error) {
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
		return nil, fmt.Errorf("error reading configuration file '%s': %v", configFile, err)
	}
	return &ViperStorage{
		viper:      v,
		configFile: configFile,
	}, nil
}

func (c *ViperStorage) Get(key string) interface{} {
	return c.viper.Get(key)
}

func (c *ViperStorage) Set(key string, value interface{}) error {
	in, err := ioutil.ReadFile(c.configFile)
	if err != nil {
		return err
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(in, &cfg); err != nil {
		return err
	}
	cfg[key] = value
	bin, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := atomicWrite(bin, c.configFile); err != nil {
		return err
	}
	return c.viper.ReadInConfig()
}

func (c *ViperStorage) Unset(key string) error {
	in, err := ioutil.ReadFile(c.configFile)
	if err != nil {
		return err
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(in, &cfg); err != nil {
		return err
	}
	delete(cfg, key)
	bin, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := atomicWrite(bin, c.configFile); err != nil {
		return err
	}
	return c.viper.ReadInConfig()
}

// BindFlagset binds a flagset to their respective config properties
func (c *ViperStorage) BindFlagSet(flagSet *pflag.FlagSet) error {
	return c.viper.BindPFlags(flagSet)
}

// ensureConfigFileExists creates the viper config file if it does not exists
func ensureConfigFileExists(file string) error {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return ioutil.WriteFile(file, []byte("{}\n"), 0600)
	}
	return err
}

func atomicWrite(bin []byte, configFile string) error {
	ext := filepath.Ext(configFile)
	pattern := fmt.Sprintf("%s*%s", strings.TrimSuffix(filepath.Base(configFile), ext), ext)
	tmpFile, err := ioutil.TempFile(filepath.Dir(configFile), pattern)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()
	if err := tmpFile.Close(); err != nil {
		return err
	}
	if err := ioutil.WriteFile(tmpFile.Name(), bin, 0600); err != nil {
		return err
	}
	return os.Rename(tmpFile.Name(), configFile)
}
