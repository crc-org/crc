package config

import (
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/output"
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
	unsetMsg, err := config.Unset(key)
	if err != nil {
		errors.ExitWithMessage(1, err.Error())
	}
	if err := config.WriteConfig(); err != nil {
		errors.ExitWithMessage(1, "Error Writing config to file %s: %s", constants.ConfigPath, err.Error())
	}

	if unsetMsg != "" {
		output.Outln(unsetMsg)
	}
}
