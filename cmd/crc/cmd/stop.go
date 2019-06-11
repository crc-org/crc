package cmd

import (
	"fmt"
	"github.com/code-ready/machine/libmachine/state"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/spf13/cobra"

	"github.com/code-ready/crc/pkg/crc/machine"
)

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().BoolVarP(&force, "force", "f", false, "Stop the VM forcefully")
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop cluster",
	Long:  "Stop cluster",
	Run: func(cmd *cobra.Command, args []string) {
		runStop(args)
	},
}

var (
	force bool
)

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
			if force {
				killVM(killConfig)
				errors.Exit(0)
			}
			var userInput string
			output.OutF("Do you want to force power off [y/n]: ")
			fmt.Scan(&userInput)
			if strings.ToLower(userInput) == "y" {
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
