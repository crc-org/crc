package preflight

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
	"github.com/code-ready/crc/pkg/crc/podman"
	"github.com/code-ready/crc/pkg/embed"
	crcos "github.com/code-ready/crc/pkg/os"
)

var genericPreflightChecks = [...]PreflightCheck{
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
		configKeySuffix:  "check-bundle-cached",
		checkDescription: "Checking if CRC bundle is cached in '$HOME/.crc'",
		check:            checkBundleCached,
		fixDescription:   "Unpacking bundle from the CRC binary",
		fix:              fixBundleCached,
		flags:            SetupOnly,
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
	oc := oc.OcCached{}
	if !oc.IsCached() {
		return errors.New("oc binary is not cached")
	}
	logging.Debug("oc binary already cached")
	return nil
}

func fixOcBinaryCached() error {
	oc := oc.OcCached{}
	if err := oc.EnsureIsCached(); err != nil {
		return fmt.Errorf("Unable to download oc %v", err)
	}
	logging.Debug("oc binary cached")
	return nil
}

// Check if podman binary is cached or not
func checkPodmanBinaryCached() error {

	// See issue #961; Currently does not work on Windows in combination with the CRC vm.
	if crcos.CurrentOS() == crcos.WINDOWS {
		logging.Debug("podman remote currently not supported on this platform")
		return nil
	}

	podman := podman.PodmanCached{}
	if !podman.IsCached() {
		return errors.New("podman remote binary is not cached")
	}
	logging.Debug("podman remote binary already cached")
	return nil
}

func fixPodmanBinaryCached() error {
	podman := podman.PodmanCached{}
	if err := podman.EnsureIsCached(); err != nil {
		return fmt.Errorf("Unable to download podman remote binary %v", err)
	}
	logging.Debug("podman remote binary cached")
	return nil
}
