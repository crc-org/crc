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
	"crypto/tls"
	"fmt"
	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	clicumber "github.com/code-ready/clicumber/testsuite"
	"github.com/code-ready/crc/pkg/crc/oc"
)

var (
	CRCHome        string
	CRCBinary      string
	bundleEmbedded bool
	bundleName     string
	bundleURL      string
	bundleVersion  string
	pullSecretFile string
	goPath         string
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
	s.Step(`^starting CRC with default bundle and nameserver "(.*)" (succeeds|fails)$`,
		StartCRCWithDefaultBundleAndNameServerSucceedsOrFails)
	s.Step(`^setting config property "(.*)" to value "(.*)" (succeeds|fails)$`,
		SetConfigPropertyToValueSucceedsOrFails)
	s.Step(`^unsetting config property "(.*)" (succeeds|fails)$`,
		UnsetConfigPropertySucceedsOrFails)
	s.Step(`^login to the oc cluster (succeeds|fails)$`,
		LoginToOcClusterSucceedsOrFails)
	s.Step(`^with up to "(\d+)" retries with wait period of "(\d*(?:ms|s|m))" all cluster operators are running$`,
		CheckClusterOperatorsWithRetry)
	s.Step(`^with up to "(\d+)" retries with wait period of "(\d*(?:ms|s|m))" http response from "(.*)" has status code "(\d+)"$`,
		CheckHTTPResponseWithRetry)
	s.Step(`^with up to "(\d+)" retries with wait period of "(\d*(?:ms|s|m))" command "(.*)" output (?:should match|matches) "(.*)"$`,
		CheckOutputMatchWithRetry)
	s.Step(`stdout (?:should contain|contains) "(.*)" if bundle (is|is not) embedded$`,
		StdoutContainsIfBundleEmbeddedOrNot)

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

		usr, _ := user.Current()
		CRCHome = filepath.Join(usr.HomeDir, ".crc")

		// init CRCBinary if no location provided by user
		if CRCBinary == "" {
			fmt.Println("Expecting the CRC binary to be in $HOME/go/bin.")
			usr, _ := user.Current()
			CRCBinary = filepath.Join(usr.HomeDir, "go", "bin")
		}

		// put CRC binary location on top of PATH
		path := os.Getenv("PATH")
		newPath := fmt.Sprintf("%s%s%s", CRCBinary, os.PathListSeparator, path)
		err := os.Setenv("PATH", newPath)
		if err != nil {
			fmt.Println("Could not put CRC location on top of PATH")
			os.Exit(1)
		}

		if bundleURL == "embedded" {
			fmt.Println("Expecting the bundle to be embedded in the CRC binary.")
			bundleEmbedded = true
			if bundleVersion == "" {
				fmt.Println("User must specify --bundle-version if bundle is embedded")
				os.Exit(1)
			}
			// assume default hypervisor
			var hypervisor string
			switch platform := runtime.GOOS; platform {
			case "darwin":
				hypervisor = "hyperkit"
			case "linux":
				hypervisor = "libvirt"
			case "windows":
				hypervisor = "hyperv"
			default:
				fmt.Printf("Unsupported OS: %s", platform)
				os.Exit(1)
			}
			bundleName = fmt.Sprintf("crc_%s_%s.crcbundle", hypervisor, bundleVersion)
		} else {
			bundleEmbedded = false
			_, bundleName = filepath.Split(bundleURL)
		}

		if pullSecretFile == "" {
			fmt.Println("User must specify the pull secret file via --pull-secret-file flag.")
			os.Exit(1)
		}

		// remove $HOME/.crc
		err = RemoveCRCHome()
		if err != nil {
			fmt.Println(err)
		}

	})

	s.AfterSuite(func() {
		err := DeleteCRC()
		if err != nil {
			fmt.Printf("Could not delete CRC VM: %s.", err)
		}
	})

	s.BeforeFeature(func(this *gherkin.Feature) {

		if bundleEmbedded == false {
			if _, err := os.Stat(bundleName); os.IsNotExist(err) {
				// Obtain the bundle to current dir
				fmt.Println("Obtaining bundle...")
				bundle, err := DownloadBundle(bundleURL, ".")
				if err != nil {
					fmt.Printf("Failed to obtain CRC bundle, %v\n", err)
					os.Exit(1)
				}
				fmt.Println("Using bundle:", bundle)
			} else if err != nil {
				fmt.Printf("Unexpected error obtaining the bundle %v.\n", bundleName)
				os.Exit(1)
			} else {
				fmt.Println("Using existing bundle:", bundleName)
			}
		}
	})
}

