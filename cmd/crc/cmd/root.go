package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	cmdBundle "github.com/code-ready/crc/cmd/crc/cmd/bundle"
	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	crcErr "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/code-ready/crc/pkg/crc/segment"
	"github.com/code-ready/crc/pkg/crc/telemetry"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/exec"
)

var rootCmd = &cobra.Command{
	Use:   commandName,
	Short: descriptionShort,
	Long:  descriptionLong,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return runPrerun(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		runRoot()
		_ = cmd.Help()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

var (
	globalForce   bool
	viper         *crcConfig.ViperStorage
	config        *crcConfig.Config
	segmentClient *segment.Client
)

func init() {
	if err := constants.EnsureBaseDirectoriesExist(); err != nil {
		logging.Fatal(err.Error())
	}
	var err error
	config, viper, err = newViperConfig()
	if err != nil {
		logging.Fatal(err.Error())
	}

	if err := setProxyDefaults(); err != nil {
		logging.Fatal(err.Error())
	}

	// Initiate segment client
	if segmentClient, err = segment.NewClient(config, httpTransport()); err != nil {
		logging.Fatal(err.Error())
	}

	// subcommands
	rootCmd.AddCommand(cmdConfig.GetConfigCmd(config))
	rootCmd.AddCommand(cmdBundle.GetBundleCmd(config))

	rootCmd.PersistentFlags().StringVar(&logging.LogLevel, "log-level", constants.DefaultLogLevel, "log level (e.g. \"debug | info | warn | error\")")
}

func runPrerun(cmd *cobra.Command) error {
	// Setting up logrus
	logFile := constants.LogFilePath
	if cmd == daemonCmd {
		logFile = constants.DaemonLogFilePath
	}
	logging.InitLogrus(logging.LogLevel, logFile)

	for _, str := range defaultVersion().lines() {
		logging.Debugf(str)
	}
	return nil
}

func runPostrun() {
	segmentClient.Close()
	logging.CloseLogging()
}

func runRoot() {
	fmt.Println("No command given")
}

const (
	defaultErrorExitCode    = 1
	preflightFailedExitCode = 2
)

func Execute() {
	attachMiddleware([]string{}, rootCmd)

	if err := rootCmd.ExecuteContext(telemetry.NewContext(context.Background())); err != nil {
		runPostrun()
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		var e exec.CodeExitError
		if errors.As(err, &e) {
			os.Exit(e.ExitStatus())
		} else {
			os.Exit(defaultErrorExitCode)
		}
	}
	runPostrun()
}

func checkIfMachineMissing(client machine.Client) error {
	exists, err := client.Exists()
	if err != nil {
		return err
	}
	if !exists {
		return crcErr.VMNotExist
	}
	return nil
}

func setProxyDefaults() error {
	httpProxy := config.Get(crcConfig.HTTPProxy).AsString()
	httpsProxy := config.Get(crcConfig.HTTPSProxy).AsString()
	noProxy := config.Get(crcConfig.NoProxy).AsString()
	proxyCAFile := config.Get(crcConfig.ProxyCAFile).AsString()

	proxyConfig, err := network.NewProxyDefaults(httpProxy, httpsProxy, noProxy, proxyCAFile)
	if err != nil {
		return err
	}

	if proxyConfig.IsEnabled() {
		var caFileForDisplay string
		if proxyCAFile != "" {
			caFileForDisplay = fmt.Sprintf(", proxyCAFile: %s", proxyCAFile)
		}
		logging.Debugf("HTTP-PROXY: %s, HTTPS-PROXY: %s, NO-PROXY: %s%s", proxyConfig.HTTPProxyForDisplay(),
			proxyConfig.HTTPSProxyForDisplay(), proxyConfig.GetNoProxyString(), caFileForDisplay)
		proxyConfig.ApplyToEnvironment()
	}
	return nil
}

func newViperConfig() (*crcConfig.Config, *crcConfig.ViperStorage, error) {
	viper, err := crcConfig.NewViperStorage(constants.ConfigPath, constants.CrcEnvPrefix)
	if err != nil {
		return nil, nil, err
	}
	cfg := crcConfig.New(viper)
	crcConfig.RegisterSettings(cfg)
	preflight.RegisterSettings(cfg)
	return cfg, viper, nil
}

func newMachine() machine.Client {
	return machine.NewSynchronizedMachine(machine.NewClient(constants.DefaultName, isDebugLog(), config))
}

func addForceFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&globalForce, "force", "f", false, "Forcefully perform this action")
}

func executeWithLogging(fullCmd string, input func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		logging.Debugf("Running '%s'", fullCmd)
		startTime := time.Now()
		err := input(cmd, args)
		if serr := segmentClient.UploadCmd(cmd.Context(), fullCmd, time.Since(startTime), err); serr != nil {
			logging.Debugf("Cannot send data to telemetry: %v", serr)
		}
		return err
	}
}

func attachMiddleware(names []string, cmd *cobra.Command) {
	if cmd.HasSubCommands() {
		for _, command := range cmd.Commands() {
			attachMiddleware(append(names, cmd.Name()), command)
		}
	} else if cmd.RunE != nil {
		fullCmd := strings.Join(append(names, cmd.Name()), " ")
		src := cmd.RunE
		cmd.RunE = executeWithLogging(fullCmd, src)
	}
}

func defaultTransport() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport)
	proxyConfig, err := network.NewProxyConfig()
	if err != nil || !proxyConfig.IsEnabled() {
		return transport
	}

	transport = transport.Clone()
	transport.Proxy = proxyConfig.ProxyFunc()
	return transport
}

func httpTransport() http.RoundTripper {
	if config.Get(crcConfig.ProxyCAFile).IsDefault {
		return defaultTransport()
	}
	caCert, err := ioutil.ReadFile(config.Get(crcConfig.ProxyCAFile).AsString())
	if err != nil {
		logging.Errorf("Cannot read proxy-ca-file, using default http transport: %v", err)
		return defaultTransport()
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	transport := defaultTransport()
	transport.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    caCertPool,
	}
	return transport
}
