package cmd

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/input"
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
	Short: "Stop cluster",
	Long:  "Stop cluster",
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
		errors.ExitWithMessage(1, err.Error())
	}
	output.Out(commandResult.Success)
}

func killVM(killConfig machine.PowerOffConfig) {
	commandResult, err := machine.PowerOff(killConfig)
	output.Out(commandResult.Success)
	if err != nil {
		errors.ExitWithMessage(1, err.Error())
	}
}
