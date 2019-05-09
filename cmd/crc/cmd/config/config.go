package config

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/spf13/cobra"
	"strings"

	validations "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/machine"
)

// validationFnType takes the key, value as args and checks if valid
type validationFnType func(interface{}) (bool, string)

type setting struct {
	Name          string
	DefaultValue  interface{}
	ValidationFns []validationFnType
}

// SettingsList holds all the config settings
var SettingsList = make(map[string]*setting)

var (
	// Start command settings in config
	VMDriver = createSetting("vm-driver", machine.DefaultDriver.Driver, []validationFnType{validations.ValidateDriver})
	Bundle   = createSetting("bundle", nil, []validationFnType{validations.ValidateBundle})
	CPUs     = createSetting("cpus", constants.DefaultCPUs, []validationFnType{validations.ValidateCPUs})
	Memory   = createSetting("memory", constants.DefaultMemory, []validationFnType{validations.ValidateMemory})

	// Preflight checks
	SkipCheckVirtualBoxInstalled = createSetting("skip-check-virtualbox-installed", nil, []validationFnType{validations.ValidateBool})
	WarnCheckVirtualBoxInstalled = createSetting("warn-check-virtualbox-installed", nil, []validationFnType{validations.ValidateBool})
)

// CreateSetting returns a filled struct of ConfigSetting
// takes the config name and default value as arguments
func createSetting(name string, defValue interface{}, validationFn []validationFnType) *setting {
	s := setting{Name: name, DefaultValue: defValue, ValidationFns: validationFn}
	SettingsList[name] = &s
	return &s
}

var (
	ConfigCmd = &cobra.Command{
		Use:   "config SUBCOMMAND [flags]",
		Short: "Modifies crc configuration properties.",
		Long: `Modifies crc configuration properties. Some of the configuration properties are equivalent
to the options that you set when you run the 'crc start' command.
Configurable properties (enter as SUBCOMMAND): ` + "\n\n" + configurableFields(),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

func configurableFields() string {
	var fields []string
	for _, s := range SettingsList {
		fields = append(fields, " * "+s.Name)
	}
	return strings.Join(fields, "\n")
}
