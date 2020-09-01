package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/code-ready/crc/cmd/crc/cmd/config"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/exit"
	"github.com/code-ready/crc/pkg/crc/input"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/code-ready/crc/pkg/crc/validation"
	crcversion "github.com/code-ready/crc/pkg/crc/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	rootCmd.AddCommand(startCmd)
	addOutputFormatFlag(startCmd)
	startCmd.Flags().AddFlagSet(startCmdFlagSet)

	_ = crcConfig.BindFlagSet(startCmd.Flags())
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the OpenShift cluster",
	Long:  "Start the OpenShift cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runStart(args); err != nil {
			exit.WithMessage(1, err.Error())
		}
	},
}

var (
	startCmdFlagSet = initStartCmdFlagSet()
)

func runStart(arguments []string) error {
	if err := validateStartFlags(); err != nil {
		return err
	}

	checkIfNewVersionAvailable(crcConfig.GetBool(config.DisableUpdateCheck.Name))

	preflight.StartPreflightChecks()

	startConfig := machine.StartConfig{
		Name:          constants.DefaultName,
		BundlePath:    crcConfig.GetString(config.Bundle.Name),
		Memory:        crcConfig.GetInt(config.Memory.Name),
		CPUs:          crcConfig.GetInt(config.CPUs.Name),
		NameServer:    crcConfig.GetString(config.NameServer.Name),
		GetPullSecret: getPullSecretFileContent,
		Debug:         isDebugLog(),
	}

	client := machine.NewClient()
	result, err := client.Start(startConfig)
	return render(&startResult{
		Success:       err == nil,
		Error:         errorMessage(err),
		ClusterConfig: toClusterConfig(result),
	}, os.Stdout, outputFormat)
}

func toClusterConfig(result machine.StartResult) *clusterConfig {
	if result.Error != "" {
		return nil
	}
	return &clusterConfig{
		URL: result.ClusterConfig.ClusterAPI,
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

func errorMessage(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

type clusterConfig struct {
	URL                  string      `json:"url"`
	AdminCredentials     credentials `json:"adminCredentials"`
	DeveloperCredentials credentials `json:"developerCredentials"`
}

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type startResult struct {
	Success       bool           `json:"success"`
	Error         string         `json:"error,omitempty"`
	ClusterConfig *clusterConfig `json:"clusterConfig,omitempty"`
}

func (s *startResult) prettyPrintTo(writer io.Writer) error {
	if s.Error != "" {
		return errors.New(s.Error)
	}
	if s.ClusterConfig == nil {
		return errors.New("either Error or ClusterConfig are needed")
	}

	_, err := fmt.Fprintln(writer, strings.Join([]string{
		"Started the OpenShift cluster",
		"",
		"To access the cluster, first set up your environment by following 'crc oc-env' instructions.",
		fmt.Sprintf("Then you can access it by running 'oc login -u %s -p %s %s'.", s.ClusterConfig.DeveloperCredentials.Username, s.ClusterConfig.DeveloperCredentials.Password, s.ClusterConfig.URL),
		fmt.Sprintf("To login as an admin, run 'oc login -u %s -p %s %s'.", s.ClusterConfig.AdminCredentials.Username, s.ClusterConfig.AdminCredentials.Password, s.ClusterConfig.URL),
		"To access the cluster, first set up your environment by following 'crc oc-env' instructions.",
		"",
		"You can now run 'crc console' and use these credentials to access the OpenShift web console.",
	}, "\n"))
	return err
}

func initStartCmdFlagSet() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("start", pflag.ExitOnError)
	flagSet.StringP(config.Bundle.Name, "b", constants.DefaultBundlePath, "The system bundle used for deployment of the OpenShift cluster")
	flagSet.StringP(config.PullSecretFile.Name, "p", "", fmt.Sprintf("File path of image pull secret (download from %s)", constants.CrcLandingPageURL))
	flagSet.IntP(config.CPUs.Name, "c", constants.DefaultCPUs, "Number of CPU cores to allocate to the OpenShift cluster")
	flagSet.IntP(config.Memory.Name, "m", constants.DefaultMemory, "MiB of memory to allocate to the OpenShift cluster")
	flagSet.StringP(config.NameServer.Name, "n", "", "IPv4 address of nameserver to use for the OpenShift cluster")
	flagSet.Bool(config.DisableUpdateCheck.Name, false, "Don't check for update")

	return flagSet
}

func isDebugLog() bool {
	return logging.LogLevel == "debug"
}

func validateStartFlags() error {
	if err := validation.ValidateMemory(crcConfig.GetInt(config.Memory.Name)); err != nil {
		return err
	}
	if err := validation.ValidateCPUs(crcConfig.GetInt(config.CPUs.Name)); err != nil {
		return err
	}
	if err := validation.ValidateBundle(crcConfig.GetString(config.Bundle.Name)); err != nil {
		return err
	}
	if crcConfig.GetString(config.NameServer.Name) != "" {
		if err := validation.ValidateIPAddress(crcConfig.GetString(config.NameServer.Name)); err != nil {
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

	// In case user doesn't provide a file in start command or in config then ask for it.
	if crcConfig.GetString(config.PullSecretFile.Name) == "" {
		pullsecret, err = input.PromptUserForSecret("Image pull secret", fmt.Sprintf("Copy it from %s", constants.CrcLandingPageURL))
		// This is just to provide a new line after user enter the pull secret.
		fmt.Println()
		if err != nil {
			return "", errors.New(err.Error())
		}
	} else {
		// Read the file content
		data, err := ioutil.ReadFile(crcConfig.GetString(config.PullSecretFile.Name))
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
