package cmd

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/exit"
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
		if err := runOcEnv(args); err != nil {
			exit.WithMessage(1, err.Error())
		}
	},
}

func runOcEnv(args []string) error {
	userShell, err := shell.GetShell(forceShell)
	if err != nil {
		return fmt.Errorf("Error running the oc-env command: %s", err.Error())
	}

	consoleResult, err := machine.GetConsoleURL(machine.ConsoleConfig{
		Name: constants.DefaultName,
	})
	if err != nil {
		return err
	}
	proxyConfig := consoleResult.ClusterConfig.ProxyConfig
	output.Outln(shell.GetPathEnvString(userShell, constants.CrcOcBinDir))
	if proxyConfig.IsEnabled() {
		output.Outln(shell.GetEnvString(userShell, "HTTP_PROXY", proxyConfig.HTTPProxy))
		output.Outln(shell.GetEnvString(userShell, "HTTPS_PROXY", proxyConfig.HTTPSProxy))
		output.Outln(shell.GetEnvString(userShell, "NO_PROXY", proxyConfig.GetNoProxyString()))
	}
	output.Outln(shell.GenerateUsageHint(userShell, "crc oc-env"))
	return nil
}

func init() {
	rootCmd.AddCommand(ocEnvCmd)
	ocEnvCmd.Flags().StringVar(&forceShell, "shell", "", "Set the environment for the specified shell: [fish, cmd, powershell, tcsh, bash, zsh]. Default is auto-detect.")
}