func CheckClusterOperatorsWithRetry(retryCount int, retryWait string) error {

	retryDuration, err := time.ParseDuration(retryWait)
	if err != nil {
		return err
	}

	ocConfig := oc.UseOCWithConfig("crc")
	for i := 0; i < retryCount; i++ {
		s, err := oc.GetClusterOperatorStatus(ocConfig)
		if err != nil {
			return err
		}
		if s == true {
			return nil
		}
		time.Sleep(retryDuration)
	}

	return fmt.Errorf("Some cluster operators are still not running.\n")
}

func CheckHTTPResponseWithRetry(retryCount int, retryWait string, address string, expectedStatusCode int) error {

	retryDuration, err := time.ParseDuration(retryWait)
	if err != nil {
		return err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	var resp *http.Response
	for i := 0; i < retryCount; i++ {
		resp, err = client.Get(address)
		if err != nil {
			return err
		}
		if resp.StatusCode == expectedStatusCode {
			return nil
		}
		time.Sleep(retryDuration)
	}

	return fmt.Errorf("Got %d as Status Code instead of expected %d.", resp.StatusCode, expectedStatusCode)
}

func CheckOutputMatchWithRetry(retryCount int, retryTime string, command string, expected string) error {

	retryDuration, err := time.ParseDuration(retryTime)
	if err != nil {
		return err
	}

	var match_err error

	for i := 0; i < retryCount; i++ {
		exec_err := clicumber.ExecuteCommand(command)
		if exec_err == nil {
			match_err = clicumber.CommandReturnShouldMatch("stdout", expected)
			if match_err == nil {
				return nil
			}
		}
		time.Sleep(retryDuration)
	}

	return match_err
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
		expectedValue = bundleName
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

func LoginToOcClusterSucceedsOrFails(expected string) error {

	bundle := strings.Split(bundleName, ".crcbundle")[0]
	pswdLocation := filepath.Join(CRCHome, "cache", bundle, "kubeadmin-password")

	pswd, err := ioutil.ReadFile(pswdLocation)

	if err != nil {
		return err
	}

	cmd := fmt.Sprintf("oc login --insecure-skip-tls-verify -u kubeadmin -p %s https://api.crc.testing:6443", pswd)
	err = clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)

	return err
}

func StartCRCWithDefaultBundleAndDefaultHypervisorSucceedsOrFails(expected string) error {

	var cmd string
	var extraBundleArgs string

	if bundleEmbedded == false {
		extraBundleArgs = fmt.Sprintf("-b %s", bundleName)
	}
	cmd = fmt.Sprintf("crc start -p '%s' %s --log-level debug", pullSecretFile, extraBundleArgs)
	err := clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)

	return err
}

func StartCRCWithDefaultBundleAndHypervisorSucceedsOrFails(hypervisor string, expected string) error {

	var cmd string
	var extraBundleArgs string

	if bundleEmbedded == false {
		extraBundleArgs = fmt.Sprintf("-b %s", bundleName)
	}
	cmd = fmt.Sprintf("crc start -d %s -p '%s' %s --log-level debug", hypervisor, pullSecretFile, extraBundleArgs)
	err := clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)

	return err
}

func StartCRCWithDefaultBundleAndNameServerSucceedsOrFails(nameserver string, expected string) error {

	var extraBundleArgs string
	if bundleEmbedded == false {
		extraBundleArgs = fmt.Sprintf("-b %s", bundleName)
	}

	var cmd string

	cmd = fmt.Sprintf("crc start -n %s -p '%s' %s --log-level debug", nameserver, pullSecretFile, extraBundleArgs)
	err := clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)

	return err
}

func StdoutContainsIfBundleEmbeddedOrNot(value string, expected string) error {

	if expected == "is" { // expect embedded
		if bundleEmbedded { // really embedded
			return clicumber.CommandReturnShouldContain("stdout", value)
		} else {
			return clicumber.CommandReturnShouldNotContain("stdout", value)
		}
	} else { // expect not embedded
		if !bundleEmbedded { // really not embedded
			return clicumber.CommandReturnShouldContain("stdout", value)
		} else {
			return clicumber.CommandReturnShouldNotContain("stdout", value)
		}
	}

}

func SetConfigPropertyToValueSucceedsOrFails(property string, value string, expected string) error {

	if value == "current bundle" {

		if bundleEmbedded {
			value = filepath.Join(CRCHome, bundleName)
		} else {
			value = bundleName
		}
	}

	cmd := "crc config set " + property + " " + value
	err := clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)

	return err
}

func UnsetConfigPropertySucceedsOrFails(property string, expected string) error {

	cmd := "crc config unset " + property
	err := clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)

	return err
}
