package config

import (
	validations "github.com/code-ready/crc/pkg/crc/config"
	"github.com/spf13/cobra"
	"strings"
)

type validationFnType func(interface{}, interface{}) bool

type setting struct {
	Name          string
	DefaultValue  interface{}
	ValidationFns []validationFnType
}

// SettingsList holds all the config settings
var SettingsList = make(map[string]setting)

var (
	// Preflight checks
	SkipCheckVirtEnabled        = createSetting("skip-check-virt-enabled", false, []validationFnType{validations.ValidateBool})
	WarnCheckVirtEnabled        = createSetting("warn-check-virt-enabled", false, []validationFnType{validations.ValidateBool})
	SkipCheckKvmEnabled         = createSetting("skip-check-kvm-enabled", false, []validationFnType{})
	WarnCheckKvmEnabled         = createSetting("warn-check-kvm-enabled", false, []validationFnType{})
	SkipCheckLibvirtInstalled   = createSetting("skip-check-libvirt-installed", false, []validationFnType{})
	WarnCheckLibvirtInstalled   = createSetting("warn-check-libvirt-installed", false, []validationFnType{})
	SkipCheckLibvirtEnabled     = createSetting("skip-check-libvirt-enabled", false, []validationFnType{validations.ValidateBool})
	WarnCheckLibvirtEnabled     = createSetting("warn-check-libvirt-enabled", true, []validationFnType{})
	SkipCheckLibvirtRunning     = createSetting("skip-check-libvirt-running", false, []validationFnType{})
	WarnCheckLibvirtRunning     = createSetting("warn-check-libvirt-running", false, []validationFnType{})
	SkipCheckUserInLibvirtGroup = createSetting("skip-check-user-in-libvirt-group", false, []validationFnType{})
	WarnCheckUserInLibvirtGroup = createSetting("warn-check-user-in-libvirt-group", false, []validationFnType{})
	SkipCheckIPForwarding       = createSetting("skip-check-ip-forwarding", false, []validationFnType{})
	WarnCheckIPForwarding       = createSetting("warn-check-ip-forwarding", false, []validationFnType{})
	SkipCheckKvmDriver          = createSetting("skip-check-kvm-driver", false, []validationFnType{})
	WarnCheckKvmDriver          = createSetting("warn-check-kvm-driver", false, []validationFnType{})
	SkipCheckDefaultPool        = createSetting("skip-check-default-pool", false, []validationFnType{})
	WarnCheckDefaultPool        = createSetting("warn-check-default-pool", false, []validationFnType{})
	SkipCheckDefaultPoolSpace   = createSetting("skip-check-default-pool", false, []validationFnType{})
	WarnCheckDefaultPoolSpace   = createSetting("warn-check-default-pool", false, []validationFnType{})
	SkipCheckCrcNetwork         = createSetting("skip-check-crc-network", false, []validationFnType{})
	WarnCheckCrcNetwork         = createSetting("warn-check-crc-network", false, []validationFnType{})
	SkipCheckCrcNetworkActive   = createSetting("skip-check-crc-network-active", false, []validationFnType{})
	WarnCheckCrcNetworkActive   = createSetting("warn-check-crc-network-active", false, []validationFnType{})
)

// CreateSetting returns a filled struct of ConfigSetting
// takes the config name and default value as arguments
func createSetting(name string, defValue interface{}, validationFn []validationFnType) *setting {
	s := setting{Name: name, DefaultValue: defValue, ValidationFns: validationFn}
	SettingsList[name] = s
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
