package cmd

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var (
	consolePrintURL         bool
	consolePrintCredentials bool
)

func init() {
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
	Run: func(cmd *cobra.Command, args []string) {
		runConsole(args)
	},
}

func runConsole(arguments []string) {
	consoleConfig := machine.ConsoleConfig{
		Name: constants.DefaultName,
	}

	exitIfMachineMissing(consoleConfig.Name)

	result, err := machine.GetConsoleURL(consoleConfig)
	if err != nil {
		errors.Exit(1)
	}

	if consolePrintURL {
		output.Outln(result.ClusterConfig.WebConsoleURL)
	}
	if consolePrintCredentials {
		output.Outf("To login as a regular user, run 'oc login -u developer -p developer %s'.\n", result.ClusterConfig.ClusterAPI)
		output.Outf("To login as an admin, run 'oc login -u kubeadmin -p %s %s'\n", result.ClusterConfig.KubeAdminPass, result.ClusterConfig.ClusterAPI)
	}
	if consolePrintURL || consolePrintCredentials {
		return
	}

	if !machine.IsRunning(result.State) {
		errors.ExitWithMessage(1, "The OpenShift cluster is not running, cannot open the OpenShift Web Console.")
	}
	output.Outln("Opening the OpenShift Web Console in the default browser...")
	err = browser.OpenURL(result.ClusterConfig.WebConsoleURL)
	if err != nil {
		errors.ExitWithMessage(1, "Failed to open the OpenShift Web Console, you can access it by opening %s in your web browser.", result.ClusterConfig.WebConsoleURL)
	}
}
