package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type ViperStorage struct {
	storeLock *sync.Mutex

	// store flag values
	flagSet *pflag.FlagSet

	configFile string
	envPrefix  string
}

func NewViperStorage(configFile, envPrefix string) (*ViperStorage, error) {
	return &ViperStorage{
		storeLock:  &sync.Mutex{},
		configFile: configFile,
		envPrefix:  envPrefix,
	}, nil
}

func (c *ViperStorage) viperInstance() (*viper.Viper, error) {
	if err := ensureConfigFileExists(c.configFile); err != nil {
		return nil, err
	}
	v := viper.New()
	v.SetConfigFile(c.configFile)
	v.SetConfigType("json")
	v.SetEnvPrefix(c.envPrefix)
	// Replaces '-' in flags with '_' in env variables
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()
	v.SetTypeByDefaultValue(true)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading configuration file '%s': %v", c.configFile, err)
	}
	if c.flagSet == nil {
		return v, nil
	}
	return v, v.BindPFlags(c.flagSet)
}

func (c *ViperStorage) Get(key string) interface{} {
	c.storeLock.Lock()
	defer c.storeLock.Unlock()
	viperInstance, err := c.viperInstance()
	if err != nil {
		return nil
	}
	return viperInstance.Get(key)
}

func (c *ViperStorage) Set(key string, value interface{}) error {
	c.storeLock.Lock()
	defer c.storeLock.Unlock()
	if err := ensureConfigFileExists(c.configFile); err != nil {
		return err
	}
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
	return atomicWrite(bin, c.configFile)
}

func (c *ViperStorage) Unset(key string) error {
	c.storeLock.Lock()
	defer c.storeLock.Unlock()
	if err := ensureConfigFileExists(c.configFile); err != nil {
		return err
	}
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
	return atomicWrite(bin, c.configFile)
}

// BindFlagset binds a flagset to their respective config properties
func (c *ViperStorage) BindFlagSet(flagSet *pflag.FlagSet) error {
	c.storeLock.Lock()
	defer c.storeLock.Unlock()
	c.flagSet = flagSet
	return nil
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
