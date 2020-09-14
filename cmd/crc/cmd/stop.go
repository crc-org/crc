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
		if err := runStop(); err != nil {
			exit.WithMessage(1, err.Error())
		}
	},
}

func runStop() error {
	client := machine.NewClient()
	if err := checkIfMachineMissing(client, constants.DefaultName); err != nil {
		return err
	}

	result, err := client.Stop(machine.StopConfig{
		Name:  constants.DefaultName,
		Debug: isDebugLog(),
	})
	if err != nil {
		// Here we are checking the VM state and if it is still running then
		// Ask user to forcefully power off it.
		if result.State == state.Running {
			// Most of the time force kill don't work and libvirt throw
			// Device or resource busy error. To make sure we give some
			// graceful time to cluster before kill it.
			yes := input.PromptUserForYesOrNo("Do you want to force power off", globalForce)
			if yes {
				_, err := client.PowerOff(machine.PowerOffConfig{
					Name: constants.DefaultName,
				})
				if err != nil {
					return err
				}
				output.Outln("Forcibly stopped the OpenShift cluster")
				return nil
			}
		}
		return err
	}
	if result.Success {
		output.Outln("Stopped the OpenShift cluster")
	} else {
		/* If we did not get an error, the status should be true */
		logging.Warnf("Unexpected status of the OpenShift cluster: %v", result.Success)
	}
	return nil
}
