package config

import (
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	ConfigCmd.AddCommand(configGetCmd)
}

var configGetCmd = &cobra.Command{
	Use:   "get CONFIG-KEY",
	Short: "Gets a crc configuration property.",
	Long: `Gets a crc configuration property. Some of the configuration properties are equivalent
to the options that you set when you run the 'crc start' command.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			output.Out("Please a provide a configuration property to retrieve")
			os.Exit(1)
		}
		runConfigGet(args[0])
	},
}

func runConfigGet(key string) {
	if v, ok := config.ViperConfig[key]; ok {
		output.Out(key, ":", v)
		return
	}
	output.Out("Config not found:", key)
}
