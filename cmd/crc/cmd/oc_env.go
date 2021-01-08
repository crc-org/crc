package cmd

import (
	"fmt"
	"os"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/os/shell"
	"github.com/spf13/cobra"
)

var (
	forceShell string
)

var ocEnvCmd = &cobra.Command{
	Use:   "oc-env",
	Short: "Add the 'oc' executable to PATH",
	Long:  `Add the OpenShift client executable 'oc' to PATH`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runOcEnv(args)
	},
}

func runOcEnv(args []string) error {
	userShell, err := shell.GetShell(forceShell)
	if err != nil {
		return fmt.Errorf("Error running the oc-env command: %s", err.Error())
	}

	client := newMachine()
	consoleResult, err := client.GetConsoleURL()
	if err != nil {
		return err
	}
	proxyConfig := consoleResult.ClusterConfig.ProxyConfig
	fmt.Println(shell.GetPathEnvString(userShell, constants.CrcOcBinDir))
	if proxyConfig.IsEnabled() {
		fmt.Println(shell.GetEnvString(userShell, "HTTP_PROXY", proxyConfig.HTTPProxy))
		fmt.Println(shell.GetEnvString(userShell, "HTTPS_PROXY", proxyConfig.HTTPSProxy))
		fmt.Println(shell.GetEnvString(userShell, "NO_PROXY", proxyConfig.GetNoProxyString()))
	}
	fmt.Println(shell.GenerateUsageHint(userShell, "crc oc-env"))
	return nil
}

func init() {
	if os.Getenv("CRC_PACKAGE") == "" {
		rootCmd.AddCommand(ocEnvCmd)
	}
	ocEnvCmd.Flags().StringVar(&forceShell, "shell", "", "Set the environment for the specified shell: [fish, cmd, powershell, tcsh, bash, zsh]. Default is auto-detect.")
}
