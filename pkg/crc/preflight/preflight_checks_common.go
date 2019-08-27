package preflight

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/YourFin/binappend"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
)

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

		bundleDir := filepath.Base(constants.DefaultBundlePath)
		err = os.MkdirAll(bundleDir, 0600)
		if err != nil && !os.IsExist(err) {
			return false, fmt.Errorf("Cannot create directory %s", bundleDir)
		}
		f, err := os.Create(constants.DefaultBundlePath)
		if err != nil {
			return false, err
		}
		available := extractor.AvalibleData()
		data, err := extractor.ByteArray(available[0])
		if err != nil {
			return false, err
		}
		w := bufio.NewWriter(f)
		defer w.Flush()
		_, err = w.Write(data)
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
