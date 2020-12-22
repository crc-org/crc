package config

import (
	"errors"
	"fmt"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/output"
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
			switch {
			case v.Invalid:
				return fmt.Errorf("Configuration property '%s' does not exist", key)
			case v.IsDefault:
				return fmt.Errorf("Configuration property '%s' is not set. Default value is '%s'", key, v.AsString())
			default:
				output.Outln(key, ":", v.AsString())
			}
			return nil
		},
	}
}
