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

package util

import (
	"fmt"
	"os"
	"path/filepath"
)

var TestDir string
var TestRunDir string
var TestResultsDir string

func PrepareForE2eTest() error {

	var err error
	if TestDir == "" {
		TestDir, err = os.MkdirTemp("", "crc-e2e-test-")
		if err != nil {
			return fmt.Errorf("error creating temporary directory for test run: %v", err)
		}
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		TestDir = filepath.Join(wd, TestDir)
		err = os.MkdirAll(TestDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating directory for test run: %v", err)
		}
	}

	TestRunDir = filepath.Join(TestDir, "test-run")
	TestResultsDir = filepath.Join(TestDir, "test-results")

	err = PrepareTestRunDir()
	if err != nil {
		return err
	}

	err = PrepareTestResultsDir()
	if err != nil {
		return err
	}

	err = StartLog(TestResultsDir)
	if err != nil {
		return fmt.Errorf("error starting the log: %v", err)
	}

	err = os.Chdir(TestRunDir)
	if err != nil {
		return err
	}

	fmt.Printf("Running e2e test in: %v\n", TestRunDir)
	fmt.Printf("Working directory set to: %v\n", TestRunDir)

	return nil
}

func PrepareTestRunDir() error {
	err := os.MkdirAll(TestRunDir, os.ModePerm)
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
	files, err := os.ReadDir(TestRunDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		err := os.RemoveAll(filepath.Join(TestRunDir, file.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

func PrepareTestResultsDir() error {
	err := os.MkdirAll(TestResultsDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory for test results: %v", err)
	}

	return nil
}
