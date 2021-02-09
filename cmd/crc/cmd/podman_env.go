package cmd

import (
	"errors"
	"fmt"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/os/shell"
	"github.com/spf13/cobra"
)

var podmanEnvCmd = &cobra.Command{
	Use:   "podman-env",
	Short: "Setup podman environment",
	Long:  `Setup environment for 'podman' executable to access podman on CRC VM`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// See issue #961; Currently does not work on Windows in combination with the CRC vm.
		return errors.New("currently not supported")
	},
}

func RunPodmanEnv(args []string) error {
	userShell, err := shell.GetShell(forceShell)
	if err != nil {
		return fmt.Errorf("Error running the podman-env command: %s", err.Error())
	}

	client := newMachine()
	if err := checkIfMachineMissing(client); err != nil {
		return err
	}

	ip, err := client.IP()
	if err != nil {
		return err
	}

	fmt.Println(shell.GetPathEnvString(userShell, constants.CrcBinDir))
	fmt.Println(shell.GetEnvString(userShell, "PODMAN_USER", constants.DefaultSSHUser))
	fmt.Println(shell.GetEnvString(userShell, "PODMAN_HOST", ip))
	fmt.Println(shell.GetEnvString(userShell, "PODMAN_IDENTITY_FILE", constants.GetPrivateKeyPath()))
	fmt.Println(shell.GetEnvString(userShell, "PODMAN_IGNORE_HOSTS", "1"))
	fmt.Println(shell.GenerateUsageHintWithComment(userShell, "crc podman-env"))
	return nil
}

func init() {
	rootCmd.AddCommand(podmanEnvCmd)
	podmanEnvCmd.Flags().StringVar(&forceShell, "shell", "", "Set the environment for the specified shell: [fish, cmd, powershell, tcsh, bash, zsh]. Default is auto-detect.")
}
