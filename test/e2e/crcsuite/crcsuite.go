package crcsuite

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v10"

	clicumber "github.com/code-ready/clicumber/testsuite"
)

var (
	CRCHome        string
	CRCExecutable  string
	bundleEmbedded bool
	bundleName     string
	bundleLocation string
	bundleVersion  string
	pullSecretFile string
)

// FeatureContext defines godog.Suite steps for the test suite.
func FeatureContext(s *godog.Suite) {

	// CRC related steps
	s.Step(`^removing CRC home directory succeeds$`,
		RemoveCRCHome)
	s.Step(`^starting CRC with default bundle (succeeds|fails)$`,
		StartCRCWithDefaultBundleSucceedsOrFails)
	s.Step(`^starting CRC with default bundle along with stopped network time synchronization (succeeds|fails)$`,
		StartCRCWithDefaultBundleWithStopNetworkTimeSynchronizationSucceedsOrFails)
	s.Step(`^starting CRC with default bundle and nameserver "(.*)" (succeeds|fails)$`,
		StartCRCWithDefaultBundleAndNameServerSucceedsOrFails)
	s.Step(`^setting config property "(.*)" to value "(.*)" (succeeds|fails)$`,
		SetConfigPropertyToValueSucceedsOrFails)
	s.Step(`^unsetting config property "(.*)" (succeeds|fails)$`,
		UnsetConfigPropertySucceedsOrFails)
	s.Step(`^login to the oc cluster (succeeds|fails)$`,
		LoginToOcClusterSucceedsOrFails)
	s.Step(`^setting kubeconfig context to "(.*)" (succeeds|fails)$`,
		SetKubeConfigContextSucceedsOrFails)
	s.Step(`^with up to "(\d+)" retries with wait period of "(\d*(?:ms|s|m))" http response from "(.*)" has status code "(\d+)"$`,
		CheckHTTPResponseWithRetry)
	s.Step(`^with up to "(\d+)" retries with wait period of "(\d*(?:ms|s|m))" command "(.*)" output (should match|matches|should not match|does not match) "(.*)"$`,
		CheckOutputMatchWithRetry)
	s.Step(`^checking that CRC is (running|stopped)$`,
		CheckCRCStatus)
	s.Step(`^(stdout|stderr) (?:should contain|contains) "(.*)" if bundle (is|is not) embedded$`,
		CommandReturnShouldContainIfBundleEmbeddedOrNot)

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

		// init CRCExecutable if no location provided by user
		if CRCExecutable == "" {
			fmt.Println("Expecting the CRC executable to be in $HOME/go/bin.")
			usr, _ := user.Current()
			CRCExecutable = filepath.Join(usr.HomeDir, "go", "bin")
		}

		// put CRC executable location on top of PATH
		path := os.Getenv("PATH")
		newPath := fmt.Sprintf("%s%c%s", CRCExecutable, os.PathListSeparator, path)
		err := os.Setenv("PATH", newPath)
		if err != nil {
			fmt.Println("Could not put CRC location on top of PATH")
			os.Exit(1)
		}

		if bundleLocation == "" {
			fmt.Println("Expecting the bundle to be embedded in the CRC executable.")
			bundleEmbedded = true
			if bundleVersion == "" {
				fmt.Println("User must specify --bundle-version if bundle is embedded")
				os.Exit(1)
			}
			bundleName = constants.GetBundleFosOs(runtime.GOOS, bundleVersion)
		} else {
			bundleEmbedded = false
			_, bundleName = filepath.Split(bundleLocation)
		}

		if pullSecretFile == "" {
			fmt.Println("User must specify the pull secret file via --pull-secret-file flag.")
			os.Exit(1)
		}

		// remove $HOME/.crc
		err = RemoveCRCHome()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if !bundleEmbedded {
			if _, err := os.Stat(bundleLocation); err != nil {
				if !os.IsNotExist(err) {
					fmt.Printf("Unexpected error obtaining the bundle %v.\n", bundleLocation)
					os.Exit(1)
				}
				// Obtain the bundle to current dir
				fmt.Println("Obtaining bundle...")
				bundleLocation, err = DownloadBundle(bundleLocation, ".")
				if err != nil {
					fmt.Printf("Failed to obtain CRC bundle, %v\n", err)
					os.Exit(1)
				}
				fmt.Println("Using bundle:", bundleLocation)
			} else {
				fmt.Println("Using existing bundle:", bundleLocation)
			}
		}
	})

	s.BeforeScenario(func(pickle *messages.Pickle) {
		// copy data/config files to test dir
		for _, tag := range pickle.GetTags() {
			if tag.Name == "@testdata" {
				err := CopyFilesToTestDir()
				if err != nil {
					os.Exit(1)
				}
			}
		}
	})

	s.AfterScenario(func(pickle *messages.Pickle, err error) {
		if err != nil {
			if err := runDiagnose(filepath.Join("..", "test-results")); err != nil {
				fmt.Printf("Failed to collect diagnostic: %v\n", err)
			}
		}
	})

	s.AfterSuite(func() {
		err := DeleteCRC()
		if err != nil {
			fmt.Printf("Could not delete CRC VM: %s.", err)
		}
	})
}

