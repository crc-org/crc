package config

import (
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/spf13/cobra"
)

func init() {
	ConfigCmd.AddCommand(configUnsetCmd)
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset CONFIG-KEY",
	Short: "Unsets a crc configuration property.",
	Long: `Unsets a crc configuration property. Some of the configuration properties are equivalent
to the options that you set when you run the 'crc start' command.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			errors.ExitWithMessage(1, "Please provide only a configuration property to unset")
		}
		runConfigUnset(args[0])
	},
}

func runConfigUnset(key string) {
	_, ok := config.SettingsList[key]
	if !ok {
		errors.ExitWithMessage(1, "Config property does not exist: %s", key)
	}

	if !config.IsSet(key) {
		errors.ExitWithMessage(1, "Config property is not set: %s", key)
	}
	if err := config.Unset(key); err != nil {
		errors.ExitWithMessage(1, "Error unsetting config property: %s : %v", key, err)
	}
	if err := config.WriteConfig(); err != nil {
		errors.ExitWithMessage(1, "Error Writing config to file %s: %s", constants.ConfigPath, err.Error())
	}
}
