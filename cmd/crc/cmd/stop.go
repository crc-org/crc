package cmd

import (
	"github.com/spf13/cobra"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/output"

	"github.com/code-ready/crc/pkg/crc/machine"
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

	commandResult, err := machine.Stop(stopConfig)
	output.Out(commandResult.Success)
	if err != nil {
		output.Out(err.Error())
	}
}
