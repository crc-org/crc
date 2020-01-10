package cmd

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/code-ready/crc/pkg/os/shell"
	"github.com/spf13/cobra"
)

var (
	forceShell string
)

var ocEnvCmd = &cobra.Command{
	Use:   "oc-env",
	Short: "Add the 'oc' binary to PATH",
	Long:  `Add the OpenShift client binary 'oc' to PATH`,
	Run: func(cmd *cobra.Command, args []string) {
		userShell, err := shell.GetShell(forceShell)
		if err != nil {
			errors.ExitWithMessage(1, "Error running the oc-env command: %s", err.Error())
		}

		clusterConfig, err := machine.GetClusterConfig(constants.DefaultName)
		if err != nil {
			errors.Exit(1)
		}
		output.Outln(shell.GetPathEnvString(userShell, constants.CrcBinDir))
		if clusterConfig.ProxyConfig.IsEnabled() {
			output.Outln(shell.GetEnvString(userShell, "HTTP_PROXY", clusterConfig.ProxyConfig.HttpProxy))
			output.Outln(shell.GetEnvString(userShell, "HTTPS_PROXY", clusterConfig.ProxyConfig.HttpsProxy))
			output.Outln(shell.GetEnvString(userShell, "NO_PROXY", clusterConfig.ProxyConfig.GetNoProxyString()))
		}
		output.Outf("echo \"You can now login to the OpenShift cluster with 'oc login -p developer -u developer %s'\"\n", clusterConfig.ClusterAPI)
		output.Outln(shell.GenerateUsageHint(userShell, "crc oc-env"))
	},
}

func init() {
	rootCmd.AddCommand(ocEnvCmd)
	ocEnvCmd.Flags().StringVar(&forceShell, "shell", "", "Set the environment for the specified shell: [fish, cmd, powershell, tcsh, bash, zsh]. Default is auto-detect.")
}
