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

		proxyConfig, err := machine.GetProxyConfig(constants.DefaultName)
		if err != nil {
			errors.Exit(1)
		}
		output.Outln(shell.GetPathEnvString(userShell, constants.CrcOcBinDir))
		if proxyConfig.IsEnabled() {
			output.Outln(shell.GetEnvString(userShell, "HTTP_PROXY", proxyConfig.HttpProxy))
			output.Outln(shell.GetEnvString(userShell, "HTTPS_PROXY", proxyConfig.HttpsProxy))
			output.Outln(shell.GetEnvString(userShell, "NO_PROXY", proxyConfig.GetNoProxyString()))
		}
		output.Outln(shell.GenerateUsageHint(userShell, "crc oc-env"))
	},
}

func init() {
	rootCmd.AddCommand(ocEnvCmd)
	ocEnvCmd.Flags().StringVar(&forceShell, "shell", "", "Set the environment for the specified shell: [fish, cmd, powershell, tcsh, bash, zsh]. Default is auto-detect.")
}
