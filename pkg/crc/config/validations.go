package config

import (
	"fmt"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/network"
	crcpreset "github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/crc/validation"
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

func ValidateString(value interface{}) (bool, string) {
	if _, err := cast.ToStringE(value); err != nil {
		return false, "must be a valid string"
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
func ValidateCPUs(value interface{}, preset crcpreset.Preset) (bool, string) {
	v, err := cast.ToIntE(value)
	if err != nil {
		return false, fmt.Sprintf("requires integer value >= %d", constants.GetDefaultCPUs(preset))
	}
	if err := validation.ValidateCPUs(v, preset); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// ValidateMemory checks if provided memory is valid in the config
func ValidateMemory(value interface{}, preset crcpreset.Preset) (bool, string) {
	v, err := cast.ToIntE(value)
	if err != nil {
		return false, fmt.Sprintf("requires integer value in MiB >= %d", constants.GetDefaultMemory(preset))
	}
	if err := validation.ValidateMemory(v, preset); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// ValidateBundlePath checks if the provided bundle path is valid
func ValidateBundlePath(value interface{}, preset crcpreset.Preset) (bool, string) {
	if err := validation.ValidateBundlePath(cast.ToString(value), preset); err != nil {
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

// ValidateHTTPProxy checks if given URI is valid for a HTTP proxy
func ValidateHTTPProxy(value interface{}) (bool, string) {
	if err := network.ValidateProxyURL(cast.ToString(value), false); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// ValidateHTTPSProxy checks if given URI is valid for a HTTPS proxy
func ValidateHTTPSProxy(value interface{}) (bool, string) {
	if err := network.ValidateProxyURL(cast.ToString(value), true); err != nil {
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

func validatePreset(value interface{}) (bool, string) {
	_, err := crcpreset.ParsePresetE(cast.ToString(value))
	if err != nil {
		return false, fmt.Sprintf("Unknown preset. Only %s and %s are valid.", crcpreset.Podman, crcpreset.OpenShift)
	}
	return true, ""
}
