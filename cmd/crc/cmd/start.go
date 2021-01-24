package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/constants"
	crcErrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/input"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/code-ready/crc/pkg/crc/telemetry"
	"github.com/code-ready/crc/pkg/crc/validation"
	crcversion "github.com/code-ready/crc/pkg/crc/version"
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

	if err := preflight.StartPreflightChecks(config); err != nil {
		return nil, err
	}

	for _, key := range []string{cmdConfig.CPUs, cmdConfig.Memory, cmdConfig.DiskSize} {
		telemetry.SetContextProperty(ctx, key, config.Get(key).AsInt())
	}

	startConfig := machine.StartConfig{
		BundlePath: config.Get(cmdConfig.Bundle).AsString(),
		Memory:     config.Get(cmdConfig.Memory).AsInt(),
		DiskSize:   config.Get(cmdConfig.DiskSize).AsInt(),
		CPUs:       config.Get(cmdConfig.CPUs).AsInt(),
		NameServer: config.Get(cmdConfig.NameServer).AsString(),
		PullSecret: &cluster.PullSecret{
			Getter: getPullSecretFileContent,
		},
	}

	client := newMachine()
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
		return s.Error
	}
	if s.ClusterConfig == nil {
		return errors.New("either Error or ClusterConfig are needed")
	}

	_, err := fmt.Fprintln(writer, strings.Join([]string{
		"\n",
		"Started the OpenShift cluster",
		"",
		"To access the cluster, first set up your environment by following the instructions returned by executing 'crc oc-env'.",
		fmt.Sprintf("Then you can access your cluster by running 'oc login -u %s -p %s %s'.", s.ClusterConfig.DeveloperCredentials.Username, s.ClusterConfig.DeveloperCredentials.Password, s.ClusterConfig.URL),
		fmt.Sprintf("To login as a cluster admin, run 'oc login -u %s -p %s %s'.", s.ClusterConfig.AdminCredentials.Username, s.ClusterConfig.AdminCredentials.Password, s.ClusterConfig.URL),
		"",
		"You can also run 'crc console' and use the above credentials to access the OpenShift web console.",
		"The console will open in your default browser.",
	}, "\n"))
	if crcversion.IsOkdBuild() {
		fmt.Fprintln(writer, strings.Join([]string{
			"\n",
			"NOTE:",
			"This cluster was built from OKD - The Community Distribution of Kubernetes that powers Red Hat OpenShift.",
			"If you find an issue, please report it at https://github.com/openshift/okd"}, "\n"))
	}
	return err
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

func getPullSecretFileContent() (string, error) {
	var (
		pullsecret string
		err        error
	)

	// If crc is built from an OKD bundle, then use the fake pull secret in contants.
	if crcversion.IsOkdBuild() {
		pullsecret = constants.OkdPullSecret
		return pullsecret, nil
	}
	// In case user doesn't provide a file in start command or in config then ask for it.
	if config.Get(cmdConfig.PullSecretFile).AsString() == "" {
		pullsecret, err = input.PromptUserForSecret("Image pull secret", fmt.Sprintf("Copy it from %s", constants.CrcLandingPageURL))
		// This is just to provide a new line after user enter the pull secret.
		fmt.Println()
		if err != nil {
			return "", errors.New(err.Error())
		}
	} else {
		// Read the file content
		data, err := ioutil.ReadFile(config.Get(cmdConfig.PullSecretFile).AsString())
		if err != nil {
			return "", errors.New(err.Error())
		}
		pullsecret = string(data)
	}
	if err := validation.ImagePullSecret(pullsecret); err != nil {
		return "", errors.New(err.Error())
	}

	return pullsecret, nil
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
