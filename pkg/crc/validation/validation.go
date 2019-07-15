package validation

import (
	"fmt"
	"github.com/code-ready/crc/pkg/crc/version"
	"net"
	"os"
	"path/filepath"
	"strings"

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
	// Check if the version of the bundle provided by user is same as what is released with crc.
	releaseBundleVersion := version.GetBundleVersion()
	userProvidedBundleVersion := filepath.Base(bundle)
	if !strings.Contains(userProvidedBundleVersion, fmt.Sprintf("%s.crcbundle", releaseBundleVersion)) {
		return errors.NewF("%s bundle is not supported for this release use updated one (crc_<hypervisor>_%s.crcbundle)", userProvidedBundleVersion, releaseBundleVersion)
	}
	return nil
}

// ValidateIpAddress checks if provided IP is valid
func ValidateIpAddress(ipAddress string) error {
	ip := net.ParseIP(ipAddress).To4()
	if ip == nil {
		return errors.NewF("IPv4 address is not valid: '%s'", ipAddress)
	}
	return nil
}
