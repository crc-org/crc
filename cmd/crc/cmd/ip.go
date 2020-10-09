package cmd

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/exit"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(ipCmd)
}

var ipCmd = &cobra.Command{
	Use:   "ip",
	Short: "Get IP address of the running OpenShift cluster",
	Long:  "Get IP address of the running OpenShift cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runIP(args); err != nil {
			exit.WithMessage(1, err.Error())
		}
	},
}

func runIP(arguments []string) error {
	ipConfig := machine.IPConfig{
		Name:  constants.DefaultName,
		Debug: isDebugLog(),
	}

	client := machine.NewClient()
	if err := checkIfMachineMissing(client); err != nil {
		return err
	}

	ip, err := client.IP(ipConfig)
	if err != nil {
		return err
	}
	output.Outln(ip)
	return nil
}
