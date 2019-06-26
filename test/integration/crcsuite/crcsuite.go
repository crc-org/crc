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

package crcsuite

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"

	clicumber "github.com/code-ready/clicumber/testsuite"
)

var (
	CRCHome        string
	BundleLocation string
	BundleName     string
)

// FeatureContext defines godog.Suite steps for the test suite.
func FeatureContext(s *godog.Suite) {

	// CRC related steps
	s.Step(`^removing CRC home directory succeeds$`,
		RemoveCRCHome)
	s.Step(`^starting CRC with default bundle and default hypervisor (succeeds|fails)$`,
		StartCRCWithDefaultBundleAndDefaultHypervisorSucceedsOrFails)
	s.Step(`^starting CRC with default bundle and hypervisor "(.*)" (succeeds|fails)$`,
		StartCRCWithDefaultBundleAndHypervisorSucceedsOrFails)
	s.Step(`^setting config property "(.*)" to value "(.*)" (succeeds|fails)$`,
		SetConfigPropertyToValueSucceedsOrFails)

	// CRC file operations
	s.Step(`^file "([^"]*)" exists in CRC home folder$`,
		FileExistsInCRCHome)
	s.Step(`"(JSON|YAML)" config file "(.*)" in CRC home folder (contains|does not contain) key "(.*)" with value matching "(.*)"$`,
		ConfigFileInCRCHomeContainsKeyMatchingValue)
	s.Step(`"(JSON|YAML)" config file "(.*)" in CRC home folder (contains|does not contain) key "(.*)"$`,
		ConfigFileInCRCHomeContainsKey)
	s.Step(`removing file "(.*)" from CRC home folder succeeds$`,
		DeleteFileFromCRCHome)

	// OpenShift steps

	s.Step(`^cluster operator "(.*)" (is|is not|is not known) (available|progressing)$`,
		verifyCOStatus)
	s.Step(`^check at most "(\d+)" times with delay of "(.*)" that cluster operator "(.*)" (is|is not|is not known) (available|progressing)$`,
		verifyCOStatusWithRetry)
	s.Step(`^check at most "(\d+)" times with delay of "(.*)" that pod "(.*)" (is|is not) (initialized|ready)$`,
		verifyPodStatusWithRetry)

	s.BeforeSuite(func() {
		// Set suite vars
		CRCHome = SetCRCHome()
		BundleName = SetBundleName()

		// Remove $HOME/.crc
		err := RemoveCRCHome()
		if err != nil {
			fmt.Println(err)
		}

	})

	s.AfterSuite(func() {

		ForceStopCRC()
		DeleteCRC()
	})

	s.BeforeFeature(func(this *gherkin.Feature) {

		if _, err := os.Stat(BundleName); os.IsNotExist(err) {
			// Obtain the bundle to current dir
			fmt.Println("Obtaining bundle...")
			bundle, err := GetBundle(BundleLocation, ".")
			if err != nil {
				fmt.Errorf("Failed to obtain CRC bundle, %v\n", err)
				os.Exit(1)
			} else {
				fmt.Println("Using bundle:", bundle)
			}
		} else if err != nil {
			fmt.Errorf("Unknown error obtaining the bundle %v.\n", BundleName)
		} else {
			fmt.Println("Using existing bundle:", BundleName)
		}

	})
}

func DeleteFileFromCRCHome(fileName string) error {

	theFile := filepath.Join(CRCHome, fileName)

	if _, err := os.Stat(theFile); os.IsNotExist(err) {
		return nil
	}

	err := clicumber.DeleteFile(theFile)
	if err != nil {
		fmt.Errorf("Error deleting file %v", theFile)
	}
	return nil
}

func FileExistsInCRCHome(fileName string) error {

	theFile := filepath.Join(CRCHome, fileName)

	_, err := os.Stat(theFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exists, error: %v ", theFile, err)
	}

	return err
}

func ConfigFileInCRCHomeContainsKeyMatchingValue(format string, configFile string, condition string, keyPath string, expectedValue string) error {

	if expectedValue == "current bundle" {
		expectedValue = BundleName
	}
	configPath := filepath.Join(CRCHome, configFile)

	config, err := clicumber.GetFileContent(configPath)
	if err != nil {
		return err
	}

	keyValue, err := clicumber.GetConfigKeyValue([]byte(config), format, keyPath)
	if err != nil {
		return err
	}

	matches, err := clicumber.PerformRegexMatch(expectedValue, keyValue)
	if err != nil {
		return err
	} else if (condition == "contains") && !matches {
		return fmt.Errorf("For key '%s' config contains unexpected value '%s'", keyPath, keyValue)
	} else if (condition == "does not contain") && matches {
		return fmt.Errorf("For key '%s' config contains value '%s', which it should not contain", keyPath, keyValue)
	}

	return nil
}

func ConfigFileInCRCHomeContainsKey(format string, configFile string, condition string, keyPath string) error {

	configPath := filepath.Join(CRCHome, configFile)

	config, err := clicumber.GetFileContent(configPath)
	if err != nil {
		return err
	}

	keyValue, err := clicumber.GetConfigKeyValue([]byte(config), format, keyPath)
	if err != nil {
		return err
	}

	if (condition == "contains") && (keyValue == "<nil>") {
		return fmt.Errorf("Config does not contain any value for key %s", keyPath)
	} else if (condition == "does not contain") && (keyValue != "<nil>") {
		return fmt.Errorf("Config contains key %s with assigned value: %s", keyPath, keyValue)
	}

	return nil
}

func StartCRCWithDefaultBundleAndDefaultHypervisorSucceedsOrFails(expected string) error {

	cmd := "crc start -b " + BundleName
	err := clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)

	return err
}

func StartCRCWithDefaultBundleAndHypervisorSucceedsOrFails(hypervisor string, expected string) error {

	cmd := "crc start -b " + BundleName + " -d " + hypervisor
	err := clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)

	return err
}

func SetConfigPropertyToValueSucceedsOrFails(property string, value string, expected string) error {

	if value == "current bundle" {
		value = BundleName
	}

	cmd := "crc config set " + property + " " + value
	err := clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)

	return err
}
