package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/exit"
	"github.com/code-ready/crc/pkg/crc/input"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/machine/libmachine/state"
	"github.com/spf13/cobra"
)

func init() {
	addOutputFormatFlag(stopCmd)
	rootCmd.AddCommand(stopCmd)
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the OpenShift cluster",
	Long:  "Stop the OpenShift cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if renderErr := runStop(os.Stdout, machine.NewClient(), outputFormat != jsonFormat, globalForce, outputFormat); renderErr != nil {
			exit.WithMessage(1, renderErr.Error())
		}
	},
}

func stopMachine(client machine.Client, interactive, force bool) (bool, error) {
	if err := checkIfMachineMissing(client); err != nil {
		return false, err
	}

	result, err := client.Stop(machine.StopConfig{
		Name:  constants.DefaultName,
		Debug: isDebugLog(),
	})
	if err != nil {
		if !interactive && !force {
			return false, err
		}
		// Here we are checking the VM state and if it is still running then
		// Ask user to forcefully power off it.
		if result.State == state.Running {
			// Most of the time force kill don't work and libvirt throw
			// Device or resource busy error. To make sure we give some
			// graceful time to cluster before kill it.
			yes := input.PromptUserForYesOrNo("Do you want to force power off", force)
			if yes {
				_, err := client.PowerOff(machine.PowerOffConfig{
					Name: constants.DefaultName,
				})
				return true, err
			}
		}
		return false, err
	}
	return false, nil
}

func runStop(writer io.Writer, client machine.Client, interactive, force bool, outputFormat string) error {
	forced, err := stopMachine(client, interactive, force)
	return render(&stopResult{
		Success: err == nil,
		Forced:  forced,
		Error:   errorMessage(err),
	}, writer, outputFormat)
}

type stopResult struct {
	Success bool   `json:"success"`
	Forced  bool   `json:"forced"`
	Error   string `json:"error,omitempty"`
}

func (s *stopResult) prettyPrintTo(writer io.Writer) error {
	if s.Error != "" {
		return errors.New(s.Error)
	}
	if s.Forced {
		_, err := fmt.Fprintln(writer, "Forcibly stopped the OpenShift cluster")
		return err
	}
	_, err := fmt.Fprintln(writer, "Stopped the OpenShift cluster")
	return err
}
