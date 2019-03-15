package preflight

import (
	"fmt"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/spf13/viper"
	"os"
)

type PreflightCheckFixFuncType func() (bool, error)

func preflightCheckSucceedsOrFails(configNameOverrideIfSkipped string, check PreflightCheckFixFuncType, message string, configNameOverrideIfWarning string) {
	output.Out("-- %s ... ", message)
	isConfiguredToSkip := viper.GetBool(configNameOverrideIfSkipped)
	isConfiguredToWarn := viper.GetBool(configNameOverrideIfWarning)

	if isConfiguredToSkip {
		output.Out("SKIP")
		return
	}

	ok, err := check()
	if ok {
		output.Out("OK")
		return
	}

	errorMessage := fmt.Sprintf("   %s", err.Error())
	if isConfiguredToWarn {
		output.Out("WARN")
		output.Out(errorMessage)
		return
	}

	output.Out("FAIL")
	output.Out(errorMessage)
	os.Exit(1)
}

func preflightCheckAndFix(check, fix PreflightCheckFixFuncType, message string) {
	output.Out("-- %s ... ", message)

	if ok, _ := check(); ok {
		output.Out("OK")
		return
	}
	ok, err := fix()
	if ok {
		output.Out("OK")
		return
	}

	output.Out("FAIL")
	output.Out(err.Error())
}
