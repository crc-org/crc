package cmd

import (
	"os"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/code-ready/crc/pkg/os/shell"
	"github.com/spf13/cobra"
)

var podmanEnvCmd = &cobra.Command{
	Use:   "podman-env",
	Short: "Setup podman environment",
	Long:  `Setup environment for 'podman' binary to access podman on CRC VM`,
	Run: func(cmd *cobra.Command, args []string) {

		// See issue #961; Currently does not work on Windows in combination with the CRC vm.
		errors.ExitWithMessage(1, "Currently not supported.")

		userShell, err := shell.GetShell(forceShell)
		if err != nil {
			errors.ExitWithMessage(1, "Error running the podman-env command: %s", err.Error())
		}

		ipConfig := machine.IpConfig{
			Name:  constants.DefaultName,
			Debug: isDebugLog(),
		}

		exitIfMachineMissing(ipConfig.Name)

		result, err := machine.Ip(ipConfig)
		if err != nil {
			errors.Exit(1)
		}

		output.Outln(shell.GetPathEnvString(userShell, constants.CrcBinDir))
		output.Outln(shell.GetEnvString(userShell, "PODMAN_USER", constants.DefaultSSHUser))
		output.Outln(shell.GetEnvString(userShell, "PODMAN_HOST", result.IP))
		output.Outln(shell.GetEnvString(userShell, "PODMAN_IDENTITY_FILE", constants.GetPrivateKeyPath()))
		output.Outln(shell.GetEnvString(userShell, "PODMAN_IGNORE_HOSTS", "1"))
		output.Outln(shell.GenerateUsageHint(userShell, os.Args[0]+" podman-env"))
	},
}

func init() {
	rootCmd.AddCommand(podmanEnvCmd)
	podmanEnvCmd.Flags().StringVar(&forceShell, "shell", "", "Set the environment for the specified shell: [fish, cmd, powershell, tcsh, bash, zsh]. Default is auto-detect.")
}
