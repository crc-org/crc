package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/os/shell"
	"github.com/spf13/cobra"
)

var (
	root bool
)

var podmanEnvCmd = &cobra.Command{
	Use:   "podman-env",
	Short: "Setup podman environment",
	Long:  `Setup environment for 'podman' executable to access podman on CRC VM`,
	RunE: func(_ *cobra.Command, _ []string) error {
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

	socket := constants.RootlessPodmanSocket
	if root {
		socket = constants.RootfulPodmanSocket
	}

	// Warning about podman-remote binary not being shipped and this command will be removed in future releases
	// (printed to stderr so it doesn't break eval)
	fmt.Fprintln(os.Stderr, "NOTE: The podman-remote binary is no longer shipped with crc.")
	fmt.Fprintln(os.Stderr, "To use podman with crc, install podman-remote from:")
	fmt.Fprintln(os.Stderr, "https://podman.io/docs/installation")
	fmt.Fprintln(os.Stderr, "In future releases, the podman-env command will be removed from crc.")
	fmt.Fprintln(os.Stderr, "Please create an issue at https://github.com/crc-org/crc/issues with details")
	fmt.Fprintln(os.Stderr, "if you think this command is still useful for your workflow.")
	fmt.Fprintln(os.Stderr)

	fmt.Println(shell.GetEnvString(userShell, "CONTAINER_SSHKEY", connectionDetails.SSHKeys[0]))
	fmt.Println(shell.GetEnvString(userShell, "CONTAINER_HOST",
		fmt.Sprintf("ssh://%s@%s:%d%s",
			connectionDetails.SSHUsername,
			connectionDetails.IP,
			connectionDetails.SSHPort,
			socket)))
	// Todo: This need to fixed by using named pipe for windows
	// https://docs.docker.com/desktop/faqs/#how-do-i-connect-to-the-remote-docker-engine-api
	if runtime.GOOS != "windows" {
		fmt.Println(shell.GetEnvString(userShell, "DOCKER_HOST", fmt.Sprintf("unix://%s", constants.GetHostDockerSocketPath())))
	} else {
		fmt.Println(shell.GetEnvString(userShell, "DOCKER_HOST", "npipe:////./pipe/crc-podman"))
	}
	if root {
		fmt.Println(shell.GenerateUsageHintWithComment(userShell, "crc podman-env --root"))
	} else {
		fmt.Println(shell.GenerateUsageHintWithComment(userShell, "crc podman-env"))
	}

	return nil
}

func init() {
	podmanEnvCmd.Flags().BoolVar(&root, "root", false, "Use root podman in the virtual machine")
	podmanEnvCmd.Flags().StringVar(&forceShell, "shell", "", "Set the environment for the specified shell: [fish, cmd, powershell, tcsh, bash, zsh]. Default is auto-detect.")
	rootCmd.AddCommand(podmanEnvCmd)
}
