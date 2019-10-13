package config

import (
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	ConfigCmd.AddCommand(configGetCmd)
}

var configGetCmd = &cobra.Command{
	Use:   "get CONFIG-KEY",
	Short: "Get a crc configuration property",
	Long:  `Gets a crc configuration property.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			output.Outln("Please provide a configuration property to get")
			os.Exit(1)
		}
		runConfigGet(args[0])
	},
}

func runConfigGet(key string) {
	v, err := config.Get(key)
	if err != nil {
		errors.ExitWithMessage(1, err.Error())
	}
	output.Outln(key, ":", v)
}
