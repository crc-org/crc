package config

import (
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/exit"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/spf13/cobra"
)

func init() {
	configCmd.AddCommand(configSetCmd)
}

var configSetCmd = &cobra.Command{
	Use:   "set CONFIG-KEY VALUE",
	Short: "Set a crc configuration property",
	Long:  `Sets a crc configuration property.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			exit.WithMessage(1, "Please provide a configuration property and its value as in 'crc config set KEY VALUE'")
		}
		runConfigSet(args[0], args[1])
	},
}

func runConfigSet(key string, value interface{}) {
	setMessage, err := config.Set(key, value)
	if err != nil {
		exit.WithMessage(1, err.Error())
	}

	if err := config.WriteConfig(); err != nil {
		exit.WithMessage(1, "Error writing configuration to file '%s': %s", constants.ConfigPath, err.Error())
	}

	if setMessage != "" {
		output.Outln(setMessage)
	}
}
