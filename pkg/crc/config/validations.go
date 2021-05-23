package config

import (
	"fmt"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/validation"
	"github.com/docker/go-units"
	"github.com/spf13/cast"
)

// ValidateBool is a fail safe in the case user
// makes a typo for boolean config values
func ValidateBool(value interface{}) (bool, string) {
	if _, err := cast.ToBoolE(value); err != nil {
		return false, "must be true or false"
	}

	return true, ""
}

// ValidateDiskSize checks if provided disk size is valid in the config
func ValidateDiskSize(value interface{}) (bool, string) {
	diskSize, err := cast.ToIntE(value)
	if err != nil {
		return false, fmt.Sprintf("could not convert '%s' to integer", value)
	}
	if err := validation.ValidateDiskSize(diskSize); err != nil {
		return false, err.Error()
	}

	return true, ""
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
	s, err := cast.ToStringE(value)
	if err != nil {
		return false, "value is not a string type, odd?"
	}

	v, err := units.FromHumanSize(s)
	if err != nil {
		return false, fmt.Sprintf("requires integer value in MiB >= %d", constants.DefaultMemory)
	}
	res := int(v / 1024 / 1024) // result is in bytes

	if err := validation.ValidateMemory(res); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// ValidateBundlePath checks if the provided bundle path is valid
func ValidateBundlePath(value interface{}) (bool, string) {
	if err := validation.ValidateBundlePath(cast.ToString(value)); err != nil {
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

func ValidateYesNo(value interface{}) (bool, string) {
	if cast.ToString(value) == "yes" || cast.ToString(value) == "no" {
		return true, ""
	}
	return false, "must be yes or no"
}
