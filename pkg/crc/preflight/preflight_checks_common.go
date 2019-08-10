package preflight

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/YourFin/binappend"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
)

var genericPreflightChecks = [...]PreflightCheck{
	{
		checkDescription: "Checking if oc binary is cached",
		check:            checkOcBinaryCached,
		fixDescription:   "Caching oc binary",
		fix:              fixOcBinaryCached,
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

func checkBundleCached() (bool, error) {
	if !constants.BundleEmbedded() {
		return true, nil
	}
	if _, err := os.Stat(constants.DefaultBundlePath); os.IsNotExist(err) {
		return false, err
	}
	return true, nil
}

func fixBundleCached() (bool, error) {
	if constants.BundleEmbedded() {
		currentExecutable, err := os.Executable()
		if err != nil {
			return false, err
		}
		extractor, err := binappend.MakeExtractor(currentExecutable)
		if err != nil {
			return false, err
		}
		available := extractor.AvalibleData()
		if len(available) < 1 {
			return false, fmt.Errorf("Invalid bundle data")
		}
		reader, err := extractor.GetReader(available[0])
		if err != nil {
			return false, err
		}
		defer reader.Close()

		bundleDir := filepath.Dir(constants.DefaultBundlePath)
		err = os.MkdirAll(bundleDir, 0700)
		if err != nil && !os.IsExist(err) {
			return false, fmt.Errorf("Cannot create directory %s", bundleDir)
		}
		writer, err := os.Create(constants.DefaultBundlePath)
		if err != nil {
			return false, err
		}
		defer writer.Close()
		_, err = io.Copy(writer, reader)
		if err != nil {
			return false, err
		}

		return true, nil
	}
	return false, fmt.Errorf("CRC bundle is not embedded in the binary")
}

// Check if oc binary is cached or not
func checkOcBinaryCached() (bool, error) {
	oc := oc.OcCached{}
	if !oc.IsCached() {
		return false, errors.New("oc binary is not cached")
	}
	logging.Debug("oc binary already cached")
	return true, nil
}

func fixOcBinaryCached() (bool, error) {
	oc := oc.OcCached{}
	if err := oc.EnsureIsCached(); err != nil {
		return false, fmt.Errorf("Unable to download oc %v", err)
	}
	logging.Debug("oc binary cached")
	return true, nil
}