func CheckHTTPResponseWithRetry(retryCount int, retryWait string, address string, expectedStatusCode int) error {

	retryDuration, err := time.ParseDuration(retryWait)
	if err != nil {
		return err
	}

	tr := &http.Transport{
		// #nosec G402
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

	return fmt.Errorf("Got %d as Status Code instead of expected %d", resp.StatusCode, expectedStatusCode)
}

func CheckOutputMatchWithRetry(retryCount int, retryTime string, command string, expected string, expectedOutput string) error {

	retryDuration, err := time.ParseDuration(retryTime)
	if err != nil {
		return err
	}

	var matchErr error

	for i := 0; i < retryCount; i++ {
		execErr := clicumber.ExecuteCommand(command)
		if execErr == nil {
			if strings.Contains(expected, " not ") {
				matchErr = clicumber.CommandReturnShouldNotMatch("stdout", expectedOutput)
			} else {
				matchErr = clicumber.CommandReturnShouldMatch("stdout", expectedOutput)
			}
			if matchErr == nil {
				return nil
			}
		}
		time.Sleep(retryDuration)
	}

	return matchErr
}

// CheckCRCStatus checks that output of status command
// matches given regex number of consecutive times
func CheckCRCStatus(state string) error {
	retryDuration := 1 * time.Minute
	retryCount := 15
	cmd := "crc status --log-level debug"
	expression := ""
	var numConsecutive int

	if state == "stopped" {
		numConsecutive = 1
		expression = ".*Stopped.*"
	} else {
		numConsecutive = 3
		expression = `.*Running \(v\d+\.\d+\.\d+.*\).*`
	}

	var matchErr error
	var count int // holds num of consecutive matches

	for i := 0; i < retryCount; i++ {
		execErr := clicumber.ExecuteCommand(cmd)
		if execErr == nil {
			matchErr = clicumber.CommandReturnShouldMatch("stdout", expression)
			// update counter for consecutive matches
			if matchErr == nil {
				count++
			} else {
				count = 0
			}
			// break if done
			if count == numConsecutive {
				return nil
			}
		}
		time.Sleep(retryDuration)
	}

	return fmt.Errorf("did not get enough consecutive matches: have %d but need %d", count, numConsecutive)
}

func DeleteFileFromCRCHome(fileName string) error {

	theFile := filepath.Join(CRCHome, fileName)

	if _, err := os.Stat(theFile); os.IsNotExist(err) {
		return nil
	}

	if err := clicumber.DeleteFile(theFile); err != nil {
		return fmt.Errorf("Error deleting file %v", theFile)
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
		expectedValue = fmt.Sprintf(".*%s", bundleName)
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
	}
	if (condition == "contains") && !matches {
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

func SetKubeConfigContextSucceedsOrFails(context, expected string) error {
	cmd := fmt.Sprintf("oc config use-context %s", context)
	return clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)
}

func StartCRCWithDefaultBundleSucceedsOrFails(expected string) error {

	var cmd string
	var extraBundleArgs string

	if !bundleEmbedded {
		extraBundleArgs = fmt.Sprintf("-b %s", bundleLocation)
	}
	cmd = fmt.Sprintf("crc start -p '%s' %s --log-level debug", pullSecretFile, extraBundleArgs)
	err := clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)

	return err
}

func StartCRCWithDefaultBundleWithStopNetworkTimeSynchronizationSucceedsOrFails(expected string) error {

	var cmd string
	var extraBundleArgs string

	if !bundleEmbedded {
		extraBundleArgs = fmt.Sprintf("-b %s", bundleLocation)
	}
	cmd = fmt.Sprintf("CRC_DEBUG_ENABLE_STOP_NTP=true crc start -p '%s' %s --log-level debug", pullSecretFile, extraBundleArgs)
	err := clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)

	return err
}

func StartCRCWithDefaultBundleAndNameServerSucceedsOrFails(nameserver string, expected string) error {

	var extraBundleArgs string
	if !bundleEmbedded {
		extraBundleArgs = fmt.Sprintf("-b %s", bundleLocation)
	}

	cmd := fmt.Sprintf("crc start -n %s -p '%s' %s --log-level debug", nameserver, pullSecretFile, extraBundleArgs)
	return clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)
}

func CommandReturnShouldContainIfBundleEmbeddedOrNot(commandField string, value string, expected string) error {
	if expected == "is" { // expect embedded
		if bundleEmbedded { // really embedded
			return clicumber.CommandReturnShouldContain(commandField, value)
		}
		return clicumber.CommandReturnShouldNotContain(commandField, value)
	}
	// expect not embedded
	if !bundleEmbedded { // really not embedded
		return clicumber.CommandReturnShouldContain(commandField, value)
	}
	return clicumber.CommandReturnShouldNotContain(commandField, value)
}

func SetConfigPropertyToValueSucceedsOrFails(property string, value string, expected string) error {

	if value == "current bundle" {

		if bundleEmbedded {
			value = filepath.Join(CRCHome, "cache", bundleName)
		} else {
			value = bundleLocation
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
