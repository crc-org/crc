package preflight

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/cache"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/validation"
	"github.com/code-ready/crc/pkg/embed"
	"github.com/docker/go-units"
)

var genericPreflightChecks = [...]Check{
	{
		order:            100,
		configKeySuffix:  "check-oc-cached",
		checkDescription: "Checking if oc executable is cached",
		check:            checkOcExecutableCached,
		fixDescription:   "Caching oc executable",
		fix:              fixOcExecutableCached,
	},
	{
		order:            100,
		configKeySuffix:  "check-podman-cached",
		checkDescription: "Checking if podman remote executable is cached",
		check:            checkPodmanExecutableCached,
		fixDescription:   "Caching podman remote executable",
		fix:              fixPodmanExecutableCached,
	},
	{
		order:            100,
		configKeySuffix:  "check-goodhosts-cached",
		checkDescription: "Checking if goodhosts executable is cached",
		check:            checkGoodhostsExecutableCached,
		fixDescription:   "Caching goodhosts executable",
		fix:              fixGoodhostsExecutableCached,
	},
	{
		order:            100,
		configKeySuffix:  "check-bundle-cached",
		checkDescription: "Checking if CRC bundle is cached in '$HOME/.crc'",
		check:            checkBundleCached,
		fixDescription:   "Unpacking bundle from the CRC executable",
		fix:              fixBundleCached,
		flags:            SetupOnly,
	},
	{
		order:            10,
		configKeySuffix:  "check-ram",
		checkDescription: "Checking minimum RAM requirements",
		check: func() error {
			return validation.ValidateEnoughMemory(constants.DefaultMemory)
		},
		fixDescription: fmt.Sprintf("crc requires at least %s to run", units.HumanSize(float64(constants.DefaultMemory*1024*1024))),
		flags:          NoFix,
	},
}

func checkBundleCached() error {
	if !constants.BundleEmbedded() {
		return nil
	}
	if _, err := os.Stat(constants.DefaultBundlePath); os.IsNotExist(err) {
		return err
	}
	return nil
}

func fixBundleCached() error {
	// Should be removed after 1.19 release
	// This check will ensure correct mode for `~/.crc/cache` directory
	// in case it exists.
	if err := os.Chmod(constants.MachineCacheDir, 0775); err != nil {
		logging.Debugf("Error changing %s permissions to 0775", constants.MachineCacheDir)
	}
	if constants.BundleEmbedded() {
		bundleDir := filepath.Dir(constants.DefaultBundlePath)
		err := os.MkdirAll(bundleDir, 0775)
		if err != nil && !os.IsExist(err) {
			return fmt.Errorf("Cannot create directory %s", bundleDir)
		}

		return embed.Extract(filepath.Base(constants.DefaultBundlePath), constants.DefaultBundlePath)
	}
	return fmt.Errorf("CRC bundle is not embedded in the executable")
}

// Check if oc executable is cached or not
func checkOcExecutableCached() error {
	// Remove oc executable from older location and ignore the error
	// We should remove this code after 3-4 releases. (after 2020-07-10)
	os.Remove(filepath.Join(constants.CrcBinDir, "oc"))

	oc := cache.NewOcCache()
	if !oc.IsCached() {
		return errors.New("oc executable is not cached")
	}
	if err := oc.CheckVersion(); err != nil {
		return err
	}
	logging.Debug("oc executable already cached")
	return nil
}

func fixOcExecutableCached() error {
	oc := cache.NewOcCache()
	if err := oc.EnsureIsCached(); err != nil {
		return fmt.Errorf("Unable to download oc %v", err)
	}
	logging.Debug("oc executable cached")
	return nil
}

// Check if podman executable is cached or not
func checkPodmanExecutableCached() error {
	// Disable the podman cache until further notice
	logging.Debug("Currently podman remote is not supported")
	return nil
}

func fixPodmanExecutableCached() error {
	podman := cache.NewPodmanCache()
	if err := podman.EnsureIsCached(); err != nil {
		return fmt.Errorf("Unable to download podman remote executable %v", err)
	}
	logging.Debug("podman remote executable cached")
	return nil
}

// Check if goodhost executable is cached or not
func checkGoodhostsExecutableCached() error {
	goodhost := cache.NewGoodhostsCache()
	if !goodhost.IsCached() {
		return errors.New("goodhost executable is not cached")
	}
	logging.Debug("goodhost executable already cached")
	return checkSuid(goodhost.GetExecutablePath())
}

func fixGoodhostsExecutableCached() error {
	goodhost := cache.NewGoodhostsCache()
	if err := goodhost.EnsureIsCached(); err != nil {
		return fmt.Errorf("Unable to download goodhost executable %v", err)
	}
	logging.Debug("goodhost executable cached")
	return setSuid(goodhost.GetExecutablePath())
}
