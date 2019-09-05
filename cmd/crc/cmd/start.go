package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/code-ready/crc/cmd/crc/cmd/config"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/input"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/code-ready/crc/pkg/crc/validation"
)

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().AddFlagSet(startCmdFlagSet)

	_ = crcConfig.BindFlagSet(startCmd.Flags())
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start cluster",
	Long:  "Start cluster",
	Run: func(cmd *cobra.Command, args []string) {
		runStart(args)
	},
}

var (
	startCmdFlagSet = initStartCmdFlagSet()
)

func runStart(arguments []string) {
	if err := validateStartFlags(); err != nil {
		errors.Exit(1)
	}

	preflight.StartPreflightChecks(crcConfig.GetString(config.VMDriver.Name))

	pullsecretFileContent, err := getPullSecretFileContent()
	if err != nil {
		errors.Exit(1)
	}

	startConfig := machine.StartConfig{
		Name:       constants.DefaultName,
		BundlePath: crcConfig.GetString(config.Bundle.Name),
		VMDriver:   crcConfig.GetString(config.VMDriver.Name),
		Memory:     crcConfig.GetInt(config.Memory.Name),
		CPUs:       crcConfig.GetInt(config.CPUs.Name),
		NameServer: crcConfig.GetString(config.NameServer.Name),
		PullSecret: pullsecretFileContent,
		Debug:      isDebugLog(),
	}

	commandResult, err := machine.Start(startConfig)
	if err != nil {
		errors.Exit(1)
	}
	if commandResult.Status == "Running" {
		output.Out("CodeReady Containers instance is running")
	} else {
		logging.Warnf("Unexpected status: %s", commandResult.Status)
	}
}

func initStartCmdFlagSet() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("start", pflag.ExitOnError)
	flagSet.StringP(config.Bundle.Name, "b", constants.DefaultBundlePath, "The system bundle used for deployment of the OpenShift cluster.")
	flagSet.StringP(config.VMDriver.Name, "d", machine.DefaultDriver.Driver, fmt.Sprintf("The driver to use for the CRC VM. Possible values: %v", machine.SupportedDriverValues()))
	flagSet.StringP(config.PullSecretFile.Name, "p", "", fmt.Sprintf("File path of Image pull secret for User (Download it from %s)", constants.DefaultPullSecretURL))
	flagSet.IntP(config.CPUs.Name, "c", constants.DefaultCPUs, "Number of CPU cores to allocate to the CRC VM")
	flagSet.IntP(config.Memory.Name, "m", constants.DefaultMemory, "MiB of Memory to allocate to the CRC VM")
	flagSet.StringP(config.NameServer.Name, "n", "", "Specify nameserver to use for the instance. (i.e. 8.8.8.8)")

	return flagSet
}

func isDebugLog() bool {
	return logging.LogLevel == "debug"
}

func validateStartFlags() error {
	if err := validation.ValidateDriver(crcConfig.GetString(config.VMDriver.Name)); err != nil {
		return err
	}
	if err := validation.ValidateMemory(crcConfig.GetInt(config.Memory.Name)); err != nil {
		return err
	}
	if err := validation.ValidateCPUs(crcConfig.GetInt(config.CPUs.Name)); err != nil {
		return err
	}
	if err := validation.ValidateBundle(crcConfig.GetString(config.Bundle.Name)); err != nil {
		return err
	}
	if crcConfig.GetString(config.NameServer.Name) != "" {
		if err := validation.ValidateIpAddress(crcConfig.GetString(config.NameServer.Name)); err != nil {
			return err
		}
	}
	return nil
}

func getPullSecretFileContent() (string, error) {
	var (
		pullsecret string
		err        error
	)

	// In case user doesn't provide a file in start command or in config then ask for it.
	if crcConfig.GetString(config.PullSecretFile.Name) == "" {
		pullsecret, err = input.PromptUserForSecret("Image pull secret", fmt.Sprintf("Copy it from %s", constants.DefaultPullSecretURL))
		// This is just to provide a new line after user enter the pull secret.
		fmt.Println()
		if err != nil {
			return "", errors.New(err.Error())
		}
	} else {
		// Read the file content
		data, err := ioutil.ReadFile(crcConfig.GetString(config.PullSecretFile.Name))
		if err != nil {
			return "", errors.New(err.Error())
		}
		pullsecret = string(data)
	}
	if err := validation.ImagePullSecret(pullsecret); err != nil {
		return "", errors.New(err.Error())
	}

	return pullsecret, nil
}
