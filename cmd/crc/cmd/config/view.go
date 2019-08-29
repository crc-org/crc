package config

import (
	"bytes"
	"fmt"
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/spf13/cobra"
	"io"
	"os"
	"sort"
	"text/template"
)

const (
	DefaultConfigViewFormat = "- {{.ConfigKey | printf \"%-38s\"}}: {{.ConfigValue}}"
)

var configViewFormat string

type configViewTemplate struct {
	ConfigKey   string
	ConfigValue interface{}
}

func init() {
	ConfigCmd.AddCommand(configViewCmd)
	configViewCmd.Flags().StringVar(&configViewFormat, "format", DefaultConfigViewFormat,
		`Go template format to apply to the configuration file. For more information about Go templates, see: https://golang.org/pkg/text/template/`)
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Display complete crc configurations.",
	Long: `Retrieves full list of crc configurations. Some of the configuration properties are equivalent
to the options that you set when you run the 'crc start' command.`,
	Run: func(cmd *cobra.Command, args []string) {
		tmpl, err := determineTemplate(configViewFormat)
		if err != nil {
			logging.Fatal(err)
		}
		if err := runConfigView(config.ChangedConfigs(), tmpl, os.Stdout); err != nil {
			logging.Fatal(err)
		}
	},
}

func determineTemplate(tempFormat string) (*template.Template, error) {
	tmpl, err := template.New("view").Parse(tempFormat)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func runConfigView(cfg map[string]interface{}, tmpl *template.Template, writer io.Writer) error {
	var lines []string
	for k, v := range cfg {
		viewTmplt := configViewTemplate{k, v}
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
