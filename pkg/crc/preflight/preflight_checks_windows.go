package preflight

import (
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/oc"
)

// Check if oc binary is cached or not
func checkOcBinaryCached() (bool, error) {
	oc := oc.OcCached{}
	if !oc.IsCached() {
		return false, errors.New("oc binary is not cached.")
	}
	return true, nil
}

func fixOcBinaryCached() (bool, error) {
	oc := oc.OcCached{}
	if err := oc.EnsureIsCached(); err != nil {
		return false, errors.NewF("Not able to download oc %v", err)
	}
	return true, nil
}
