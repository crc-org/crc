package preflight

import (
	"fmt"

	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
)

type Flags uint32

const (
	// Indicates a PreflightCheck should only be run as part of "crc setup"
	SetupOnly Flags = 1 << iota
	NoFix
	CleanUpOnly
	StartUpOnly
)

var (
	preset string
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

	labels labels
}

func (check *Check) getSkipConfigName() string {
	if check.configKeySuffix == "" {
		return ""
	}
	return "skip-" + check.configKeySuffix
}

func (check *Check) shouldSkip(config crcConfig.Storage) bool {
	if check.configKeySuffix == "" {
		return false
	}
	return config.Get(check.getSkipConfigName()).AsBool()
}

func (check *Check) doCheck(config crcConfig.Storage) error {
	if check.checkDescription == "" {
		panic(fmt.Sprintf("Should not happen, empty description for check '%s'", check.configKeySuffix))
	} else {
		logging.Infof("%s", check.checkDescription)
	}
	if check.shouldSkip(config) {
		logging.Warn("Skipping above check...")
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

func doPreflightChecks(config crcConfig.Storage, checks []Check) error {
	for _, check := range checks {
		if check.flags&SetupOnly == SetupOnly || check.flags&CleanUpOnly == CleanUpOnly {
			continue
		}
		if err := check.doCheck(config); err != nil {
			return err
		}
	}
	return nil
}

func doFixPreflightChecks(config crcConfig.Storage, checks []Check, checkOnly bool) error {
	for _, check := range checks {
		if check.flags&CleanUpOnly == CleanUpOnly || check.flags&StartUpOnly == StartUpOnly {
			continue
		}
		err := check.doCheck(config)
		if err == nil {
			continue
		} else if checkOnly {
			return err
		}
		if err = check.doFix(); err != nil {
			return err
		}
	}
	return nil
}

func doCleanUpPreflightChecks(checks []Check) error {
	var mErr errors.MultiError
	// Do the cleanup in reverse order to avoid any dependency during cleanup
	for i := len(checks) - 1; i >= 0; i-- {
		check := checks[i]
		if check.cleanup == nil {
			continue
		}
		err := check.doCleanUp()
		if err != nil {
			// If an error occurs in a cleanup function
			// we log/collect it and  move to the  next
			logging.Debug(err)
			mErr.Collect(err)
		}
	}
	if len(mErr.Errors) == 0 {
		return nil
	}
	return mErr
}

func doRegisterSettings(cfg crcConfig.Schema, checks []Check) {
	for _, check := range checks {
		if check.configKeySuffix != "" {
			cfg.AddSetting(check.getSkipConfigName(), false, crcConfig.ValidateBool, crcConfig.SuccessfullyApplied,
				"Skip preflight check (true/false, default: false)")
		}
	}
}

// StartPreflightChecks performs the preflight checks before starting the cluster
func StartPreflightChecks(config crcConfig.Storage) error {
	experimentalFeatures := config.Get(crcConfig.ExperimentalFeatures).AsBool()
	mode := crcConfig.GetNetworkMode(config)
	trayAutostart := config.Get(crcConfig.AutostartTray).AsBool()
	bundlePath := config.Get(crcConfig.Bundle).AsString()
	preset = config.Get(crcConfig.Preset).AsString()
	if err := doPreflightChecks(config, getPreflightChecks(experimentalFeatures, trayAutostart, mode, bundlePath, preset)); err != nil {
		return &errors.PreflightError{Err: err}
	}
	return nil
}

// SetupHost performs the prerequisite checks and setups the host to run the cluster
func SetupHost(config crcConfig.Storage, checkOnly bool) error {
	experimentalFeatures := config.Get(crcConfig.ExperimentalFeatures).AsBool()
	mode := crcConfig.GetNetworkMode(config)
	trayAutostart := config.Get(crcConfig.AutostartTray).AsBool()
	bundlePath := config.Get(crcConfig.Bundle).AsString()
	preset = config.Get(crcConfig.Preset).AsString()
	logging.Infof("Using bundle path %s", bundlePath)
	return doFixPreflightChecks(config, getPreflightChecks(experimentalFeatures, trayAutostart, mode, bundlePath, preset), checkOnly)
}

func RegisterSettings(config crcConfig.Schema) {
	doRegisterSettings(config, getAllPreflightChecks())
}

func CleanUpHost() error {
	// A user can use setup with experiment flag
	// and not use cleanup with same flag, to avoid
	// any extra step/confusion we are just adding the checks
	// which are behind the experiment flag. This way cleanup
	// perform action in a sane way.
	return doCleanUpPreflightChecks(getAllPreflightChecks())
}
