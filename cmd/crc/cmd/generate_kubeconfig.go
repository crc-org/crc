package cmd

import (
	"fmt"
	"os"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(generateKubeconfigCmd)
}

var generateKubeconfigCmd = &cobra.Command{
	Use:   "generate-kubeconfig",
	Short: "Generate a kubeconfig file for the CRC instance",
	Long:  "Output the kubeconfig file for the running CRC instance to stdout",
	RunE: func(_ *cobra.Command, _ []string) error {
		return runGenerateKubeconfig()
	},
}

func runGenerateKubeconfig() error {
	client := newMachine()
	if err := checkIfMachineMissing(client); err != nil {
		return err
	}

	data, err := os.ReadFile(constants.KubeconfigFilePath)
	if err != nil {
		return fmt.Errorf("Error reading kubeconfig: %v", err)
	}
	fmt.Print(string(data))
	return nil
}
