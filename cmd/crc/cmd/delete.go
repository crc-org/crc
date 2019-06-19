package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/input"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/output"
)

func init() {
	deleteCmd.Flags().BoolVarP(&clearCache, "clear-cache", "", false,
		fmt.Sprintf("Clear the cache directory: %s", constants.MachineCacheDir))
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

var clearCache bool

func runDelete(arguments []string) {
	deleteConfig := machine.DeleteConfig{
		Name: constants.DefaultName,
	}
	if clearCache {
		deleteCache()
	}

	commandResult, err := machine.Delete(deleteConfig)
	output.Out(commandResult.Success)
	if err != nil {
		output.Out(err.Error())
	}
}

func deleteCache() {
	yes := input.PromptUserForYesOrNo("Do you want to delete cache", globalForce)
	if yes {
		os.RemoveAll(constants.MachineCacheDir)
	}
}
