package preflight

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/YourFin/binappend"
	"github.com/kardianos/osext"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
)

func checkBundleCached() (bool, error) {
	if _, err := os.Stat(config.GetString(cmdConfig.Bundle.Name)); os.IsNotExist(err) {
		return false, err
	}
	return true, nil
}

func fixBundleCached() (bool, error) {
	if constants.BundleEmbedded() {
		currentExecutable, err := osext.Executable()
		if err != nil {
			return false, err
		}
		extractor, err := binappend.MakeExtractor(currentExecutable)
		if err != nil {
			return false, err
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
		return false, errors.New("oc binary is not cached.")
	}
	logging.Debug("oc binary already cached")
	return true, nil
}

func fixOcBinaryCached() (bool, error) {
	oc := oc.OcCached{}
	if err := oc.EnsureIsCached(); err != nil {
		return false, fmt.Errorf("Not able to download oc %v", err)
	}
	logging.Debug("oc binary cached")
	return true, nil
}

func checkIfRunningAsNormalUser() (bool, error) {
	if os.Geteuid() != 0 {
		return true, nil
	}
	logging.Debug("Ran as root")
	return false, fmt.Errorf("crc should be ran as a normal user")
}

func fixRunAsNormalUser() (bool, error) {
	return false, fmt.Errorf("crc should be ran as a normal user")
}
