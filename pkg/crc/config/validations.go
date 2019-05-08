package config

import (
	"fmt"
	"github.com/code-ready/crc/pkg/crc/constants"
)

// ValidateBool is a fail safe in the case user
// makes a typo for boolean config values
func ValidateBool(value interface{}) (bool, string) {
	if value.(string) == "true" || value.(string) == "false" {
		return true, ""
	}
	return false, "true/false"
}

// validateDriver is check if driver is valid in the config
func ValidateDriver(value interface{}) (bool, string) {
	if err := constants.ValidateDriver(value.(string)); err != nil {
		return false, fmt.Sprintf("%s", constants.SupportedVMDrivers)
	}
	return true, ""
}
