package cmd

import (
	"fmt"
	"io/ioutil"
	"strings"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
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

var (
	globalForce bool
	viper       *crcConfig.ViperStorage
	config      *crcConfig.Config
)

func init() {
	if err := constants.EnsureBaseDirExists(); err != nil {
		logging.Fatal(err.Error())
	}
	var err error
	config, viper, err = newViperConfig()
	if err != nil {
		logging.Fatal(err.Error())
	}

	// subcommands
	rootCmd.AddCommand(cmdConfig.GetConfigCmd(config))

	rootCmd.PersistentFlags().StringVar(&logging.LogLevel, "log-level", constants.DefaultLogLevel, "log level (e.g. \"debug | info | warn | error\")")
	rootCmd.PersistentFlags().BoolVarP(&globalForce, "force", "f", false, "Forcefully perform an action")
}

func runPrerun() {
	// Setting up logrus
	logging.InitLogrus(logging.LogLevel, constants.LogFilePath)
	setProxyDefaults()

	for _, str := range defaultVersion().lines() {
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

func checkIfMachineMissing(client machine.Client, name string) error {
	exists, err := client.Exists(name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Machine '%s' does not exist. Use 'crc start' to create it", name)
	}
	return nil
}

func setProxyDefaults() {
	httpProxy := config.Get(cmdConfig.HTTPProxy).AsString()
	httpsProxy := config.Get(cmdConfig.HTTPSProxy).AsString()
	noProxy := config.Get(cmdConfig.NoProxy).AsString()
	proxyCAFile := config.Get(cmdConfig.ProxyCAFile).AsString()

	proxyCAData, err := getProxyCAData(proxyCAFile)
	if err != nil {
		exit.WithMessage(1, fmt.Sprintf("not able to read proxyCAFile %s: %v", proxyCAFile, err.Error()))
	}

	proxyConfig, err := network.NewProxyDefaults(httpProxy, httpsProxy, noProxy, proxyCAData)
	if err != nil {
		exit.WithMessage(1, err.Error())
	}

	if proxyConfig.IsEnabled() {
		logging.Debugf("HTTP-PROXY: %s, HTTPS-PROXY: %s, NO-PROXY: %s, proxyCAFile: %s", proxyConfig.HTTPProxyForDisplay(),
			proxyConfig.HTTPSProxyForDisplay(), proxyConfig.GetNoProxyString(), proxyCAFile)
		proxyConfig.ApplyToEnvironment()
	}
}

func getProxyCAData(proxyCAFile string) (string, error) {
	if proxyCAFile == "" {
		return "", nil
	}
	proxyCACert, err := ioutil.ReadFile(proxyCAFile)
	if err != nil {
		return "", err
	}
	// Before passing string back to caller function, remove the empty lines in the end
	return trimTrailingEOL(string(proxyCACert)), nil
}

func trimTrailingEOL(s string) string {
	return strings.TrimRight(s, "\n")
}

func newViperConfig() (*crcConfig.Config, *crcConfig.ViperStorage, error) {
	viper, err := crcConfig.NewViperStorage(constants.ConfigPath, constants.CrcEnvPrefix)
	if err != nil {
		return nil, nil, err
	}
	cfg := crcConfig.New(viper)
	cmdConfig.RegisterSettings(cfg)
	preflight.RegisterSettings(cfg)
	return cfg, viper, nil
}
