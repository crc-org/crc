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
	"os"
	"os/user"
	"path/filepath"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"

	clicumber "github.com/code-ready/clicumber/testsuite"
)

// FeatureContext defines godog.Suite steps for the test suite.
func FeatureContext(s *godog.Suite) {

	// CRC related steps
	s.Step(`^removing CRC home directory succeeds$`,
		RemoveCRCHome)

	// CRC file operations
	s.Step(`^file "([^"]*)" exists in CRC home folder$`,
		FileExistsInCRCHome)
	s.Step(`"(JSON|YAML)" config file "(.*)" in CRC home folder (contains|does not contain) key "(.*)" with value matching "(.*)"$`,
		ConfigFileInCRCHomeContainsKeyMatchingValue)
	s.Step(`"(JSON|YAML)" config file "(.*)" in CRC home folder (contains|does not contain) key "(.*)"$`,
		ConfigFileInCRCHomeContainsKey)
	s.Step(`removing file "(.*)" from CRC home folder succeeds$`,
		DeleteFileFromCRCHome)

	s.BeforeSuite(func() {

		// delete existing crc instance (if any)
		err := DeleteCRC()
		if err != nil {
			fmt.Println(err)
		}

		// remove $HOME/.crc
		err = RemoveCRCHome()
		if err != nil {
			fmt.Println(err)
		}
	})

	s.AfterFeature(func(this *gherkin.Feature) {
		err := DeleteCRC()
		if err != nil {
			fmt.Println(err)
		}
	})

}

func DeleteFileFromCRCHome(fileName string) error {

	usr, err := user.Current()
	if err != nil {
		return err
	}
	theFile := filepath.Join(usr.HomeDir, ".crc", fileName)

	_, err = os.Stat(theFile)

	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("file %s neither exists nor doesn't exist, error: %v", theFile, err)
	}

	err = clicumber.DeleteFile(theFile)
	if err != nil {
		fmt.Errorf("Error deleting file %v", theFile)
	}
	return nil // == err
}

func FileExistsInCRCHome(fileName string) error {

	usr, err := user.Current()
	if err != nil {
		return err
	}
	theFile := filepath.Join(usr.HomeDir, ".crc", fileName)

	_, err = os.Stat(theFile)

	if os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exists, error: %v ", theFile, err)
	} else if err != nil {
		return fmt.Errorf("file %s neither exists nor doesn't exist, error: %v", theFile, err)
	}
	return nil // == err
}

func ConfigFileInCRCHomeContainsKeyMatchingValue(format string, configFile string, condition string, keyPath string, expectedValue string) error {

	usr, err := user.Current()
	if err != nil {
		return err
	}
	configPath := filepath.Join(usr.HomeDir, ".crc", configFile)

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

	usr, err := user.Current()
	if err != nil {
		return err
	}
	configPath := filepath.Join(usr.HomeDir, ".crc", configFile)

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

func RemoveCRCHome() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	crcHome := filepath.Join(usr.HomeDir, ".crc")
	command := "rm -rf " + crcHome

	err = clicumber.ExecuteCommand(command)
	if err != nil {
		return err
	}
	fmt.Printf("- Deleted CRC home dir (if present).\n")
	return nil
}

func DeleteCRC() error {

	command := "crc delete"
	_ = clicumber.ExecuteCommand(command)

	fmt.Printf("- Deleted CRC instance (if one existed).\n")
	return nil
}
