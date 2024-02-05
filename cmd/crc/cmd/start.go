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

	"github.com/Masterminds/semver/v3"
	"github.com/crc-org/crc/v2/pkg/crc/cluster"
	crcConfig "github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/daemonclient"
	crcErrors "github.com/crc-org/crc/v2/pkg/crc/errors"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine/bundle"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	"github.com/crc-org/crc/v2/pkg/crc/network"
	"github.com/crc-org/crc/v2/pkg/crc/preflight"
	"github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/crc/validation"
	crcversion "github.com/crc-org/crc/v2/pkg/crc/version"
	crcos "github.com/crc-org/crc/v2/pkg/os"
	"github.com/crc-org/crc/v2/pkg/os/shell"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	rootCmd.AddCommand(startCmd)
	addOutputFormatFlag(startCmd)

	flagSet := pflag.NewFlagSet("start", pflag.ExitOnError)
	flagSet.StringP(crcConfig.Bundle, "b", constants.GetDefaultBundlePath(crcConfig.GetPreset(config)), crcConfig.BundleHelpMsg(config))
	flagSet.StringP(crcConfig.PullSecretFile, "p", "", fmt.Sprintf("File path of image pull secret (download from %s)", constants.CrcLandingPageURL))
	flagSet.IntP(crcConfig.CPUs, "c", constants.GetDefaultCPUs(crcConfig.GetPreset(config)), "Number of CPU cores to allocate to the instance")
	flagSet.IntP(crcConfig.Memory, "m", constants.GetDefaultMemory(crcConfig.GetPreset(config)), "MiB of memory to allocate to the instance")
	flagSet.UintP(crcConfig.DiskSize, "d", constants.DefaultDiskSize, "Total size in GiB of the disk used by the instance")
	flagSet.StringP(crcConfig.NameServer, "n", "", "IPv4 address of nameserver to use for the instance")
	flagSet.Bool(crcConfig.DisableUpdateCheck, false, "Don't check for update")

	startCmd.Flags().AddFlagSet(flagSet)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the instance",
	Long:  "Start the instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindFlagSet(cmd.Flags()); err != nil {
			return err
		}
		return renderStartResult(runStart(cmd.Context()))
	},
}

func runStart(ctx context.Context) (*types.StartResult, error) {
	if err := validateStartFlags(); err != nil {
		return nil, err
	}

	if err := checkIfNewVersionAvailable(config.Get(crcConfig.DisableUpdateCheck).AsBool()); err != nil {
		logging.Debugf("Unable to find out if a new version is available: %v", err)
	}

	startConfig := types.StartConfig{
		BundlePath:        config.Get(crcConfig.Bundle).AsString(),
		Memory:            config.Get(crcConfig.Memory).AsInt(),
		DiskSize:          config.Get(crcConfig.DiskSize).AsInt(),
		CPUs:              config.Get(crcConfig.CPUs).AsInt(),
		NameServer:        config.Get(crcConfig.NameServer).AsString(),
		PullSecret:        cluster.NewInteractivePullSecretLoader(config),
		KubeAdminPassword: config.Get(crcConfig.KubeAdminPassword).AsString(),
		Preset:            crcConfig.GetPreset(config),
		IngressHTTPPort:   config.Get(crcConfig.IngressHTTPPort).AsUInt(),
		IngressHTTPSPort:  config.Get(crcConfig.IngressHTTPSPort).AsUInt(),
		EnableSharedDirs:  config.Get(crcConfig.EnableSharedDirs).AsBool(),

		EmergencyLogin: config.Get(crcConfig.EmergencyLogin).AsBool(),

		PersistentVolumeSize: config.Get(crcConfig.PersistentVolumeSize).AsInt(),

		EnableBundleQuayFallback: config.Get(crcConfig.EnableBundleQuayFallback).AsBool(),
	}

	if startConfig.Preset == preset.Podman {
		logging.Warn(preset.PodmanDeprecatedWarning)
	}

	client := newMachine()
	isRunning, _ := client.IsRunning()

	if !isRunning {
		if err := checkDaemonStarted(); err != nil {
			return nil, err
		}

		if err := preflight.StartPreflightChecks(config); err != nil {
			return nil, crcos.CodeExitError{
				Err:  err,
				Code: preflightFailedExitCode,
			}
		}
	}

	if runtime.GOOS == "windows" {
		username, err := crcos.GetCurrentUsername()
		if err != nil {
			return nil, err
		}

		// config SharedDirPassword ('shared-dir-password') only exists in windows
		startConfig.SharedDirPassword = config.Get(crcConfig.SharedDirPassword).AsString()
		startConfig.SharedDirUsername = username
	}

	return client.Start(ctx, startConfig)
}

