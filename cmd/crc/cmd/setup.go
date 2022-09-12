package cmd

import (
	"fmt"
	"io"
	"os"

	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	crcErrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/input"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/code-ready/crc/pkg/crc/validation"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/exec"
)

var (
	checkOnly bool
)

func init() {
	setupCmd.Flags().Bool(crcConfig.ExperimentalFeatures, false, "Allow the use of experimental features")
	setupCmd.Flags().StringP(crcConfig.Bundle, "b", constants.GetDefaultBundlePath(crcConfig.GetPreset(config)), "Bundle to use for instance")
	setupCmd.Flags().BoolVar(&checkOnly, "check-only", false, "Only run the preflight checks, don't try to fix any misconfiguration")
	addOutputFormatFlag(setupCmd)
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up prerequisites for using CRC",
	Long:  "Set up local virtualization and networking infrastructure for using CRC",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindFlagSet(cmd.Flags()); err != nil {
			return err
		}
		return runSetup(args)
	},
}

func runSetup(arguments []string) error {
	if config.Get(crcConfig.ConsentTelemetry).AsString() == "" {
		fmt.Println("CRC is constantly improving and we would like to know more about usage (more details at https://developers.redhat.com/article/tool-data-collection)")
		fmt.Println("Your preference can be changed manually if desired using 'crc config set consent-telemetry <yes/no>'")
		if input.PromptUserForYesOrNo("Would you like to contribute anonymous usage statistics", false) {
			if _, err := config.Set(crcConfig.ConsentTelemetry, "yes"); err != nil {
				return err
			}
			fmt.Printf("Thanks for helping us! You can disable telemetry with the command 'crc config set %s no'.\n", crcConfig.ConsentTelemetry)
		} else {
			if _, err := config.Set(crcConfig.ConsentTelemetry, "no"); err != nil {
				return err
			}
			fmt.Printf("No worry, you can still enable telemetry manually with the command 'crc config set %s yes'.\n", crcConfig.ConsentTelemetry)
		}
	}

	if !config.Get(crcConfig.Bundle).IsDefault {
		if err := validation.ValidateBundlePath(config.Get(crcConfig.Bundle).AsString(), crcConfig.GetPreset(config)); err != nil {
			return err
		}
	}

	err := preflight.SetupHost(config, checkOnly)
	if err != nil && checkOnly {
		err = exec.CodeExitError{
			Err:  err,
			Code: preflightFailedExitCode,
		}
	}

	return render(&setupResult{
		Success: err == nil,
		Error:   crcErrors.ToSerializableError(err),
	}, os.Stdout, outputFormat)
}

type setupResult struct {
	Success bool                         `json:"success"`
	Error   *crcErrors.SerializableError `json:"error,omitempty"`
}

func (s *setupResult) prettyPrintTo(writer io.Writer) error {
	if s.Error != nil {
		return s.Error
	}
	_, err := fmt.Fprintln(writer, "Your system is correctly setup for using CRC. "+
		"Use 'crc start' to start the instance")
	return err
}
