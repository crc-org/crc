package validation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/constants"
	crcErrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/containers/common/pkg/auth"
	"github.com/docker/go-units"
	"github.com/pbnjay/memory"
)

// ValidateCPUs checks if provided cpus count is valid
func ValidateCPUs(value int) error {
	if value < constants.DefaultCPUs {
		return fmt.Errorf("requires CPUs >= %d", constants.DefaultCPUs)
	}
	return nil
}

// ValidateMemory checks if provided Memory count is valid
func ValidateMemory(value int) error {
	if value < constants.DefaultMemory {
		return fmt.Errorf("requires memory in MiB >= %d", constants.DefaultMemory)
	}
	return ValidateEnoughMemory(value)
}

func ValidateDiskSize(value int) error {
	if value < constants.DefaultDiskSize {
		return fmt.Errorf("requires disk size in GiB >= %d", constants.DefaultDiskSize)
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
func ValidateBundlePath(bundlePath string) error {
	if err := ValidatePath(bundlePath); err != nil {
		if constants.BundleEmbedded() {
			return fmt.Errorf("Run 'crc setup' to unpack the bundle to disk")
		}
		return fmt.Errorf("%s not found, please provide the path to a valid bundle using the -b option", bundlePath)
	}

	userProvidedBundle := filepath.Base(bundlePath)
	if userProvidedBundle != constants.GetDefaultBundle() {
		// Should append underscore (_) here, as we don't want crc_libvirt_4.7.15.crcbundle
		// to be detected as a custom bundle for crc_libvirt_4.7.1.crcbundle
		usingCustomBundle := strings.HasPrefix(bundle.GetBundleNameWithoutExtension(userProvidedBundle),
			fmt.Sprintf("%s_", bundle.GetBundleNameWithoutExtension(constants.GetDefaultBundle())))
		if usingCustomBundle {
			logging.Warnf("Using custom bundle %s", userProvidedBundle)
			return nil
		}
		if !constants.IsRelease() {
			logging.Warnf("Using unsupported bundle %s", userProvidedBundle)
			return nil
		}
		return fmt.Errorf("%s is not supported by this crc executable, please use %s", userProvidedBundle, constants.GetDefaultBundle())
	}
	return nil
}

func ValidateBundle(bundlePath string) error {
	bundleName := filepath.Base(bundlePath)
	_, err := bundle.Get(bundleName)
	if err != nil {
		return ValidateBundlePath(bundlePath)
	}
	/* 'bundle' is already unpacked in ~/.crc/cache */
	return nil
}

// ValidateIPAddress checks if provided IP is valid
func ValidateIPAddress(ipAddress string) error {
	ip := net.ParseIP(ipAddress).To4()
	if ip == nil {
		return fmt.Errorf("'%s' is not a valid IPv4 address", ipAddress)
	}
	return nil
}

// ValidatePath check if provide path is exist
func ValidatePath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file '%s' does not exist", path)
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

	if version.IsOkdBuild() {
		return nil
	}

	for d, a := range s.Auths {
		_, authPresent := a["auth"]
		_, credsStorePresent := a["credsStore"]
		if !authPresent && !credsStorePresent {
			return fmt.Errorf("invalid pull secret, '%q' JSON-object requires either 'auth' or 'credsStore' field", d)
		}
	}

	return tryLoginToRegistries(secret)
}

func getRegistriesFromPullSecret(secret string) ([]string, error) {
	var pullSecret imagePullSecret
	var registries []string
	err := json.Unmarshal([]byte(secret), &pullSecret)
	if err != nil {
		return []string{}, err
	}
	for registry := range pullSecret.Auths {
		registries = append(registries, registry)
	}
	return registries, nil
}

func tryLoginToRegistries(secret string) error {
	var mErr = crcErrors.MultiError{}
	registries, err := getRegistriesFromPullSecret(secret)
	if err != nil {
		return err
	}

	pullSecretFile, cleanup, err := writePullSecretToTempFile(secret)
	if err != nil {
		return err
	}
	defer cleanup()

	for _, registry := range registries {
		// cloud.openshift.com is not a docker registry but included in the pull-secret
		if registry == "cloud.openshift.com" {
			continue
		}
		logging.Debugf("Trying to login to %s with pull-secret file", registry)
		err := authenticateWithPullSecret(pullSecretFile, registry)
		if err != nil {
			mErr.Collect(err)
			logging.Debugf("Failed to login to %s: %v", registry, err)
		}
	}
	if len(mErr.Errors) != 0 {
		return fmt.Errorf("unable to login to registries with provided pull secret: %s", mErr.Error())
	}
	return nil
}

func authenticateWithPullSecret(pullSecretPath, registry string) error {
	opts := auth.LoginOptions{
		AuthFile: pullSecretPath,
		Stdout:   ioutil.Discard,
		Stdin:    os.Stdin,
	}
	timeOutDuration := 6 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeOutDuration)
	defer cancel()

	return auth.Login(ctx, nil, &opts, []string{registry})
}

func writePullSecretToTempFile(secret string) (string, func(), error) {
	tempDir, err := ioutil.TempDir("", "crc-pull")
	if err != nil {
		return "", func() {}, err
	}
	pullSecretFile := filepath.Join(tempDir, "temp-crc-pull")
	if err := ioutil.WriteFile(pullSecretFile, []byte(secret), 0600); err != nil {
		os.RemoveAll(tempDir)
		return "", func() {}, err
	}
	return pullSecretFile, func() { os.RemoveAll(tempDir) }, nil
}
