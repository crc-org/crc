package config

import (
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/spf13/cobra"
)

func init() {
	ConfigCmd.AddCommand(configViewCmd)
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Display complete crc configurations.",
	Long: `Retrieves full list of crc configurations. Some of the configuration properties are equivalent
to the options that you set when you run the 'crc start' command.`,
	Run: func(cmd *cobra.Command, args []string) {
		runConfigView()
	},
}

func runConfigView() {
	settings := config.AllConfigs()
	for k, v := range settings {
		output.Out(k, ":", v)
	}
}
