package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

const jsonFormat = "json"

var (
	outputFormat string
)

func addOutputFormatFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format. One of: json")
}

type prettyPrintable interface {
	prettyPrintTo(writer io.Writer) error
}

func render(obj prettyPrintable, writer io.Writer, outputFormat string) error {
	switch outputFormat {
	case jsonFormat:
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(obj)
	case "":
		return obj.prettyPrintTo(writer)
	default:
		return fmt.Errorf("invalid format: %s", outputFormat)
	}
}
