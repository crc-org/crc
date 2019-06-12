package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/os/shell"
	"github.com/spf13/cobra"
)

const (
	ocEnvTmpl = `{{ .Prefix }}PATH{{ .Delimiter }}{{ .OcDirPath }}{{ .PathSuffix }}{{ .UsageHint }}`
)

var (
	forceShell string
)

type OcShellConfig struct {
	shell.ShellConfig
	OcDirPath string
	UsageHint string
}

func getOcShellConfig(ocPath string, forcedShell string) (*OcShellConfig, error) {
	userShell, err := shell.GetShell(forcedShell)
	if err != nil {
		return nil, err
	}

	cmdLine := "crc oc-env"

	shellCfg := &OcShellConfig{
		OcDirPath: ocPath,
	}

	shellCfg.UsageHint = shell.GenerateUsageHint(userShell, cmdLine)
	shellCfg.Prefix, shellCfg.Delimiter, shellCfg.Suffix, shellCfg.PathSuffix = shell.GetPrefixSuffixDelimiterForSet(userShell)

	return shellCfg, nil
}

func executeOcTemplateStdout(shellCfg *OcShellConfig) error {
	tmpl := template.Must(template.New("envConfig").Parse(ocEnvTmpl))
	return tmpl.Execute(os.Stdout, shellCfg)
}

var ocEnvCmd = &cobra.Command{
	Use:   "oc-env",
	Short: "Sets the path of the 'oc' binary.",
	Long:  `Sets the path of OpenShift client binary 'oc'.`,
	// This is required to make sure root command Persistent PreRun not run.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {},
	Run: func(cmd *cobra.Command, args []string) {
		var shellCfg *OcShellConfig
		shellCfg, err := getOcShellConfig(constants.OcCacheDir, forceShell)
		if err != nil {
			errors.ExitWithMessage(1, fmt.Sprintf("Error running the oc-env command: %s", err.Error()))
		}
		executeOcTemplateStdout(shellCfg)
	},
}

func init() {
	rootCmd.AddCommand(ocEnvCmd)
	ocEnvCmd.Flags().StringVar(&forceShell, "shell", "", "Force setting the environment for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh]. Default is auto-detect.")
}
