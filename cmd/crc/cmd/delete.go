package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	crcErrors "github.com/crc-org/crc/v2/pkg/crc/errors"
	"github.com/crc-org/crc/v2/pkg/crc/input"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine"
	crcos "github.com/crc-org/crc/v2/pkg/os"
	"github.com/spf13/cobra"
)

var clearCache bool

func init() {
	deleteCmd.Flags().BoolVarP(&clearCache, "clear-cache", "", false,
		fmt.Sprintf("Clear the instance cache at: %s", constants.MachineCacheDir))
	addOutputFormatFlag(deleteCmd)
	addForceFlag(deleteCmd)
	rootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the instance",
	Long:  "Delete the instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDelete(os.Stdout, newMachine(), clearCache, constants.MachineCacheDir, outputFormat != jsonFormat, globalForce, outputFormat)
	},
}

func deleteMachine(client machine.Client, clearCache bool, cacheDir string, interactive, force bool) (bool, error) {
	if clearCache {
		if !interactive && !force {
			return false, errors.New("non-interactive deletion requires --force")
		}
		yes := input.PromptUserForYesOrNo("Do you want to delete the instance cache", force)
		if yes {
			_ = os.RemoveAll(cacheDir)
			// also delete the crc-*.log files
			if err := crcos.RemoveFileGlob(filepath.Join(constants.CrcBaseDir, "crc-*.log")); err != nil {
				logging.Debug("Failed to find log files: ", err)
			}
		}
	}

	if err := checkIfMachineMissing(client); err != nil {
		return false, err
	}

	if !interactive && !force {
		return false, errors.New("non-interactive deletion requires --force")
	}

	yes := input.PromptUserForYesOrNo("Do you want to delete the instance",
		force)
	if yes {
		defer logging.BackupLogFile()
		return true, client.Delete()
	}
	return false, nil
}

func runDelete(writer io.Writer, client machine.Client, clearCache bool, cacheDir string, interactive, force bool, outputFormat string) error {
	machineDeleted, err := deleteMachine(client, clearCache, cacheDir, interactive, force)
	return render(&deleteResult{
		Success:        err == nil,
		Error:          crcErrors.ToSerializableError(err),
		machineDeleted: machineDeleted,
	}, writer, outputFormat)
}

type deleteResult struct {
	Success        bool                         `json:"success"`
	Error          *crcErrors.SerializableError `json:"error,omitempty"`
	machineDeleted bool
}

func (s *deleteResult) prettyPrintTo(writer io.Writer) error {
	if s.Error != nil {
		return s.Error
	}
	if s.machineDeleted {
		if _, err := fmt.Fprintln(writer, "Deleted the instance"); err != nil {
			return err
		}
	}
	return nil
}
