package config

import (
	"errors"
	"fmt"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/telemetry"
	"github.com/spf13/cobra"
)

func configGetCmd(config config.Storage) *cobra.Command {
	return &cobra.Command{
		Use:   "get CONFIG-KEY",
		Short: "Get a crc configuration property",
		Long:  `Gets a crc configuration property.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("Please provide a configuration property to get")
			}
			key := args[0]

			v := config.Get(key)
			if v.Invalid {
				return fmt.Errorf("Configuration property '%s' does not exist", key)
			}

			telemetry.SetConfigurationKey(cmd.Context(), args[0])

			if v.IsDefault {
				return fmt.Errorf("Configuration property '%s' is not set. Default value is '%s'", key, v.AsString())

			}
			fmt.Println(key, ":", v.AsString())
			return nil
		},
	}
}
