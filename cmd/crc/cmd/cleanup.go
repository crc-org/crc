package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/code-ready/crc/pkg/crc/exit"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/spf13/cobra"
)

func init() {
	addOutputFormatFlag(cleanupCmd)
	rootCmd.AddCommand(cleanupCmd)
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Undo config changes",
	Long:  "Undo all the configuration changes done by 'crc setup' command",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runCleanup(); err != nil {
			exit.WithMessage(1, err.Error())
		}
	},
}

func runCleanup() error {
	err := preflight.CleanUpHost()
	return render(&cleanupResult{
		Success: err == nil,
		Error:   errorMessage(err),
	}, os.Stdout, outputFormat)
}

type cleanupResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (s *cleanupResult) prettyPrintTo(writer io.Writer) error {
	if s.Error != "" {
		return errors.New(s.Error)
	}
	_, err := fmt.Fprintln(writer, "Cleanup finished")
	return err
}
