package preflight

import (
	"bufio"
	"fmt"
	"os"

	"github.com/YourFin/binappend"
	"github.com/kardianos/osext"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
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
		w.Write(data)
		return true, nil
	}
	return false, fmt.Errorf("CRC bundle is not embedded in the binary, see 'crc help' for more details.")
}
