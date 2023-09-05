package config

import (
	"fmt"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/network/httpproxy"
	crcpreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/crc/validation"
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

func validateString(value interface{}) (bool, string) {
	if _, err := cast.ToStringE(value); err != nil {
		return false, "must be a valid string"
	}
	return true, ""
}

// validateDiskSize checks if provided disk size is valid in the config
func validateDiskSize(value interface{}) (bool, string) {
	diskSize, err := cast.ToIntE(value)
	if err != nil {
		return false, fmt.Sprintf("could not convert '%s' to integer", value)
	}
	if err := validation.ValidateDiskSize(diskSize); err != nil {
		return false, err.Error()
	}

	return true, ""
}

// validatePersistentVolumeSize checks if provided disk size is valid in the config
func validatePersistentVolumeSize(value interface{}) (bool, string) {
	diskSize, err := cast.ToIntE(value)
	if err != nil {
		return false, fmt.Sprintf("could not convert '%s' to integer", value)
	}
	if err := validation.ValidatePersistentVolumeSize(diskSize); err != nil {
		return false, err.Error()
	}

	return true, ""
}

// validateCPUs checks if provided cpus count is valid in the config
func validateCPUs(value interface{}, preset crcpreset.Preset) (bool, string) {
	v, err := cast.ToIntE(value)
	if err != nil {
		return false, fmt.Sprintf("requires integer value >= %d", constants.GetDefaultCPUs(preset))
	}
	if err := validation.ValidateCPUs(v, preset); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// validateMemory checks if provided memory is valid in the config
// It's defined as a variable so that it can be overridden in tests to disable the physical memory check
var validateMemory = func(value interface{}, preset crcpreset.Preset) (bool, string) {
	v, err := cast.ToIntE(value)
	if err != nil {
		return false, fmt.Sprintf("requires integer value in MiB >= %d", constants.GetDefaultMemory(preset))
	}
	if err := validation.ValidateMemory(v, preset); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// validateBundlePath checks if the provided bundle path is valid
func validateBundlePath(value interface{}, preset crcpreset.Preset) (bool, string) {
	if err := validation.ValidateBundlePath(cast.ToString(value), preset); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// validateIP checks if provided IP is valid
func validateIPAddress(value interface{}) (bool, string) {
	if err := validation.ValidateIPAddress(cast.ToString(value)); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// validatePath checks if provided path is exist
func validatePath(value interface{}) (bool, string) {
	if err := validation.ValidatePath(cast.ToString(value)); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// validateHTTPProxy checks if given URI is valid for a HTTP proxy
func validateHTTPProxy(value interface{}) (bool, string) {
	if err := httpproxy.ValidateProxyURL(cast.ToString(value), false); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// validateHTTPSProxy checks if given URI is valid for a HTTPS proxy
func validateHTTPSProxy(value interface{}) (bool, string) {
	if err := httpproxy.ValidateProxyURL(cast.ToString(value), true); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// validateNoProxy checks if the NoProxy string has the correct format
func validateNoProxy(value interface{}) (bool, string) {
	if strings.Contains(cast.ToString(value), " ") {
		return false, "NoProxy string can't contain spaces"
	}
	return true, ""
}

func validateYesNo(value interface{}) (bool, string) {
	if cast.ToString(value) == "yes" || cast.ToString(value) == "no" {
		return true, ""
	}
	return false, "must be yes or no"
}

func validatePreset(value interface{}) (bool, string) {
	_, err := crcpreset.ParsePresetE(cast.ToString(value))
	if err != nil {
		return false, fmt.Sprintf("Unknown preset. Only %s are valid.", crcpreset.AllPresets())
	}
	return true, ""
}

func validatePort(value interface{}) (bool, string) {
	port, err := cast.ToUintE(value)
	if err != nil {
		return false, "Requires integer value in range of 1024-65535"
	}
	if port < 1024 || port > 65535 {
		return false, fmt.Sprintf("Provided %d but requires value in range of 1024-65535", port)
	}
	return true, ""
}
