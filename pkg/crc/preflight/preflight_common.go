package preflight

import (
	"io/ioutil"
	"os"

	"github.com/code-ready/crc/bundle_bindata"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
)

type PreflightCheckFixFuncType func() (bool, error)

func preflightCheckSucceedsOrFails(configuredToSkip bool, check PreflightCheckFixFuncType, message string, configuredToWarn bool) {
	logging.InfoF("%s", message)
	if configuredToSkip {
		logging.Warn("Skipping above check ...")
		return
	}

	ok, err := check()
	if ok {
		return
	}

	if configuredToWarn {
		logging.Warn(err.Error())
		return
	}

	logging.Fatal(err.Error())
}

func preflightCheckAndFix(configuredToSkip bool, check, fix PreflightCheckFixFuncType, message string, configuredToWarn bool) {
	logging.InfoF("%s", message)
	if configuredToSkip {
		logging.Warn("Skipping above check ...")
		return
	}

	ok, err := check()
	if ok {
		return
	}

	if configuredToWarn {
		logging.Warn(err.Error())
		return
	}
	logging.Debug(err.Error())

	ok, err = fix()
	if ok {
		return
	}

	logging.Fatal(err.Error())
}

func checkBundlePresent() (bool, error) {
	if _, err := os.Stat(constants.DefaultBundlePath); os.IsNotExist(err) {
		return false, err
	}
	return true, nil
}

func fixBundlePresent() (bool, error) {
	// unpack bundle using bindata.Asset
	data, err := bundle_bindata.Asset(constants.GetDefaultBundle())
	if err != nil {
		return false, err
	}
	err = ioutil.WriteFile(constants.DefaultBundlePath, data, 0664)
	if err != nil {
		return false, err
	}
	return true, nil
}
