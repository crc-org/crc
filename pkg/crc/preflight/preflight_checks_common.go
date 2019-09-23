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
		currentExecutable, err := os.Executable()
		if err != nil {
			return err
		}
		extractor, err := binappend.MakeExtractor(currentExecutable)
		if err != nil {
			return err
		}
		available := extractor.AvalibleData()
		if len(available) < 1 {
			return fmt.Errorf("Invalid bundle data")
		}
		reader, err := extractor.GetReader(available[0])
		if err != nil {
			return err
		}
		defer reader.Close()

		bundleDir := filepath.Dir(constants.DefaultBundlePath)
		err = os.MkdirAll(bundleDir, 0700)
		if err != nil && !os.IsExist(err) {
			return fmt.Errorf("Cannot create directory %s", bundleDir)
		}
		writer, err := os.Create(constants.DefaultBundlePath)
		if err != nil {
			return err
		}
		defer writer.Close()
		_, err = io.Copy(writer, reader)
		if err != nil {
			return err
		}

		return nil
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
