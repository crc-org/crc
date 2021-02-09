package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"text/template"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/constants"
	crcErrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/code-ready/crc/pkg/crc/telemetry"
	"github.com/code-ready/crc/pkg/crc/validation"
	crcversion "github.com/code-ready/crc/pkg/crc/version"
	"github.com/code-ready/crc/pkg/os/shell"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	rootCmd.AddCommand(startCmd)
	addOutputFormatFlag(startCmd)

	flagSet := pflag.NewFlagSet("start", pflag.ExitOnError)
	flagSet.StringP(cmdConfig.Bundle, "b", constants.DefaultBundlePath, "The system bundle used for deployment of the OpenShift cluster")
	flagSet.StringP(cmdConfig.PullSecretFile, "p", "", fmt.Sprintf("File path of image pull secret (download from %s)", constants.CrcLandingPageURL))
	flagSet.IntP(cmdConfig.CPUs, "c", constants.DefaultCPUs, "Number of CPU cores to allocate to the OpenShift cluster")
	flagSet.IntP(cmdConfig.Memory, "m", constants.DefaultMemory, "MiB of memory to allocate to the OpenShift cluster")
	flagSet.UintP(cmdConfig.DiskSize, "d", constants.DefaultDiskSize, "Total size in GiB of the disk used by the OpenShift cluster")
	flagSet.StringP(cmdConfig.NameServer, "n", "", "IPv4 address of nameserver to use for the OpenShift cluster")
	flagSet.Bool(cmdConfig.DisableUpdateCheck, false, "Don't check for update")

	startCmd.Flags().AddFlagSet(flagSet)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the OpenShift cluster",
	Long:  "Start the OpenShift cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindFlagSet(cmd.Flags()); err != nil {
			return err
		}
		if err := renderStartResult(runStart(cmd.Context())); err != nil {
			return err
		}
		return nil
	},
}

func runStart(ctx context.Context) (*machine.StartResult, error) {
	if err := validateStartFlags(); err != nil {
		return nil, err
	}

	checkIfNewVersionAvailable(config.Get(cmdConfig.DisableUpdateCheck).AsBool())

	telemetry.SetContextProperty(ctx, cmdConfig.CPUs, config.Get(cmdConfig.CPUs).AsInt())
	telemetry.SetContextProperty(ctx, cmdConfig.Memory, uint64(config.Get(cmdConfig.Memory).AsInt())*1024*1024)
	telemetry.SetContextProperty(ctx, cmdConfig.DiskSize, uint64(config.Get(cmdConfig.DiskSize).AsInt())*1024*1024*1024)

	startConfig := machine.StartConfig{
		BundlePath: config.Get(cmdConfig.Bundle).AsString(),
		Memory:     config.Get(cmdConfig.Memory).AsInt(),
		DiskSize:   config.Get(cmdConfig.DiskSize).AsInt(),
		CPUs:       config.Get(cmdConfig.CPUs).AsInt(),
		NameServer: config.Get(cmdConfig.NameServer).AsString(),
		PullSecret: cluster.NewInteractivePullSecretLoader(config),
	}

	client := newMachine()
	isRunning, _ := client.IsRunning()

	if !isRunning {
		if err := preflight.StartPreflightChecks(config); err != nil {
			return nil, err
		}
	}

	return client.Start(startConfig)
}

func renderStartResult(result *machine.StartResult, err error) error {
	return render(&startResult{
		Success:       err == nil,
		Error:         crcErrors.ToSerializableError(err),
		ClusterConfig: toClusterConfig(result),
	}, os.Stdout, outputFormat)
}

func toClusterConfig(result *machine.StartResult) *clusterConfig {
	if result == nil {
		return nil
	}
	return &clusterConfig{
		ClusterCACert: result.ClusterConfig.ClusterCACert,
		WebConsoleURL: result.ClusterConfig.WebConsoleURL,
		URL:           result.ClusterConfig.ClusterAPI,
		AdminCredentials: credentials{
			Username: "kubeadmin",
			Password: result.ClusterConfig.KubeAdminPass,
		},
		DeveloperCredentials: credentials{
			Username: "developer",
			Password: "developer",
		},
	}
}