func renderStartResult(result *types.StartResult, err error) error {
	return render(&startResult{
		Success:       err == nil,
		Error:         crcErrors.ToSerializableError(err),
		ClusterConfig: toClusterConfig(result),
	}, os.Stdout, outputFormat)
}

func toClusterConfig(result *types.StartResult) *clusterConfig {
	if result == nil {
		return nil
	}
	return &clusterConfig{
		ClusterType:   result.ClusterConfig.ClusterType,
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
	ClusterType          preset.Preset `json:"clusterType"`
	ClusterCACert        string        `json:"cacert"`
	WebConsoleURL        string        `json:"webConsoleUrl"`
	URL                  string        `json:"url"`
	AdminCredentials     credentials   `json:"adminCredentials"`
	DeveloperCredentials credentials   `json:"developerCredentials"`
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
		return errors.New("either Error or ClusterConfig is needed")
	}

	return writeTemplatedMessage(writer, s)
}

func validateStartFlags() error {
	if err := validation.ValidateMemory(config.Get(crcConfig.Memory).AsInt(), crcConfig.GetPreset(config)); err != nil {
		return err
	}
	if err := validation.ValidateCPUs(config.Get(crcConfig.CPUs).AsInt(), crcConfig.GetPreset(config)); err != nil {
		return err
	}
	if err := validation.ValidateDiskSize(config.Get(crcConfig.DiskSize).AsInt()); err != nil {
		return err
	}
	if err := validation.ValidateBundle(config.Get(crcConfig.Bundle).AsString(), crcConfig.GetPreset(config)); err != nil {
		return err
	}
	if config.Get(crcConfig.NameServer).AsString() != "" {
		if err := validation.ValidateIPAddress(config.Get(crcConfig.NameServer).AsString()); err != nil {
			return err
		}
	}
	return nil
}

func checkIfNewVersionAvailable(noUpdateCheck bool) error {
	if noUpdateCheck {
		return nil
	}
	isNewVersionAvailable, newVersion, link, err := newVersionAvailable()
	if err != nil {
		return err
	}
	if isNewVersionAvailable {
		logging.Warnf("A new version (%s) has been published on %s", newVersion, link)
		return nil
	}
	logging.Debugf("No new version available. The latest version is %s", newVersion)
	return nil
}

func newVersionAvailable() (bool, string, string, error) {
	release, err := bundle.FetchLatestReleaseInfo()
	if err != nil {
		return false, "", "", err
	}
	currentVersion, err := semver.NewVersion(crcversion.GetCRCVersion())
	if err != nil {
		return false, "", "", err
	}
	if release.Version.CrcVersion == nil {
		return false, "", "", errors.New("empty version")
	}
	return release.Version.CrcVersion.GreaterThan(currentVersion), release.Version.CrcVersion.String(), downloadLink(release), nil
}

func downloadLink(release *bundle.ReleaseInfo) string {
	if link, ok := release.Links[runtime.GOOS]; ok {
		return link
	}
	return constants.CrcLandingPageURL
}

const (
	startTemplateForOpenshift = `Started the OpenShift cluster.

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
  {{ .CommandLinePrefix }} oc login -u {{ .ClusterConfig.DeveloperCredentials.Username }} {{ .ClusterConfig.URL }}
`
	startTemplateForPodman = `podman runtime is now running.

Use the 'podman' command line interface:
  {{ .CommandLinePrefix }} {{ .EvalCommandLine }}
  {{ .CommandLinePrefix }} {{ .PodmanRemote }} COMMAND
`
	startTemplateForOKD = `NOTE:
This cluster was built from OKD - The Community Distribution of Kubernetes that powers Red Hat OpenShift.
If you find an issue, please report it at https://github.com/openshift/okd
`
	startTemplateForMicroShift = `Started the MicroShift cluster.

Use the 'oc' command line interface:
  {{ .CommandLinePrefix }} {{ .EvalCommandLine }}
  {{ .CommandLinePrefix }} oc COMMAND
`
)

