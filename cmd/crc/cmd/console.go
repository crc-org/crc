package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	crcErrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/machine/state"
	"github.com/code-ready/crc/pkg/crc/machine/types"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var (
	consolePrintURL         bool
	consolePrintCredentials bool
)

func init() {
	addOutputFormatFlag(consoleCmd)
	consoleCmd.Flags().BoolVar(&consolePrintURL, "url", false, "Print the URL for the OpenShift Web Console")
	consoleCmd.Flags().BoolVar(&consolePrintCredentials, "credentials", false, "Print the credentials for the OpenShift Web Console")
	rootCmd.AddCommand(consoleCmd)
}

// consoleCmd represents the console command
var consoleCmd = &cobra.Command{
	Use:     "console",
	Aliases: []string{"dashboard"},
	Short:   "Open the OpenShift Web Console in the default browser",
	Long:    `Open the OpenShift Web Console in the default browser or print its URL or credentials`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConsole(os.Stdout, newMachine(), consolePrintURL, consolePrintCredentials, outputFormat)
	},
}

func showConsole(client machine.Client) (*types.ConsoleResult, error) {
	if err := checkIfMachineMissing(client); err != nil {
		// In case of machine doesn't exist then consoleResult error
		// should be updated so that when rendering the result it have
		// error details also.
		return nil, err
	}
	return client.GetConsoleURL()
}

func runConsole(writer io.Writer, client machine.Client, consolePrintURL, consolePrintCredentials bool, outputFormat string) error {
	result, err := showConsole(client)
	return render(&consoleResult{
		Success:                 err == nil,
		state:                   toState(result),
		ClusterConfig:           toConsoleClusterConfig(result),
		Error:                   crcErrors.ToSerializableError(err),
		consolePrintURL:         consolePrintURL,
		consolePrintCredentials: consolePrintCredentials,
	}, writer, outputFormat)
}

type consoleResult struct {
	Success                 bool `json:"success"`
	state                   state.State
	Error                   *crcErrors.SerializableError `json:"error,omitempty"`
	ClusterConfig           *clusterConfig               `json:"clusterConfig,omitempty"`
	consolePrintURL         bool
	consolePrintCredentials bool
}

func (s *consoleResult) prettyPrintTo(writer io.Writer) error {
	if s.Error != nil {
		return s.Error
	}
	if s.consolePrintURL {
		if _, err := fmt.Fprintln(writer, s.ClusterConfig.WebConsoleURL); err != nil {
			return err
		}
	}

	if s.consolePrintCredentials {
		if _, err := fmt.Fprintf(writer, "To login as a regular user, run 'oc login -u %s -p %s %s'.\n",
			s.ClusterConfig.DeveloperCredentials.Username, s.ClusterConfig.DeveloperCredentials.Password, s.ClusterConfig.URL); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(writer, "To login as an admin, run 'oc login -u %s -p %s %s'\n",
			s.ClusterConfig.AdminCredentials.Username, s.ClusterConfig.AdminCredentials.Password, s.ClusterConfig.URL); err != nil {
			return err
		}
	}
	if s.consolePrintURL || s.consolePrintCredentials {
		return nil
	}

	if s.state != state.Running {
		return errors.New("The OpenShift cluster is not running, cannot open the OpenShift Web Console")
	}

	if _, err := fmt.Fprintln(writer, "Opening the OpenShift Web Console in the default browser..."); err != nil {
		return err
	}
	if err := browser.OpenURL(s.ClusterConfig.WebConsoleURL); err != nil {
		return fmt.Errorf("Failed to open the OpenShift Web Console, you can access it by opening %s in your web browser", s.ClusterConfig.WebConsoleURL)
	}

	return nil
}

func toState(result *types.ConsoleResult) state.State {
	if result == nil {
		return state.Error
	}
	return result.State
}

func toConsoleClusterConfig(result *types.ConsoleResult) *clusterConfig {
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
