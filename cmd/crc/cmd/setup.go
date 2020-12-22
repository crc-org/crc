package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/spf13/cobra"
)

func init() {
	setupCmd.Flags().Bool(cmdConfig.ExperimentalFeatures, false, "Allow the use of experimental features")
	addOutputFormatFlag(setupCmd)
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up prerequisites for the OpenShift cluster",
	Long:  "Set up local virtualization and networking infrastructure for the OpenShift cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindFlagSet(cmd.Flags()); err != nil {
			return err
		}
		return runSetup(args)
	},
}

func runSetup(arguments []string) error {
	err := preflight.SetupHost(config)
	return render(&setupResult{
		Success: err == nil,
		Error:   errorMessage(err),
	}, os.Stdout, outputFormat)
}

type setupResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (s *setupResult) prettyPrintTo(writer io.Writer) error {
	if s.Error != "" {
		return errors.New(s.Error)
	}
	_, err := fmt.Fprintf(writer, "Setup is complete, you can now run 'crc start%s' to start the OpenShift cluster\n", extraArguments())
	return err
}

func extraArguments() string {
	var bundle string
	if !constants.BundleEmbedded() {
		bundle = " -b $bundlename"
	}
	return bundle
}
