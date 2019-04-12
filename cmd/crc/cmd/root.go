package cmd

import (
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/spf13/cobra"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/config"
	log "github.com/code-ready/crc/pkg/crc/logging"
)

var logLevel string

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
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		runPostrun()
	},
}

func init() {
	if err := config.EnsureConfigFileExists(); err != nil {
		output.Out(err.Error())
	}
	config.InitViper()
	setConfigDefaults()
	rootCmd.AddCommand(cmdConfig.ConfigCmd)
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (e.g. \"debug | info | warn | error\")")
}

func runPrerun() {
	output.OutF("%s - %s\n", commandName, descriptionShort)

	// Setting up logrus
	log.InitLogrus(logLevel)
}

func runPostrun() {
	log.CloseLogFile()
}

func runRoot() {
	output.Out("No command given")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func setConfigDefaults() {
	for _, setting := range cmdConfig.SettingsList {
		config.SetDefault(setting.Name, setting.DefaultValue)
	}
}
