package config

import "github.com/spf13/cast"

type Setting struct {
	Name         string
	defaultValue interface{}
	validationFn ValidationFnType
	callbackFn   SetFn
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

// validationFnType takes the key, value as args and checks if valid
type ValidationFnType func(interface{}) (bool, string)
type SetFn func(string, interface{}) string

// RawStorage stores any key-value pair without validation
type RawStorage interface {
	Get(key string) interface{}
	Set(key string, value interface{}) error
	Unset(key string) error
}
