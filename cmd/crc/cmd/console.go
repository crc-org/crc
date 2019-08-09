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
	consoleURLMode bool
)

func init() {
	consoleCmd.Flags().BoolVar(&consoleURLMode, "url", false, "Prints the OpenShift Web Console URL to the console.")
	rootCmd.AddCommand(consoleCmd)
}

// consoleCmd represents the console command
var consoleCmd = &cobra.Command{
	Use:     "console",
	Aliases: []string{"dashboard"},
	Short:   "Opens or displays the OpenShift Web Console URL.",
	Long:    `Opens the OpenShift Web Console URL in the default browser or displays it to the console.`,
	Run: func(cmd *cobra.Command, args []string) {
		runConsole(args)
	},
}

func runConsole(arguments []string) {
	consoleConfig := machine.ConsoleConfig{
		Name: constants.DefaultName,
	}
	result, err := machine.GetConsoleURL(consoleConfig)
	if err != nil {
		errors.Exit(1)
	}

	if consoleURLMode {
		output.Out(result.URL)
		return
	}
	output.Out("Opening the OpenShift Web console in the default browser...")
	browser.OpenURL(result.URL)
}
