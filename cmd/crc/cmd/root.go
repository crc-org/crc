package cmd

import (
	"fmt"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/exit"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/spf13/cobra"
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
		_ = cmd.Help()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		runPostrun()
	},
}

var globalForce bool

func init() {
	if err := constants.EnsureBaseDirExists(); err != nil {
		logging.Fatal(err.Error())
	}
	if err := config.EnsureConfigFileExists(); err != nil {
		logging.Fatal(err.Error())
	}
	if err := config.InitViper(); err != nil {
		logging.Fatal(err.Error())
	}

	preflight.RegisterSettings()
	setConfigDefaults()

	// subcommands
	rootCmd.AddCommand(cmdConfig.GetConfigCmd())

	rootCmd.PersistentFlags().StringVar(&logging.LogLevel, "log-level", constants.DefaultLogLevel, "log level (e.g. \"debug | info | warn | error\")")
	rootCmd.PersistentFlags().BoolVarP(&globalForce, "force", "f", false, "Forcefully perform an action")
}

func runPrerun() {
	// Setting up logrus
	logging.InitLogrus(logging.LogLevel, constants.LogFilePath)
	setProxyDefaults()
	for _, str := range GetVersionStrings() {
		logging.Debugf(str)
	}
}

func runPostrun() {
	logging.CloseLogging()
}

func runRoot() {
	output.Outln("No command given")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logging.Fatal(err)
	}
}

func setConfigDefaults() {
	config.SetDefaults()
}

func checkIfMachineMissing(name string) error {
	exists, err := machine.Exists(name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Machine '%s' does not exist. Use 'crc start' to create it", constants.DefaultName)
	}
	return nil
}

func setProxyDefaults() {
	httpProxy := config.GetString(cmdConfig.HTTPProxy.Name)
	httpsProxy := config.GetString(cmdConfig.HTTPSProxy.Name)
	noProxy := config.GetString(cmdConfig.NoProxy.Name)

	proxyConfig, err := network.NewProxyDefaults(httpProxy, httpsProxy, noProxy)
	if err != nil {
		exit.WithMessage(1, err.Error())
	}

	if proxyConfig.IsEnabled() {
		logging.Debugf("HTTP-PROXY: %s, HTTPS-PROXY: %s, NO-PROXY: %s", proxyConfig.HTTPProxyForDisplay(),
			proxyConfig.HTTPSProxyForDisplay(), proxyConfig.GetNoProxyString())
		proxyConfig.ApplyToEnvironment()
	}
}
