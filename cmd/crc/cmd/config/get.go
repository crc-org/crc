package config

import (
	"fmt"

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
	v := config.Get(key)
	switch {
	case v.Invalid:
		exit.WithMessage(1, fmt.Sprintf("Configuration property '%s' does not exist", key))
	case v.IsDefault:
		exit.WithMessage(1, fmt.Sprintf("Configuration property '%s' is not set. Default value is '%s'", key, v.AsString()))
	default:
		output.Outln(key, ":", v.AsString())
	}
}
