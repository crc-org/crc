package config

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/spf13/cobra"
)

func isPreflightKey(key string) bool {
	return strings.HasPrefix(key, "skip-")
}

// less is used to sort the config keys. We want to sort first the regular keys, and
// then the keys related to preflight starting with a skip- prefix.
func less(lhsKey, rhsKey string) bool {
	if isPreflightKey(lhsKey) {
		if isPreflightKey(rhsKey) {
			// ignore skip prefix
			return lhsKey[4:] < rhsKey[4:]
		}
		// lhs is preflight, rhs is not preflight
		return false
	}

	if isPreflightKey(rhsKey) {
		// lhs is not preflight, rhs is preflight
		return true
	}

	// lhs is not preflight, rhs is not preflight
	return lhsKey < rhsKey
}

func configurableFields(config *config.Config) string {
	settings := config.AllSettings()
	sort.Slice(settings, func(i, j int) bool {
		return less(settings[i].Name, settings[j].Name)
	})

	var buf bytes.Buffer
	writer := tabwriter.NewWriter(&buf, 0, 8, 1, ' ', tabwriter.TabIndent)
	for _, cfg := range settings {
		fmt.Fprintf(writer, "* %s\t%s\n", cfg.Name, cfg.Help)
	}
	writer.Flush()

	return buf.String()
}

func GetConfigCmd(config *config.Config) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config SUBCOMMAND [flags]",
		Short: "Modify crc configuration",
		Long: `Modifies crc configuration properties.
Properties: ` + "\n\n" + configurableFields(config),
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
	configCmd.AddCommand(configGetCmd(config))
	configCmd.AddCommand(configSetCmd(config))
	configCmd.AddCommand(configUnsetCmd(config))
	configCmd.AddCommand(configViewCmd(config))
	return configCmd
}
