package validation

import (
	"os"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/machine"
)

// Validate the given driver is supported or not
func ValidateDriver(driver string) error {
	for _, d := range machine.SupportedDriverValues() {
		if driver == d {
			return nil
		}
	}
	return errors.NewF("Unsupported driver: %s, use '--vm-driver' option to provide a supported driver %s\n", driver, machine.SupportedDriverValues())
}

// ValidateCPUs checks if provided cpus count is valid
func ValidateCPUs(value int) error {
	if value < constants.DefaultCPUs {
		return errors.NewF("CPUs required >=%d", constants.DefaultCPUs)
	}
	return nil
}

// ValidateMemory checks if provided Memory count is valid
func ValidateMemory(value int) error {
	if value < constants.DefaultMemory {
		return errors.NewF("Memory required >=%d", constants.DefaultMemory)
	}
	return nil
}

// ValidateBundle checks if provided bundle path exist
func ValidateBundle(bundle string) error {
	if _, err := os.Stat(bundle); os.IsNotExist(err) {
		return errors.NewF("Expected file %s does not exist", bundle)
	}
	return nil
}
