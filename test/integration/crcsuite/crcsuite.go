// +build integration

/*
Copyright (C) 2018 Red Hat, Inc.
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

/*

This file is not needed yet.

*/

package crcsuite

import (
	"fmt"

	"github.com/DATA-DOG/godog"
	"github.com/code-ready/clicumber/testsuite"
)

var sh testsuite.ShellInstance

func crcStateIs(expected string) error {

	fmt.Println("checking state of CRC...")
	err := testsuite.ExecuteCommand("crc state")

	if err != nil {
		return err
	}

	actual := sh.GetLastCmdOutput("stdout")
	if actual != expected {
		return fmt.Errorf("CRC state does not match %s. Expected: %s, Actual: %s.\n", expected, expected, actual)
	}

	return nil
}

// FeatureContext defines godog.Suite steps for the test suite.
func FeatureContext(s *godog.Suite) {
	// Status verification
	s.Step(`^CRC (should be|is) in state "([^"]*)"$`,
		crcStateIs)
}
