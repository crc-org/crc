package cmd

import (
	"fmt"
	"io"
	"os"

	crcErrors "github.com/code-ready/crc/pkg/crc/errors"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCleanup()
	},
}

func runCleanup() error {
	err := preflight.CleanUpHost()
	return render(&cleanupResult{
		Success: err == nil,
		Error:   crcErrors.ToSerializableError(err),
	}, os.Stdout, outputFormat)
}

type cleanupResult struct {
	Success bool                         `json:"success"`
	Error   *crcErrors.SerializableError `json:"error,omitempty"`
}

func (s *cleanupResult) prettyPrintTo(writer io.Writer) error {
	if s.Error != nil {
		return s.Error
	}
	_, err := fmt.Fprintln(writer, "Cleanup finished")
	return err
}
