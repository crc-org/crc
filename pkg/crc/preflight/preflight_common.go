package preflight

import (
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