type templateVariables struct {
	ClusterConfig       *clusterConfig
	EvalCommandLine     string
	CommandLinePrefix   string
	PodmanRemote        string
	FallbackPortWarning string
	KubeConfigPath      string
}

func writeTemplatedMessage(writer io.Writer, s *startResult) error {
	if s.ClusterConfig.ClusterType == preset.OpenShift || s.ClusterConfig.ClusterType == preset.OKD {
		return writeOpenShiftTemplatedMessage(writer, s)
	}
	if s.ClusterConfig.ClusterType == preset.Microshift {
		return writeMicroShiftTemplatedMessage(writer)
	}

	return writePodmanTemplatedMessage(writer)
}

func writeOpenShiftTemplatedMessage(writer io.Writer, s *startResult) error {
	tmpl := startTemplateForOpenshift
	if s.ClusterConfig.ClusterType == preset.OKD {
		tmpl = fmt.Sprintf("%s\n\n%s", startTemplateForOpenshift, startTemplateForOKD)
	}
	fallbackPortWarning := portFallbackWarning()
	if fallbackPortWarning != "" {
		tmpl = fmt.Sprintf("%s\n%s", tmpl, fallbackPortWarning)
	}
	parsed, err := template.New("template").Parse(tmpl)
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

func writePodmanTemplatedMessage(writer io.Writer) error {
	parsed, err := template.New("template").Parse(startTemplateForPodman)
	if err != nil {
		return err
	}
	userShell, err := shell.GetShell("")
	if err != nil {
		userShell = ""
	}

	return parsed.Execute(writer, &templateVariables{
		EvalCommandLine:   shell.GenerateUsageHint(userShell, "crc podman-env"),
		CommandLinePrefix: commandLinePrefix(userShell),
		PodmanRemote:      constants.PodmanRemoteExecutableName,
	})
}

func writeMicroShiftTemplatedMessage(writer io.Writer) error {
	parsed, err := template.New("template").Parse(startTemplateForMicroShift)
	if err != nil {
		return err
	}
	userShell, err := shell.GetShell("")
	if err != nil {
		userShell = ""
	}

	return parsed.Execute(writer, &templateVariables{
		EvalCommandLine:   shell.GenerateUsageHint(userShell, "crc oc-env"),
		CommandLinePrefix: commandLinePrefix(userShell),
		KubeConfigPath:    constants.KubeconfigFilePath,
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

func checkDaemonStarted() error {
	if crcConfig.GetNetworkMode(config) == network.SystemNetworkingMode {
		return nil
	}
	v, err := daemonclient.GetVersionFromDaemonAPI()
	if err != nil {
		return err
	}
	return daemonclient.CheckVersionMismatch(v)
}

func portFallbackWarning() string {
	var fallbackPortWarning string
	httpScheme := "http"
	httpsScheme := "https"
	fallbackPortWarningTmpl := `
Warning: Port %d is used for OpenShift %s routes instead of the default
You have to add this port to %s URLs when accessing OpenShift application,
such as %s://myapp.apps-crc.testing:%d
`
	ingressHTTPPort := config.Get(crcConfig.IngressHTTPPort).AsUInt()
	ingressHTTPSPort := config.Get(crcConfig.IngressHTTPSPort).AsUInt()

	if ingressHTTPSPort != constants.OpenShiftIngressHTTPSPort {
		fallbackPortWarning += fmt.Sprintf(fallbackPortWarningTmpl,
			ingressHTTPSPort, strings.ToUpper(httpsScheme), strings.ToUpper(httpsScheme), httpsScheme, ingressHTTPSPort)
	}
	if ingressHTTPPort != constants.OpenShiftIngressHTTPPort {
		fallbackPortWarning += fmt.Sprintf(fallbackPortWarningTmpl,
			ingressHTTPPort, strings.ToUpper(httpScheme), strings.ToUpper(httpScheme), httpScheme, ingressHTTPPort)
	}
	return fallbackPortWarning
}
