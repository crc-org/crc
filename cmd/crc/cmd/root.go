package cmd

import (
	"github.com/spf13/cobra"
	"os"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/output"
)

var rootCmd = &cobra.Command{
	Use:   commandName,
	Short: descriptionShort,
	Long:  descriptionLong,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		runPrerun()
	},
	Run: func(cmd *cobra.Command, args []string) {
		runRoot()
	},
}

func init() {
	if err := config.EnsureConfigFileExists(); err != nil {
		output.Out(err.Error())
	}
	config.InitViper()
	setConfigDefaults()
	rootCmd.AddCommand(cmdConfig.ConfigCmd)
}

func runPrerun() {
	output.Out("%s - %s", commandName, descriptionShort)
}

func runRoot() {
	output.Out("No command given")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logging.Log("ERR: %s", err.Error())
		os.Exit(1)
	}
}

func setConfigDefaults() {
	for _, setting := range cmdConfig.SettingsList {
		config.SetDefault(setting.Name, setting.DefaultValue)
	}
}
