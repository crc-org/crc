package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sshCmd)
}

var sshCmd = &cobra.Command{
	Use:   "ssh [-- COMMAND...]",
	Short: "Open an SSH connection to the OpenShift cluster node",
	Long:  "Open an SSH connection to the OpenShift cluster node. Pass commands after -- to execute them remotely.",
	RunE: func(_ *cobra.Command, args []string) error {
		return runSSH(args)
	},
}

func runSSH(args []string) error {
	client := newMachine()
	if err := checkIfMachineMissing(client); err != nil {
		return err
	}

	connectionDetails, err := client.ConnectionDetails()
	if err != nil {
		return err
	}

	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("cannot find ssh binary: %w", err)
	}

	var sshKey string
	for _, key := range connectionDetails.SSHKeys {
		if _, err := os.Stat(key); err == nil {
			sshKey = key
			break
		}
	}
	if sshKey == "" {
		return fmt.Errorf("no SSH key found")
	}

	sshArgs := []string{
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		"-i", sshKey,
		"-p", strconv.Itoa(connectionDetails.SSHPort),
		fmt.Sprintf("%s@%s", connectionDetails.SSHUsername, connectionDetails.IP),
	}

	if len(args) > 0 {
		sshArgs = append(sshArgs, args...)
	}

	cmd := exec.Command(sshPath, sshArgs[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
