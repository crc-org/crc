package cmd

import (
	"github.com/spf13/cobra"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/output"

	"github.com/code-ready/crc/pkg/crc/machine"
)

func init() {
	rootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete cluster",
	Long:  "Delete cluster",
	Run: func(cmd *cobra.Command, args []string) {
		runDelete(args)
	},
}

func runDelete(arguments []string) {

	deleteConfig := machine.DeleteConfig{
		Name: constants.DefaultName,
	}

	commandResult, err := machine.Delete(deleteConfig)
	output.Out(commandResult.Success)
	if err != nil {
		output.Out(err.Error())
	}
}
