package cmd

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/input"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/code-ready/machine/libmachine/state"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stopCmd)
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the OpenShift cluster",
	Long:  "Stop the OpenShift cluster",
	Run: func(cmd *cobra.Command, args []string) {
		runStop(args)
	},
}

func runStop(arguments []string) {
	stopConfig := machine.StopConfig{
		Name:  constants.DefaultName,
		Debug: isDebugLog(),
	}

	killConfig := machine.PowerOffConfig{
		Name: constants.DefaultName,
	}

	exitIfMachineMissing(stopConfig.Name)

	output.Outln("Stopping the OpenShift cluster, this may take a few minutes...")
	commandResult, err := machine.Stop(stopConfig)
	if err != nil {
		// Here we are checking the VM state and if it is still running then
		// Ask user to forcefully power off it.
		if commandResult.State == state.Running {
			// Most of the time force kill don't work and libvirt throw
			// Device or resource busy error. To make sure we give some
			// graceful time to cluster before kill it.
			yes := input.PromptUserForYesOrNo("Do you want to force power off", globalForce)
			if yes {
				killVM(killConfig)
				errors.Exit(0)
			}
		}
		errors.Exit(1)
	}
	if commandResult.Success {
		output.Outln("Stopped the OpenShift cluster")
	} else {
		/* If we did not get an error, the status should be true */
		logging.Warnf("Unexpected status of the OpenShift cluster: %v", commandResult.Success)
	}
}

func killVM(killConfig machine.PowerOffConfig) {
	_, err := machine.PowerOff(killConfig)
	if err != nil {
		errors.Exit(1)
	}
	output.Outln("Forcibly stopped the OpenShift cluster")
}
