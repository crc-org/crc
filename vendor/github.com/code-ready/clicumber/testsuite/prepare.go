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

package testsuite

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/code-ready/clicumber/util"
)

//ParseFlags defines flags which are used by test suite.
func ParseFlags() {
	flag.StringVar(&testDir, "test-dir", "out", "Path to the directory in which to execute the tests")
	flag.StringVar(&testWithShell, "test-shell", "", "Specifies shell to be used for the testing.")

	flag.StringVar(&GodogFormat, "godog.format", "pretty", "Sets which format godog will use")
	flag.StringVar(&GodogTags, "godog.tags", "", "Tags for godog test")
	flag.BoolVar(&GodogShowStepDefinitions, "godog.definitions", false, "")
	flag.BoolVar(&GodogStopOnFailure, "godog.stop-on-failure ", false, "Stop when failure is found")
	flag.BoolVar(&GodogNoColors, "godog.no-colors", false, "Disable colors in godog output")
	flag.StringVar(&GodogPaths, "godog.paths", "./features", "")

	flag.Parse()
}

func PrepareForIntegrationTest() error {
	var err error
	if testDir == "" {
		testDir, err = ioutil.TempDir("", "crc-integration-test-")
		if err != nil {
			return fmt.Errorf("error creating temporary directory for test run: %v", err)
		}
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		testDir = filepath.Join(wd, testDir)
		err = os.MkdirAll(testDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating directory for test run: %v", err)
		}
	}

	testRunDir = filepath.Join(testDir, "test-run")
	testResultsDir = filepath.Join(testDir, "test-results")
	testDefaultHome = filepath.Join(testRunDir, ".crc")

	err = PrepareTestRunDir()
	if err != nil {
		return err
	}

	PrepareTestResultsDir()
	if err != nil {
		return err
	}

	err = util.StartLog(testResultsDir)
	if err != nil {
		return fmt.Errorf("error starting the log: %v", err)
	}

	err = os.Chdir(testRunDir)
	if err != nil {
		return err
	}

	fmt.Printf("Running integration test in: %v\n", testRunDir)
	fmt.Printf("Working directory set to: %v\n", testRunDir)

	return nil
}

func PrepareTestRunDir() error {
	err := os.MkdirAll(testRunDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory for test run: %v", err)
	}

	err = CleanTestRunDir()
	if err != nil {
		return err
	}

	return nil
}

func CleanTestRunDir() error {
	files, err := ioutil.ReadDir(testRunDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		err := os.RemoveAll(filepath.Join(testRunDir, file.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

func PrepareTestResultsDir() error {
	err := os.MkdirAll(testResultsDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory for test results: %v", err)
	}

	return nil
}
