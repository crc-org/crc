package config

import (
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/exit"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/spf13/cobra"
)

func configUnsetCmd(config config.Storage) *cobra.Command {
	return &cobra.Command{
		Use:   "unset CONFIG-KEY",
		Short: "Unset a crc configuration property",
		Long:  `Unsets a crc configuration property.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				exit.WithMessage(1, "Please provide a configuration property to unset")
			}
			unsetMessage, err := config.Unset(args[0])
			if err != nil {
				exit.WithMessage(1, err.Error())
			}
			if unsetMessage != "" {
				output.Outln(unsetMessage)
			}
		},
	}
}
