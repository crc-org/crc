package cmd

import (
	"fmt"
	"github.com/spf13/cobra"

	"github.com/code-ready/crc/cmd/crc/cmd/config"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/code-ready/crc/pkg/crc/validation"
)

func init() {
	setupCmd.Flags().StringVarP(&vmDriver, config.VMDriver.Name, "d",
		machine.DefaultDriver.Driver, fmt.Sprintf("The driver to use for the OpenShift cluster. Possible values: %v", machine.SupportedDriverValues()))
	setupCmd.Flags().Bool(config.ExperimentalFeatures.Name, false, "Allow the use of experimental features")

	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up prerequisites for the OpenShift cluster",
	Long:  "Set up local virtualization and networking infrastructure for the OpenShift cluster",
	Run: func(cmd *cobra.Command, args []string) {
		runSetup(args)
	},
}

func runSetup(arguments []string) {
	if err := validateSetupFlags(); err != nil {
		errors.Exit(1)
	}

	if crcConfig.GetBool(config.ExperimentalFeatures.Name) {
		preflight.EnableExperimentalFeatures = true
	}
	preflight.SetupHost()
	var bundle string
	if !constants.BundleEmbedded() {
		bundle = " -b $bundlename"
	}
	output.Outf("Setup is complete, you can now run 'crc start%s' to start the OpenShift cluster\n", bundle)
}

var vmDriver string

func validateSetupFlags() error {
	if err := validation.ValidateDriver(vmDriver); err != nil {
		return err
	}
	return nil
}
