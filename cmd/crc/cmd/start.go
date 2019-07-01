package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/code-ready/crc/cmd/crc/cmd/config"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/code-ready/crc/pkg/crc/validation"
)

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().AddFlagSet(startCmdFlagSet)

	crcConfig.BindFlagSet(startCmd.Flags())
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

	startConfig := machine.StartConfig{
		Name:       constants.DefaultName,
		BundlePath: crcConfig.GetString(config.Bundle.Name),
		VMDriver:   crcConfig.GetString(config.VMDriver.Name),
		Memory:     crcConfig.GetInt(config.Memory.Name),
		CPUs:       crcConfig.GetInt(config.CPUs.Name),
		Debug:      isDebugLog(),
	}

	commandResult, err := machine.Start(startConfig)
	if err != nil {
		errors.Exit(1)
	}
	if commandResult.Status == "Running" {
		output.Out("CodeReady Containers instance is running")
	} else {
		logging.WarnF("Unexpected status: %s", commandResult.Status)
	}
}

func initStartCmdFlagSet() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("start", pflag.ExitOnError)
	flagSet.StringP(config.Bundle.Name, "b", constants.DefaultBundle, "The system bundle used for deployment of the OpenShift cluster.")
	flagSet.StringP(config.VMDriver.Name, "d", machine.DefaultDriver.Driver, fmt.Sprintf("The driver to use for the CRC VM. Possible values: %v", machine.SupportedDriverValues()))
	flagSet.IntP(config.CPUs.Name, "c", constants.DefaultCPUs, "Number of CPU cores to allocate to the CRC VM")
	flagSet.IntP(config.Memory.Name, "m", constants.DefaultMemory, "MiB of Memory to allocate to the CRC VM")

	return flagSet
}

func isDebugLog() bool {
	if logging.LogLevel == "debug" {
		return true
	}
	return false
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
	return nil
}