type clusterConfig struct {
	ClusterCACert        string      `json:"cacert"`
	WebConsoleURL        string      `json:"webConsoleUrl"`
	URL                  string      `json:"url"`
	AdminCredentials     credentials `json:"adminCredentials"`
	DeveloperCredentials credentials `json:"developerCredentials"`
}

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type startResult struct {
	Success       bool                         `json:"success"`
	Error         *crcErrors.SerializableError `json:"error,omitempty"`
	ClusterConfig *clusterConfig               `json:"clusterConfig,omitempty"`
}

func (s *startResult) prettyPrintTo(writer io.Writer) error {
	if s.Error != nil {
		var e *crcErrors.PreflightError
		if errors.As(s.Error, &e) {
			logging.Warn("Preflight checks failed during `crc start`, please try to run `crc setup` first in case you haven't done so yet")
		}
		return s.Error
	}
	if s.ClusterConfig == nil {
		return errors.New("either Error or ClusterConfig are needed")
	}

	if err := writeTemplatedMessage(writer, s); err != nil {
		return err
	}
	if crcversion.IsOkdBuild() {
		_, err := fmt.Fprintln(writer, strings.Join([]string{
			"",
			"NOTE:",
			"This cluster was built from OKD - The Community Distribution of Kubernetes that powers Red Hat OpenShift.",
			"If you find an issue, please report it at https://github.com/openshift/okd"}, "\n"))
		return err
	}
	return nil
}

func isDebugLog() bool {
	return logging.LogLevel == "debug"
}

func validateStartFlags() error {
	if err := validation.ValidateMemory(config.Get(cmdConfig.Memory).AsInt()); err != nil {
		return err
	}
	if err := validation.ValidateCPUs(config.Get(cmdConfig.CPUs).AsInt()); err != nil {
		return err
	}
	if err := validation.ValidateDiskSize(config.Get(cmdConfig.DiskSize).AsInt()); err != nil {
		return err
	}
	if err := validation.ValidateBundle(config.Get(cmdConfig.Bundle).AsString()); err != nil {
		return err
	}
	if config.Get(cmdConfig.NameServer).AsString() != "" {
		if err := validation.ValidateIPAddress(config.Get(cmdConfig.NameServer).AsString()); err != nil {
			return err
		}
	}
	return nil
}

func checkIfNewVersionAvailable(noUpdateCheck bool) {
	if noUpdateCheck {
		return
	}
	isNewVersionAvailable, newVersion, err := crcversion.NewVersionAvailable()
	if err != nil {
		logging.Debugf("Unable to find out if a new version is available: %v", err)
		return
	}
	if isNewVersionAvailable {
		logging.Warnf("A new version (%s) has been published on %s", newVersion, constants.CrcLandingPageURL)
		return
	}
	logging.Debugf("No new version available. The latest version is %s", newVersion)
}

const startTemplate = `Started the OpenShift cluster.

The server is accessible via web console at:
  {{ .ClusterConfig.WebConsoleURL }}

Log in as administrator:
  Username: {{ .ClusterConfig.AdminCredentials.Username }}
  Password: {{ .ClusterConfig.AdminCredentials.Password }}

Log in as user:
  Username: {{ .ClusterConfig.DeveloperCredentials.Username }}
  Password: {{ .ClusterConfig.DeveloperCredentials.Password }}

Use the 'oc' command line interface:
  {{ .CommandLinePrefix }} {{ .EvalCommandLine }}
  {{ .CommandLinePrefix }} oc login {{ .ClusterConfig.URL }}
`

type templateVariables struct {
	ClusterConfig     *clusterConfig
	EvalCommandLine   string
	CommandLinePrefix string
}

func writeTemplatedMessage(writer io.Writer, s *startResult) error {
	parsed, err := template.New("template").Parse(startTemplate)
	if err != nil {
		return err
	}

	userShell, err := shell.GetShell("")
	if err != nil {
		userShell = ""
	}
	return parsed.Execute(writer, &templateVariables{
		ClusterConfig:     s.ClusterConfig,
		EvalCommandLine:   shell.GenerateUsageHint(userShell, "crc oc-env"),
		CommandLinePrefix: commandLinePrefix(userShell),
	})
}

func commandLinePrefix(shell string) string {
	if runtime.GOOS == "windows" {
		if shell == "powershell" {
			return "PS>"
		}
		return ">"
	}
	return "$"
}
