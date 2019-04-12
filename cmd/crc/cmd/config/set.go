package config

import (
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	ConfigCmd.AddCommand(configSetCmd)
}

var configSetCmd = &cobra.Command{
	Use:   "set CONFIG-KEY VALUE",
	Short: "Sets a crc configuration property.",
	Long: `Sets a crc configuration property. Some of the configuration properties are equivalent
to the options that you set when you run the 'crc start' command.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			output.Out("Please a provide a configuration property and value to set")
			os.Exit(1)
		}
		runConfigSet(args[0], args[1])
	},
}

func runConfigSet(key string, value interface{}) {
	if !config.Exists(key) {
		output.Out("Config property doesnot exist:", key)
		os.Exit(1)
	}

	_, ok := SettingsList[key]
	if !ok {
		output.Out("Config property doesnot exist:", key)
		os.Exit(1)
	}

	if !runValidations(SettingsList[key].ValidationFns, value) {
		output.Out("Config value is invalid:", value)
	}

	config.Set(key, value)
	if err := config.WriteConfig(); err != nil {
		output.Out("Error Writing config to file:", constants.ConfigPath, err.Error())
	}
}

func runValidations(validations []validationFnType, value interface{}) bool {
	for _, fn := range validations {
		if !fn(value) {
			return false
		}
	}
	return true
}
