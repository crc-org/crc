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
	Short: "Get IP of the instance",
	Long:  "Get IP of the instance",
	Run: func(cmd *cobra.Command, args []string) {
		runIP(args)
	},
}

func runIP(arguments []string) {
	ipConfig := machine.IpConfig{
		Name:  constants.DefaultName,
		Debug: isDebugLog(),
	}

	result, err := machine.Ip(ipConfig)
	if err != nil {
		errors.ExitWithMessage(1, err.Error())
	}
	output.Out(result.IP)
}
