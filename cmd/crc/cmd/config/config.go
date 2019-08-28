package config

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/spf13/cobra"
	"strings"

	cfg "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/machine"
)

var (
	// Start command settings in config
	VMDriver       = cfg.AddSetting("vm-driver", machine.DefaultDriver.Driver, []cfg.ValidationFnType{cfg.ValidateDriver})
	Bundle         = cfg.AddSetting("bundle", nil, []cfg.ValidationFnType{cfg.ValidateBundle})
	CPUs           = cfg.AddSetting("cpus", constants.DefaultCPUs, []cfg.ValidationFnType{cfg.ValidateCPUs})
	Memory         = cfg.AddSetting("memory", constants.DefaultMemory, []cfg.ValidationFnType{cfg.ValidateMemory})
	NameServer     = cfg.AddSetting("nameserver", nil, []cfg.ValidationFnType{cfg.ValidateIpAddress})
	PullSecretFile = cfg.AddSetting("pull-secret-file", nil, []cfg.ValidationFnType{cfg.ValidatePath})

	// Preflight checks
	SkipCheckVirtualBoxInstalled = cfg.AddSetting("skip-check-virtualbox-installed", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckVirtualBoxInstalled = cfg.AddSetting("warn-check-virtualbox-installed", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	SkipCheckBundleCached        = cfg.AddSetting("skip-check-bundle-cached", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckBundleCached        = cfg.AddSetting("warn-check-bundle-cached", true, []cfg.ValidationFnType{cfg.ValidateBool})
)

var (
	ConfigCmd = &cobra.Command{
		Use:   "config SUBCOMMAND [flags]",
		Short: "Modifies crc configuration properties.",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
)

func init() {
	ConfigCmd.Long = `Modifies crc configuration properties. Some of the configuration properties are equivalent
to the options that you set when you run the 'crc start' command.
Configurable properties (enter as SUBCOMMAND): ` + "\n\n" + configurableFields()
}

func configurableFields() string {
	var fields []string
	for _, key := range cfg.AllConfigKeys() {
		fields = append(fields, " * "+key)
	}
	return strings.Join(fields, "\n")
}
