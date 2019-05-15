package cmd

import (
	"fmt"
	"github.com/code-ready/crc/cmd/crc/cmd/config"

	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	bundlePath      string
	startCmdFlagSet = initStartCmdFlagSet()
)

func runStart(arguments []string) {
	if err := constants.ValidateDriver(crcConfig.GetString(config.VMDriver.Name)); err != nil {
		errors.ExitWithMessage(1, err.Error())
	}

	preflight.StartPreflightChecks()

	startConfig := machine.StartConfig{
		Name:       constants.DefaultName,
		BundlePath: crcConfig.GetString(config.Bundle.Name),
		VMDriver:   crcConfig.GetString(config.VMDriver.Name),
		Memory:     constants.DefaultMemory,
		CPUs:       constants.DefaultCPUs,
		Debug:      false, // TODO: make this configurable
	}

	commandResult, err := machine.Start(startConfig)
	logging.InfoF(commandResult.Status)
	if err != nil {
		logging.ErrorF(err.Error())
	}
}

func initStartCmdFlagSet() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("start", pflag.ExitOnError)
	flagSet.StringP(config.Bundle.Name, "b", constants.DefaultBundle, "The system bundle used for deployment of the OpenShift cluster.")
	flagSet.StringP(config.VMDriver.Name, "d", constants.DefaultVMDriver, fmt.Sprintf("The driver to use for the CRC VM. Possible values: %v", constants.SupportedVMDrivers))

	return flagSet
}
