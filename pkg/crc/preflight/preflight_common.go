package preflight

import (
	log "github.com/code-ready/crc/pkg/crc/logging"
)

type PreflightCheckFixFuncType func() (bool, error)

func preflightCheckSucceedsOrFails(configuredToSkip bool, check PreflightCheckFixFuncType, message string, configuredToWarn bool) {
	log.InfoF(" %s", message)
	if configuredToSkip {
		log.Warn(" Skipping above check ...")
		return
	}

	ok, err := check()
	if ok {
		return
	}

	if configuredToWarn {
		log.Warn(err.Error())
		return
	}

	log.Fatal(err.Error())
}

func preflightCheckAndFix(configuredToSkip bool, check, fix PreflightCheckFixFuncType, message string, configuredToWarn bool) {
	log.InfoF(" %s", message)
	if configuredToSkip {
		log.Warn(" Skipping above check ...")
		return
	}

	ok, err := check()
	if ok {
		return
	}

	if configuredToWarn {
		log.Warn(err.Error())
		return
	}

	ok, err = fix()
	if ok {
		return
	}

	log.Fatal(err.Error())
}
