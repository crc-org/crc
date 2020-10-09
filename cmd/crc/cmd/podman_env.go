package cmd

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/exit"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/code-ready/crc/pkg/os/shell"
	"github.com/spf13/cobra"
)

var podmanEnvCmd = &cobra.Command{
	Use:   "podman-env",
	Short: "Setup podman environment",
	Long:  `Setup environment for 'podman' binary to access podman on CRC VM`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runPodmanEnv(args); err != nil {
			exit.WithMessage(1, err.Error())
		}
	},
}

func runPodmanEnv(args []string) error {
	// See issue #961; Currently does not work on Windows in combination with the CRC vm.
	exit.WithMessage(1, "Currently not supported.")

	userShell, err := shell.GetShell(forceShell)
	if err != nil {
		return fmt.Errorf("Error running the podman-env command: %s", err.Error())
	}

	client := newMachine()
	if err := checkIfMachineMissing(client); err != nil {
		return err
	}

	ip, err := client.IP(constants.DefaultName)
	if err != nil {
		return err
	}

	output.Outln(shell.GetPathEnvString(userShell, constants.CrcBinDir))
	output.Outln(shell.GetEnvString(userShell, "PODMAN_USER", constants.DefaultSSHUser))
	output.Outln(shell.GetEnvString(userShell, "PODMAN_HOST", ip))
	output.Outln(shell.GetEnvString(userShell, "PODMAN_IDENTITY_FILE", constants.GetPrivateKeyPath()))
	output.Outln(shell.GetEnvString(userShell, "PODMAN_IGNORE_HOSTS", "1"))
	output.Outln(shell.GenerateUsageHint(userShell, "crc podman-env"))
	return nil
}

func init() {
	rootCmd.AddCommand(podmanEnvCmd)
	podmanEnvCmd.Flags().StringVar(&forceShell, "shell", "", "Set the environment for the specified shell: [fish, cmd, powershell, tcsh, bash, zsh]. Default is auto-detect.")
}
