package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/input"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/output"
)

func init() {
	deleteCmd.Flags().BoolVarP(&clearCache, "clear-cache", "", false,
		fmt.Sprintf("Clear the OpenShift cluster cache at: %s", constants.MachineCacheDir))
	rootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the OpenShift cluster",
	Long:  "Delete the OpenShift cluster",
	Run: func(cmd *cobra.Command, args []string) {
		runDelete(args)
	},
}

var clearCache bool

func runDelete(arguments []string) {
	deleteConfig := machine.DeleteConfig{
		Name: constants.DefaultName,
	}

	exitIfMachineMissing(deleteConfig.Name)

	if clearCache {
		deleteCache()
	}
	yes := input.PromptUserForYesOrNo("Do you want to delete the OpenShift cluster", globalForce)
	if yes {
		_, err := machine.Delete(deleteConfig)
		if err != nil {
			errors.Exit(1)
		}
		output.Outln("The OpenShift cluster deleted")
	}
}

func deleteCache() {
	yes := input.PromptUserForYesOrNo("Do you want to delete the OpenShift cluster cache", globalForce)
	if yes {
		os.RemoveAll(constants.MachineCacheDir)
	}
}
