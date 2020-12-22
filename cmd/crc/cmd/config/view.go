package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"text/template"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/spf13/cobra"
)

const (
	DefaultConfigViewFormat = "- {{.ConfigKey | printf \"%-38s\"}}: {{.ConfigValue}}"
)

var configViewFormat string

type configViewTemplate struct {
	ConfigKey   string
	ConfigValue interface{}
}

func configViewCmd(config config.Storage) *cobra.Command {
	configViewCmd := &cobra.Command{
		Use:   "view",
		Short: "Display all assigned crc configuration properties",
		Long:  `Displays all assigned crc configuration properties and their values.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tmpl, err := determineTemplate(configViewFormat)
			if err != nil {
				return err
			}
			return runConfigView(config.AllConfigs(), tmpl, os.Stdout)
		},
	}
	configViewCmd.Flags().StringVar(&configViewFormat, "format", DefaultConfigViewFormat,
		`Go template format to apply to the configuration file. For more information about Go templates, see: https://golang.org/pkg/text/template/`)
	return configViewCmd
}

func determineTemplate(tempFormat string) (*template.Template, error) {
	tmpl, err := template.New("view").Parse(tempFormat)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func runConfigView(cfg map[string]config.SettingValue, tmpl *template.Template, writer io.Writer) error {
	var lines []string
	for k, v := range cfg {
		if v.IsDefault {
			continue
		}
		viewTmplt := configViewTemplate{k, v.AsString()}
		var buffer bytes.Buffer
		if err := tmpl.Execute(&buffer, viewTmplt); err != nil {
			return err
		}
		lines = append(lines, buffer.String())
	}
	sort.Strings(lines)

	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}

	return nil
}
