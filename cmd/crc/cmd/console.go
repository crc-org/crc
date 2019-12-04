/*
Copyright 2016 The Kubernetes Authors All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
