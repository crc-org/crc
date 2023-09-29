package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/image"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine/bundle"
	crcpreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/docker/go-units"
	"github.com/pbnjay/memory"
)

// ValidateCPUs checks if provided cpus count is valid
func ValidateCPUs(value int, preset crcpreset.Preset) error {
	if value < constants.GetDefaultCPUs(preset) {
		return fmt.Errorf("requires CPUs >= %d", constants.GetDefaultCPUs(preset))
	}
	return nil
}

// ValidateMemory checks if provided Memory count is valid
func ValidateMemory(value int, preset crcpreset.Preset) error {
	if value < constants.GetDefaultMemory(preset) {
		return fmt.Errorf("requires memory in MiB >= %d", constants.GetDefaultMemory(preset))
	}
	return ValidateEnoughMemory(value)
}

func ValidateDiskSize(value int) error {
	if value < constants.DefaultDiskSize {
		return fmt.Errorf("requires disk size in GiB >= %d", constants.DefaultDiskSize)
	}

	return nil
}

func ValidatePersistentVolumeSize(value int) error {
	if value < constants.DefaultPersistentVolumeSize {
		return fmt.Errorf("requires disk size in GiB >= %d", constants.DefaultPersistentVolumeSize)
	}

	return nil
}

// ValidateEnoughMemory checks if enough memory is installed on the host
func ValidateEnoughMemory(value int) error {
	totalMemory := memory.TotalMemory()
	logging.Debugf("Total memory of system is %d bytes", totalMemory)
	valueBytes := value * 1024 * 1024
	if totalMemory < uint64(valueBytes) {
		return fmt.Errorf("only %s of memory found (%s required)",
			units.HumanSize(float64(totalMemory)),
			units.HumanSize(float64(valueBytes)))
	}
	return nil
}

// ValidateBundlePath checks if the provided bundle path exist
func ValidateBundlePath(bundlePath string, preset crcpreset.Preset) error {
	logging.Debugf("Got bundle path: %s", bundlePath)
	if err := ValidateURL(bundlePath); err != nil {
		var urlError *url.Error
		var invalidPathError *invalidPath
		// If error occur due to invalid path, then return with a more meaningful error message
		if errors.As(err, &invalidPathError) {
			return fmt.Errorf("%s is invalid or missing, run 'crc setup' to download the default bundle", bundlePath)
		}
		// Some local paths (for example relative paths) can't be parsed/validated by `ValidateURL` and will be validated here
		if errors.As(err, &urlError) {
			if err1 := ValidatePath(bundlePath); err1 != nil {
				return fmt.Errorf("%s is invalid or missing, run 'crc setup' to download the default bundle", bundlePath)
			}
		} else {
			return err
		}
	}

	userProvidedBundle := bundle.GetBundleNameFromURI(bundlePath)
	bundleMismatchWarning(userProvidedBundle, preset)
	return nil
}

func ValidateBundle(bundlePath string, preset crcpreset.Preset) error {
	bundleName := bundle.GetBundleNameFromURI(bundlePath)
	bundleMetadata, err := bundle.Get(bundleName)
	if err != nil {
		if bundlePath == constants.GetDefaultBundlePath(preset) {
			return nil
		}
		return ValidateBundlePath(bundlePath, preset)
	}
	bundleMismatchWarning(bundleMetadata.GetBundleName(), preset)
	/* 'bundle' is already unpacked in ~/.crc/cache */
	return nil
}

func bundleMismatchWarning(userProvidedBundle string, preset crcpreset.Preset) {
	userProvidedBundle = bundle.GetBundleNameWithExtension(userProvidedBundle)
	if userProvidedBundle != constants.GetDefaultBundle(preset) {
		// Should append underscore (_) here, as we don't want crc_libvirt_4.7.15.crcbundle
		// to be detected as a custom bundle for crc_libvirt_4.7.1.crcbundle
		usingCustomBundle := strings.HasPrefix(bundle.GetBundleNameWithoutExtension(userProvidedBundle),
			fmt.Sprintf("%s_", bundle.GetBundleNameWithoutExtension(constants.GetDefaultBundle(preset))))
		if usingCustomBundle {
			logging.Warnf("Using custom bundle %s", userProvidedBundle)
		} else {
			logging.Warnf("Using %s bundle, but %s is expected for this release", userProvidedBundle, constants.GetDefaultBundle(preset))
		}
	}
}

// ValidateIPAddress checks if provided IP is valid
func ValidateIPAddress(ipAddress string) error {
	ip := net.ParseIP(ipAddress).To4()
	if ip == nil {
		return fmt.Errorf("'%s' is not a valid IPv4 address", ipAddress)
	}
	return nil
}

func ValidateURL(uri string) error {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		logging.Debugf("Failed to parse url: %v", err)
		return err
	}
	switch {
	// If uri string is without scheme then check if it is a valid absolute path.
	// Relative paths will cause `ParseRequestURI` to error out, and will be handled in `ValidateBundlePath`
	case !u.IsAbs():
		return ValidatePath(uri)
	// In case of windows where path started with C:\<path> the uri scheme would be `C`
	// We are going to check if the uri scheme is a single letter and assume that it is windows drive path url
	// and send it to ValidatePath
	case len(u.Scheme) == 1:
		return ValidatePath(uri)
	case u.Scheme == "http", u.Scheme == "https":
		return nil
	case u.Scheme == "docker":
		return image.ValidateURI(u)
	default:
		return fmt.Errorf("invalid %s format (only supported http, https or docker)", uri)
	}
}

type invalidPath struct {
	path string
}

func (e *invalidPath) Error() string {
	return fmt.Sprintf("file '%s' does not exist", e.path)

}

// ValidatePath check if provide path is exist
func ValidatePath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &invalidPath{path: path}
	}
	return nil
}

type imagePullSecret struct {
	Auths map[string]map[string]interface{} `json:"auths"`
}

// ImagePullSecret checks if the given string is a valid image pull secret and returns an error if not.
func ImagePullSecret(secret string) error {
	if secret == "" {
		return errors.New("empty pull secret")
	}
	var s imagePullSecret
	err := json.Unmarshal([]byte(secret), &s)
	if err != nil {
		return fmt.Errorf("invalid pull secret: %v", err)
	}
	if len(s.Auths) == 0 {
		return fmt.Errorf("invalid pull secret: missing 'auths' JSON-object field")
	}

	for d, a := range s.Auths {
		_, authPresent := a["auth"]
		_, credsStorePresent := a["credsStore"]
		if !authPresent && !credsStorePresent {
			return fmt.Errorf("invalid pull secret, '%q' JSON-object requires either 'auth' or 'credsStore' field", d)
		}
	}
	return nil
}
