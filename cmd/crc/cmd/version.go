/*
Copyright (C) 2019 Red Hat, Inc.

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
	"fmt"

	crcPkg "github.com/code-ready/crc/pkg/crc"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get CodeReady Container version.",
	Long:  "Get the CodeReady Container version currently installed.",
	Run: func(cmd *cobra.Command, args []string) {
		runPrintVersion(args)
	},
}

func runPrintVersion(arguments []string) {
	fmt.Printf("version: %s+%s\n", crcPkg.GetCRCVersion(), crcPkg.GetCommitSha())
}
