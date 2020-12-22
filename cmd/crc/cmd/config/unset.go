package config

import (
	"errors"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/spf13/cobra"
)

func configUnsetCmd(config config.Storage) *cobra.Command {
	return &cobra.Command{
		Use:   "unset CONFIG-KEY",
		Short: "Unset a crc configuration property",
		Long:  `Unsets a crc configuration property.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("Please provide a configuration property to unset")
			}
			unsetMessage, err := config.Unset(args[0])
			if err != nil {
				return err
			}
			if unsetMessage != "" {
				output.Outln(unsetMessage)
			}
			return nil
		},
	}
}
