package config

import (
	"fmt"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/validation"
	"github.com/spf13/cast"
)

// ValidateBool is a fail safe in the case user
// makes a typo for boolean config values
func ValidateBool(value interface{}) (bool, string) {
	if cast.ToString(value) == "true" || cast.ToString(value) == "false" {
		return true, ""
	}
	return false, "must be true or false"
}

// ValidateCPUs checks if provided cpus count is valid in the config
func ValidateCPUs(value interface{}) (bool, string) {
	v, err := cast.ToIntE(value)
	if err != nil {
		return false, fmt.Sprintf("requires integer value >= %d", constants.DefaultCPUs)
	}
	if err := validation.ValidateCPUs(v); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// ValidateMemory checks if provided memory is valid in the config
func ValidateMemory(value interface{}) (bool, string) {
	v, err := cast.ToIntE(value)
	if err != nil {
		return false, fmt.Sprintf("requires integer value in MiB >= %d", constants.DefaultMemory)
	}
	if err := validation.ValidateMemory(v); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// ValidateBundle checks if provided bundle path is valid
func ValidateBundle(value interface{}) (bool, string) {
	if err := validation.ValidateBundle(cast.ToString(value)); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// ValidateIP checks if provided IP is valid
func ValidateIPAddress(value interface{}) (bool, string) {
	if err := validation.ValidateIPAddress(cast.ToString(value)); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// ValidatePath checks if provided path is exist
func ValidatePath(value interface{}) (bool, string) {
	if err := validation.ValidatePath(cast.ToString(value)); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// ValidateURI checks if given URI is valid
func ValidateURI(value interface{}) (bool, string) {
	if err := network.ValidateProxyURL(cast.ToString(value)); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// ValidateNoProxy checks if the NoProxy string has the correct format
func ValidateNoProxy(value interface{}) (bool, string) {
	if strings.Contains(cast.ToString(value), " ") {
		return false, "NoProxy string can't contain spaces"
	}
	return true, ""
}
