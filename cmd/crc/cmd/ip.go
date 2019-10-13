package cmd

import (
	"github.com/spf13/cobra"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/output"
)

func init() {
	rootCmd.AddCommand(ipCmd)
}

var ipCmd = &cobra.Command{
	Use:   "ip",
	Short: "Get IP address of the running OpenShift cluster",
	Long:  "Get IP address of the running OpenShift cluster",
	Run: func(cmd *cobra.Command, args []string) {
		runIP(args)
	},
}

func runIP(arguments []string) {
	ipConfig := machine.IpConfig{
		Name:  constants.DefaultName,
		Debug: isDebugLog(),
	}

	exitIfMachineMissing(ipConfig.Name)

	result, err := machine.Ip(ipConfig)
	if err != nil {
		errors.Exit(1)
	}
	output.Outln(result.IP)
}
