package cmd

import (
	"os"
	"text/template"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/output"
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
	UserShell string
	shell.ShellConfig
	OcDirPath string
	UsageHint string
}

func getOcShellConfig(ocPath string, forcedShell string) (*OcShellConfig, error) {
	userShell, err := shell.GetShell(forcedShell)
	if err != nil {
		return nil, errors.Newf("Error running the oc-env command: %s", err.Error())
	}

	cmdLine := "crc oc-env"

	shellCfg := &OcShellConfig{
		OcDirPath: ocPath,
		UserShell: userShell,
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
	Short: "Add the 'oc' binary to PATH",
	Long:  `Add the OpenShift client binary 'oc' to PATH`,
	// This is required to make sure root command Persistent PreRun not run.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {},
	Run: func(cmd *cobra.Command, args []string) {
		var shellCfg *OcShellConfig
		shellCfg, err := getOcShellConfig(constants.CrcBinDir, forceShell)
		if err != nil {
			errors.Exit(1)
		}

		output.Outln(shell.GetPathEnvString(shellCfg.UserShell, constants.CrcBinDir))
		output.Outln(shellCfg.UsageHint)
	},
}

func init() {
	rootCmd.AddCommand(ocEnvCmd)
	ocEnvCmd.Flags().StringVar(&forceShell, "shell", "", "Set the environment for the specified shell: [fish, cmd, powershell, tcsh, bash, zsh]. Default is auto-detect.")
}
