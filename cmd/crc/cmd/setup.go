package cmd

import (
	"fmt"
	"github.com/spf13/cobra"

	"github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/code-ready/crc/pkg/crc/validation"
)

func init() {
	setupCmd.Flags().StringVarP(&vmDriver, config.VMDriver.Name, "d",
		machine.DefaultDriver.Driver, fmt.Sprintf("The driver to use for the CRC VM. Possible values: %v", machine.SupportedDriverValues()))
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "setup hypervisor",
	Long:  "setup hypervisor to run the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		runSetup(args)
	},
}

func runSetup(arguments []string) {
	if err := validateSetupFlags(); err != nil {
		errors.Exit(1)
	}
	preflight.SetupHost(vmDriver)
}

var vmDriver string

func validateSetupFlags() error {
	if err := validation.ValidateDriver(vmDriver); err != nil {
		return err
	}
	return nil
}
