package preflight

import (
	"errors"
	"fmt"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
)

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
