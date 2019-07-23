package validation

import (
	"encoding/json"
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
	return errors.Newf("Unsupported driver: %s, use '--vm-driver' option to provide a supported driver %s\n", driver, machine.SupportedDriverValues())
}

// ValidateCPUs checks if provided cpus count is valid
func ValidateCPUs(value int) error {
	if value < constants.DefaultCPUs {
		return errors.Newf("CPUs required >=%d", constants.DefaultCPUs)
	}
	return nil
}

// ValidateMemory checks if provided Memory count is valid
func ValidateMemory(value int) error {
	if value < constants.DefaultMemory {
		return errors.Newf("Memory required >=%d", constants.DefaultMemory)
	}
	return nil
}

// ValidateBundle checks if provided bundle path exist
func ValidateBundle(bundle string) error {
	if err := ValidatePath(bundle); err != nil {
		return err
	}
	// Check if the version of the bundle provided by user is same as what is released with crc.
	releaseBundleVersion := version.GetBundleVersion()
	userProvidedBundleVersion := filepath.Base(bundle)
	if !strings.Contains(userProvidedBundleVersion, fmt.Sprintf("%s.crcbundle", releaseBundleVersion)) {
		return errors.Newf("%s bundle is not supported for this release use updated one (crc_<hypervisor>_%s.crcbundle)", userProvidedBundleVersion, releaseBundleVersion)
	}
	return nil
}

// ValidateIpAddress checks if provided IP is valid
func ValidateIpAddress(ipAddress string) error {
	ip := net.ParseIP(ipAddress).To4()
	if ip == nil {
		return errors.Newf("IPv4 address is not valid: '%s'", ipAddress)
	}
	return nil
}

// ValidatePath check if provide path is exist
func ValidatePath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.Newf("File %s does not exist", path)
	}
	return nil
}

type imagePullSecret struct {
	Auths map[string]map[string]interface{} `json:"auths"`
}

// ImagePullSecret checks if the given string is a valid image pull secret and returns an error if not.
func ImagePullSecret(secret string) error {
	var s imagePullSecret
	err := json.Unmarshal([]byte(secret), &s)
	if err != nil {
		return fmt.Errorf("invalid pull secret: %v", err)
	}
	if len(s.Auths) == 0 {
		return fmt.Errorf("invalid pull secret, auths required")
	}

	for d, a := range s.Auths {
		_, authPresent := a["auth"]
		_, credsStorePresent := a["credsStore"]
		if !authPresent && !credsStorePresent {
			return fmt.Errorf("invalid pull secret, %q requires either auth or credsStore", d)
		}
	}
	return nil
}
