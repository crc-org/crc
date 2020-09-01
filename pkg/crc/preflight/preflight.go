package preflight

import (
	"fmt"

	cfg "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/logging"
)

type Flags uint32

// EnableExperimentalFeatures enables the use of experimental features
var EnableExperimentalFeatures bool

const (
	// Indicates a PreflightCheck should only be run as part of "crc setup"
	SetupOnly Flags = 1 << iota
	// Indicates a PreflightCheck should only be run as part of "crc start"
	StartOnly
	NoFix
	CleanUpOnly
)

type CheckFunc func() error
type FixFunc func() error
type CleanUpFunc func() error

type Check struct {
	configKeySuffix    string
	checkDescription   string
	check              CheckFunc
	fixDescription     string
	fix                FixFunc
	flags              Flags
	cleanupDescription string
	cleanup            CleanUpFunc
}

func (check *Check) getSkipConfigName() string {
	if check.configKeySuffix == "" {
		return ""
	}
	return "skip-" + check.configKeySuffix
}

func (check *Check) shouldSkip() bool {
	if check.configKeySuffix == "" {
		return false
	}
	return cfg.GetBool(check.getSkipConfigName())
}

func (check *Check) getWarnConfigName() string {
	if check.configKeySuffix == "" {
		return ""
	}
	return "warn-" + check.configKeySuffix
}

func (check *Check) shouldWarn() bool {
	if check.configKeySuffix == "" {
		return false
	}
	return cfg.GetBool(check.getWarnConfigName())
}

func (check *Check) doCheck() error {
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

func (check *Check) doFix() error {
	if check.fixDescription == "" {
		panic(fmt.Sprintf("Should not happen, empty description for fix '%s'", check.configKeySuffix))
	}
	if check.flags&NoFix == NoFix {
		return fmt.Errorf(check.fixDescription)
	}

	logging.Infof("%s", check.fixDescription)

	return check.fix()
}

func (check *Check) doCleanUp() error {
	if check.cleanupDescription == "" {
		panic(fmt.Sprintf("Should not happen, empty description for cleanup '%s'", check.configKeySuffix))
	}

	logging.Infof("%s", check.cleanupDescription)

	return check.cleanup()
}

func doPreflightChecks(checks []Check) {
	for _, check := range checks {
		if check.flags&SetupOnly == SetupOnly || check.flags&CleanUpOnly == CleanUpOnly {
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

func doFixPreflightChecks(checks []Check) {
	for _, check := range checks {
		if check.flags&StartOnly == StartOnly || check.flags&CleanUpOnly == CleanUpOnly {
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

func doCleanUpPreflightChecks(checks []Check) {
	// Do the cleanup in reverse order to avoid any dependency during cleanup
	for i := len(checks) - 1; i >= 0; i-- {
		check := checks[i]
		if check.cleanup == nil {
			continue
		}
		err := check.doCleanUp()
		if err != nil {
			logging.Fatal(err.Error())
		}
	}
}

func doRegisterSettings(checks []Check) {
	for _, check := range checks {
		if check.configKeySuffix != "" {
			cfg.AddSetting(check.getSkipConfigName(), false, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
			cfg.AddSetting(check.getWarnConfigName(), false, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
		}
	}
}

// StartPreflightChecks performs the preflight checks before starting the cluster
func StartPreflightChecks() {
	doPreflightChecks(getPreflightChecks())
}

// SetupHost performs the prerequisite checks and setups the host to run the cluster
func SetupHost() {
	doFixPreflightChecks(getPreflightChecks())
}

func RegisterSettings() {
	doRegisterSettings(getPreflightChecks())
}
