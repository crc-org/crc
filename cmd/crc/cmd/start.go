package cmd

import (
	"github.com/spf13/cobra"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/code-ready/crc/pkg/crc/logging"
)

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVarP(&bundlePath, "Bundle", "b", constants.DefaultBundle, "The system bundle used for deployment of the OpenShift cluster.")
	startCmd.Flags().StringVarP(&vmDriver, "VMDriver", "d", constants.DefaultDriver, "The hypervisor driver to use")
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
	bundlePath string
	vmDriver   string
)

func runStart(arguments []string) {

	// TODO: this should be a validation
	if vmDriver != "libvirt" {
		errors.ExitWithMessage(1, "Unsupported driver: %s", vmDriver)
	}

	preflight.StartPreflightChecks()

	startConfig := machine.StartConfig{
		Name:       constants.DefaultName,
		BundlePath: bundlePath,
		VMDriver:   vmDriver,
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
