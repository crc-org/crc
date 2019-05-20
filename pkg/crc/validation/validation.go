package validation

import (
	"fmt"
	"github.com/code-ready/crc/pkg/crc/constants"
	"os"
)

// Validate the given driver is supported or not
func ValidateDriver(driver string) error {
	for _, d := range constants.SupportedVMDrivers {
		if driver == d {
			return nil
		}
	}
	return fmt.Errorf("Unsupported driver: %s, use '--vm-driver' option to provide a supported driver %s\n", driver, constants.SupportedVMDrivers)
}

// ValidateCPUs is check if provided cpus count is valid
func ValidateCPUs(value int) error {
	if value < constants.DefaultCPUs {
		return fmt.Errorf("CPUs required >=%d", constants.DefaultCPUs)
	}
	return nil
}

// ValidateMemory is check if provided Memory count is valid
func ValidateMemory(value int) error {
	if value < constants.DefaultMemory {
		return fmt.Errorf("Memory required >=%d", constants.DefaultMemory)
	}
	return nil
}

// ValidateBundle is check if provided bundle path exist
func ValidateBundle(bundle string) error {
	if _, err := os.Stat(bundle); os.IsNotExist(err) {
		return fmt.Errorf("Provided file %s does not exist", bundle)
	}
	return nil
}
