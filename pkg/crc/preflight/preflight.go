package preflight

import (
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/logging"
)

type PreflightCheckFlags uint32

const (
	// Indicates a PreflightCheck should only be run as part of "crc setup"
	SetupOnly PreflightCheckFlags = 1 << iota
	// Indicates a PreflightCheck should only be run as part of "crc start"
	StartOnly
)

type PreflightCheckFunc func() (bool, error)
type PreflightFixFunc func() (bool, error)

type PreflightCheck struct {
	skipConfigName   string
	warnConfigName   string
	checkDescription string
	check            PreflightCheckFunc
	fixDescription   string
	fix              PreflightFixFunc
	flags            PreflightCheckFlags
}

func (check *PreflightCheck) doCheck() error {
	if check.checkDescription == "" {
		// warning, this is a programming error
	} else {
		logging.Infof("%s", check.checkDescription)
	}
	if check.skipConfigName != "" && config.GetBool(check.skipConfigName) {
		logging.Warn("Skipping above check ...")
		return nil
	}

	_, err := check.check()
	if err != nil {
		logging.Debug(err.Error())
	}
	return err
}

func (check *PreflightCheck) doFix() error {
	if check.fixDescription == "" {
		// warning, this is a programming error
	} else {
		logging.Infof("%s", check.fixDescription)
	}
	ok, err := check.fix()
	if ok {
		return nil
	}

	return err
}

func doPreflightChecks(checks []PreflightCheck) {
	for _, check := range checks {
		if check.flags&SetupOnly == SetupOnly {
			continue
		}
		err := check.doCheck()
		if err != nil {
			if check.warnConfigName != "" && config.GetBool(check.warnConfigName) {
				logging.Warn(err.Error())
			} else {
				logging.Fatal(err.Error())
			}
		}
	}
}

func doFixPreflightChecks(checks []PreflightCheck) {
	for _, check := range checks {
		if check.flags&StartOnly == StartOnly {
			continue
		}
		err := check.doCheck()
		if err == nil {
			continue
		}
		err = check.doFix()
		if err != nil {
			if check.warnConfigName != "" && config.GetBool(check.warnConfigName) {
				logging.Warn(err.Error())
			} else {
				logging.Fatal(err.Error())
			}
		}
	}
}
