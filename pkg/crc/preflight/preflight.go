package preflight

import (
	"fmt"

	cfg "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
)

type PreflightCheckFlags uint32

const (
	// Indicates a PreflightCheck should only be run as part of "crc setup"
	SetupOnly PreflightCheckFlags = 1 << iota
	// Indicates a PreflightCheck should only be run as part of "crc start"
	StartOnly
	NoFix
)

type PreflightCheckFunc func() error
type PreflightFixFunc func() error

type PreflightCheck struct {
	configKeySuffix  string
	checkDescription string
	check            PreflightCheckFunc
	fixDescription   string
	fix              PreflightFixFunc
	flags            PreflightCheckFlags
}

func (check *PreflightCheck) getSkipConfigName() string {
	if check.configKeySuffix == "" {
		return ""
	}
	return "skip-" + check.configKeySuffix
}

func (check *PreflightCheck) shouldSkip() bool {
	if check.configKeySuffix == "" {
		return false
	}
	return cfg.GetBool(check.getSkipConfigName())
}

func (check *PreflightCheck) getWarnConfigName() string {
	if check.configKeySuffix == "" {
		return ""
	}
	return "warn-" + check.configKeySuffix
}

func (check *PreflightCheck) shouldWarn() bool {
	if check.configKeySuffix == "" {
		return false
	}
	return cfg.GetBool(check.getWarnConfigName())
}

func (check *PreflightCheck) doCheck() error {
	if check.checkDescription == "" {
		panic(fmt.Sprintf("Should not happen, empty description for check '%s'", check.configKeySuffix))
	} else {
		logging.Infof("%s", check.checkDescription)
	}
	if check.shouldSkip() {
		logging.Warn("Skipping above check ...")
		return nil
	}

	err := check.check()
	if err != nil {
		logging.Debug(err.Error())
	}
	return err
}

func (check *PreflightCheck) doFix() error {
	if check.fixDescription == "" {
		panic(fmt.Sprintf("Should not happen, empty description for fix '%s'", check.configKeySuffix))
	}
	if check.flags&NoFix == NoFix {
		return errors.Newf(check.fixDescription)
	}

	logging.Infof("%s", check.fixDescription)

	return check.fix()
}

func doPreflightChecks(checks []PreflightCheck) {
	for _, check := range checks {
		if check.flags&SetupOnly == SetupOnly {
			continue
		}
		err := check.doCheck()
		if err != nil {
			if check.shouldWarn() {
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
			if check.shouldWarn() {
				logging.Warn(err.Error())
			} else {
				logging.Fatal(err.Error())
			}
		}
	}
}

func doRegisterSettings(checks []PreflightCheck) {
	for _, check := range checks {
		if check.configKeySuffix != "" {
			cfg.AddSetting(check.getSkipConfigName(), false, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
			cfg.AddSetting(check.getWarnConfigName(), false, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
		}
	}
}
