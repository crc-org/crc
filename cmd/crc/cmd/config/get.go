package config

import (
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/exit"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/spf13/cobra"
)

func init() {
	configCmd.AddCommand(configGetCmd)
}

var configGetCmd = &cobra.Command{
	Use:   "get CONFIG-KEY",
	Short: "Get a crc configuration property",
	Long:  `Gets a crc configuration property.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			exit.WithMessage(1, "Please provide a configuration property to get")
		}
		runConfigGet(args[0])
	},
}

func runConfigGet(key string) {
	v, err := config.Get(key)
	if err != nil {
		exit.WithMessage(1, err.Error())
	}
	output.Outln(key, ":", v)
}
