package cmd

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
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
	// This is required to make sure root command Persistent PreRun not run.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {},
	Run: func(cmd *cobra.Command, args []string) {
		userShell, err := shell.GetShell(forceShell)
		if err != nil {
			errors.ExitWithMessage(1, "Error running the oc-env command: %s", err.Error())
		}

		output.Outln(shell.GetPathEnvString(userShell, constants.CrcBinDir))
		output.Outln(shell.GenerateUsageHint(userShell, "crc oc-env"))
	},
}

func init() {
	rootCmd.AddCommand(ocEnvCmd)
	ocEnvCmd.Flags().StringVar(&forceShell, "shell", "", "Set the environment for the specified shell: [fish, cmd, powershell, tcsh, bash, zsh]. Default is auto-detect.")
}
