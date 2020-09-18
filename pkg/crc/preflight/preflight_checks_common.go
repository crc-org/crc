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
		checkDescription: "Checking if oc binary is cached",
		check:            checkOcBinaryCached,
		fixDescription:   "Caching oc binary",
		fix:              fixOcBinaryCached,
	},
	{
		configKeySuffix:  "check-podman-cached",
		checkDescription: "Checking if podman remote binary is cached",
		check:            checkPodmanBinaryCached,
		fixDescription:   "Caching podman remote binary",
		fix:              fixPodmanBinaryCached,
	},
	{
		checkDescription: "Checking if goodhosts binary is cached",
		check:            checkGoodhostsBinaryCached,
		fixDescription:   "Caching goodhosts binary",
		fix:              fixGoodhostsBinaryCached,
	},
	{
		configKeySuffix:  "check-bundle-cached",
		checkDescription: "Checking if CRC bundle is cached in '$HOME/.crc'",
		check:            checkBundleCached,
		fixDescription:   "Unpacking bundle from the CRC binary",
		fix:              fixBundleCached,
		flags:            SetupOnly,
	},
	{
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
	if constants.BundleEmbedded() {
		bundleDir := filepath.Dir(constants.DefaultBundlePath)
		err := os.MkdirAll(bundleDir, 0700)
		if err != nil && !os.IsExist(err) {
			return fmt.Errorf("Cannot create directory %s", bundleDir)
		}

		return embed.Extract(filepath.Base(constants.DefaultBundlePath), constants.DefaultBundlePath)
	}
	return fmt.Errorf("CRC bundle is not embedded in the binary")
}

// Check if oc binary is cached or not
func checkOcBinaryCached() error {
	// Remove oc binary from older location and ignore the error
	// We should remove this code after 3-4 releases. (after 2020-07-10)
	os.Remove(filepath.Join(constants.CrcBinDir, "oc"))

	oc := cache.NewOcCache()
	if !oc.IsCached() {
		return errors.New("oc binary is not cached")
	}
	if err := oc.CheckVersion(); err != nil {
		return err
	}
	logging.Debug("oc binary already cached")
	return nil
}

func fixOcBinaryCached() error {
	oc := cache.NewOcCache()
	if err := oc.EnsureIsCached(); err != nil {
		return fmt.Errorf("Unable to download oc %v", err)
	}
	logging.Debug("oc binary cached")
	return nil
}

// Check if podman binary is cached or not
func checkPodmanBinaryCached() error {
	// Disable the podman cache until further notice
	logging.Debug("Currently podman remote is not supported")
	return nil
}

func fixPodmanBinaryCached() error {
	podman := cache.NewPodmanCache()
	if err := podman.EnsureIsCached(); err != nil {
		return fmt.Errorf("Unable to download podman remote binary %v", err)
	}
	logging.Debug("podman remote binary cached")
	return nil
}

// Check if goodhost binary is cached or not
func checkGoodhostsBinaryCached() error {
	goodhost := cache.NewGoodhostsCache()
	if !goodhost.IsCached() {
		return errors.New("goodhost binary is not cached")
	}
	logging.Debug("goodhost binary already cached")
	return checkSuid(goodhost.GetBinaryPath())
}

func fixGoodhostsBinaryCached() error {
	goodhost := cache.NewGoodhostsCache()
	if err := goodhost.EnsureIsCached(); err != nil {
		return fmt.Errorf("Unable to download goodhost binary %v", err)
	}
	logging.Debug("goodhost binary cached")
	return setSuid(goodhost.GetBinaryPath())
}
