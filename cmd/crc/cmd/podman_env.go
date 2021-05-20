package cmd

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/os/shell"
	"github.com/spf13/cobra"
)

var (
	rootless bool
)

var podmanEnvCmd = &cobra.Command{
	Use:   "podman-env",
	Short: "Setup podman environment",
	Long:  `Setup environment for 'podman' executable to access podman on CRC VM`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPodmanEnv()
	},
}

func runPodmanEnv() error {
	userShell, err := shell.GetShell(forceShell)
	if err != nil {
		return fmt.Errorf("Error running the podman-env command: %s", err.Error())
	}

	client := newMachine()
	if err := checkIfMachineMissing(client); err != nil {
		return err
	}

	connectionDetails, err := client.ConnectionDetails()
	if err != nil {
		return err
	}

	socket := "/run/podman/podman.sock"
	if rootless {
		socket = "/run/user/1000/podman/podman.sock"
	}
	fmt.Println(shell.GetPathEnvString(userShell, constants.CrcOcBinDir))
	fmt.Println(shell.GetEnvString(userShell, "CONTAINER_SSHKEY", connectionDetails.SSHKeys[0]))
	fmt.Println(shell.GetEnvString(userShell, "CONTAINER_HOST",
		fmt.Sprintf("ssh://%s@%s:%d%s",
			connectionDetails.SSHUsername,
			connectionDetails.IP,
			connectionDetails.SSHPort,
			socket)))
	fmt.Println(shell.GenerateUsageHintWithComment(userShell, "crc podman-env"))
	return nil
}

func init() {
	podmanEnvCmd.Flags().BoolVar(&rootless, "rootless", false, "Use rootless podman in the virtual machine")
	podmanEnvCmd.Flags().StringVar(&forceShell, "shell", "", "Set the environment for the specified shell: [fish, cmd, powershell, tcsh, bash, zsh]. Default is auto-detect.")
	rootCmd.AddCommand(podmanEnvCmd)
}
