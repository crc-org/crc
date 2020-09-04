package config

import (
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/exit"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/spf13/cobra"
)

func configSetCmd(config config.Storage) *cobra.Command {
	return &cobra.Command{
		Use:   "set CONFIG-KEY VALUE",
		Short: "Set a crc configuration property",
		Long:  `Sets a crc configuration property.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				exit.WithMessage(1, "Please provide a configuration property and its value as in 'crc config set KEY VALUE'")
			}
			setMessage, err := config.Set(args[0], args[1])
			if err != nil {
				exit.WithMessage(1, err.Error())
			}

			if setMessage != "" {
				output.Outln(setMessage)
			}
		},
	}
}
