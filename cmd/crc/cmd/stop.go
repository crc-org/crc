package cmd

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/exit"
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
		if err := runStop(args); err != nil {
			exit.WithMessage(1, err.Error())
		}
	},
}

func runStop(arguments []string) error {
	stopConfig := machine.StopConfig{
		Name:  constants.DefaultName,
		Debug: isDebugLog(),
	}

	killConfig := machine.PowerOffConfig{
		Name: constants.DefaultName,
	}

	client := machine.NewClient()
	if err := checkIfMachineMissing(client, stopConfig.Name); err != nil {
		return err
	}

	output.Outln("Stopping the OpenShift cluster, this may take a few minutes...")
	commandResult, err := client.Stop(stopConfig)
	if err != nil {
		// Here we are checking the VM state and if it is still running then
		// Ask user to forcefully power off it.
		if commandResult.State == state.Running {
			// Most of the time force kill don't work and libvirt throw
			// Device or resource busy error. To make sure we give some
			// graceful time to cluster before kill it.
			yes := input.PromptUserForYesOrNo("Do you want to force power off", globalForce)
			if yes {
				_, err := client.PowerOff(killConfig)
				if err != nil {
					return err
				}
				output.Outln("Forcibly stopped the OpenShift cluster")
				return nil
			}
		}
		return err
	}
	if commandResult.Success {
		output.Outln("Stopped the OpenShift cluster")
	} else {
		/* If we did not get an error, the status should be true */
		logging.Warnf("Unexpected status of the OpenShift cluster: %v", commandResult.Success)
	}
	return nil
}
