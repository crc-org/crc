package crcsuite

import (
	"crypto/tls"
	"flag"
	"fmt"
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
	"github.com/code-ready/crc/test/e2e/crcsuite/ux"
	crcCmd "github.com/code-ready/crc/test/extended/crc/cmd"
	"github.com/code-ready/crc/test/extended/util"
)

var (
	CRCHome        string
	CRCExecutable  string
	bundleEmbedded bool
	bundleName     string
	bundleLocation string
	bundleVersion  string
	pullSecretFile string
	cleanupHome    bool
)

func usingPreexistingCluster() bool {
	return strings.Contains(clicumber.GodogTags, "~@startstop")
}

// FeatureContext defines godog.Suite steps for the test suite.
func FeatureContext(s *godog.Suite) {

	// CRC related steps
	s.Step(`^removing CRC home directory succeeds$`,
		RemoveCRCHome)
	s.Step(`^starting CRC with default bundle (succeeds|fails)$`,
		StartCRCWithDefaultBundleSucceedsOrFails)
	s.Step(`^starting CRC with custom bundle (succeeds|fails)$`,
		StartCRCWithCustomBundleSucceedsOrFails)
	s.Step(`^starting CRC with default bundle along with stopped network time synchronization (succeeds|fails)$`,
		StartCRCWithDefaultBundleWithStopNetworkTimeSynchronizationSucceedsOrFails)
	s.Step(`^starting CRC with default bundle and nameserver "(.*)" (succeeds|fails)$`,
		StartCRCWithDefaultBundleAndNameServerSucceedsOrFails)
	s.Step(`^setting config property "(.*)" to value "(.*)" (succeeds|fails)$`,
		SetConfigPropertyToValueSucceedsOrFails)
	s.Step(`^unsetting config property "(.*)" (succeeds|fails)$`,
		crcCmd.UnsetConfigPropertySucceedsOrFails)
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
	s.Step(`^execut(?:e|ing) crc (.*) command$`,
		ExecuteCommand)
	s.Step(`^execut(?:e|ing) crc (.*) command (.*)$`,
		ExecuteCommandWithExpectedExitStatus)

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

		// Force debug logs
		err := os.Setenv("CRC_LOG_LEVEL", "debug")
		if err != nil {
			fmt.Println("Could not set `CRC_LOG_LEVEL` to `debug`:", err)
		}

		// put CRC executable location on top of PATH
		path := os.Getenv("PATH")
		newPath := fmt.Sprintf("%s%c%s", CRCExecutable, os.PathListSeparator, path)
		err = os.Setenv("PATH", newPath)
		if err != nil {
			fmt.Println("Could not put CRC location on top of PATH")
			os.Exit(1)
		}

		// If we are running the tests against an existing, already
		// running cluster, we don't need a bundle nor a pull secret,
		// and we don't want to remove ~/.crc, so bail out early.
		if usingPreexistingCluster() {
			return
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

		if cleanupHome {
			// remove $HOME/.crc
			err = util.RemoveCRCHome(CRCHome)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		if !bundleEmbedded {
			if _, err := os.Stat(bundleLocation); err != nil {
				if !os.IsNotExist(err) {
					fmt.Printf("Unexpected error obtaining the bundle %v.\n", bundleLocation)
					os.Exit(1)
				}
				// Obtain the bundle to current dir
				fmt.Println("Obtaining bundle...")
				bundleLocation, err = util.DownloadBundle(bundleLocation, ".", bundleName)
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
				err := util.CopyFilesToTestDir()
				if err != nil {
					os.Exit(1)
				}
			}
		}
	})

	s.AfterScenario(func(pickle *messages.Pickle, err error) {
		if usingPreexistingCluster() {
			// collecting diagnostics data is quite slow, and they
			// are not really useful when running the tests locally
			// against an already running cluster
			return
		}
		if err != nil {
			if err := util.RunDiagnose(filepath.Join("..", "test-results")); err != nil {
				fmt.Printf("Failed to collect diagnostic: %v\n", err)
			}
		}
	})

	s.AfterSuite(func() {
		err := crcCmd.DeleteCRC()
		if err != nil {
			fmt.Printf("Could not delete CRC VM: %s.", err)
		}
	})

	// Extend the context with tray when supported
	ux.FeatureContext(s, &bundleLocation, &pullSecretFile)
}

func ParseFlags() {
	flag.StringVar(&bundleLocation, "bundle-location", "", "Path to the bundle to be used in tests")
	flag.StringVar(&pullSecretFile, "pull-secret-file", "", "Path to the file containing pull secret")
	flag.StringVar(&CRCExecutable, "crc-binary", "", "Path to the CRC executable to be tested")
	flag.StringVar(&bundleVersion, "bundle-version", "", "Version of the bundle used in tests")
	flag.BoolVar(&cleanupHome, "cleanup-home", true, "Try to remove crc home folder before starting the suite")

	// Extend the context with tray when supported
	ux.ParseFlags()
}

func WaitForClusterInState(state string) error {
	return crcCmd.WaitForClusterInState(state)
}

func RemoveCRCHome() error {
	return util.RemoveCRCHome(CRCHome)
}

func CheckHTTPResponseWithRetry(retryCount int, retryWait string, address string, expectedStatusCode int) error {
	var err error

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
		if err == nil && resp.StatusCode == expectedStatusCode {
			return nil
		}
		time.Sleep(retryDuration)
	}

	if err != nil {
		return err
	}
	return fmt.Errorf("got %d as Status Code instead of expected %d", resp.StatusCode, expectedStatusCode)
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
// matches given regex
func CheckCRCStatus(state string) error {
	return crcCmd.CheckCRCStatus(state)
}

func DeleteFileFromCRCHome(fileName string) error {

	theFile := filepath.Join(CRCHome, fileName)

	if _, err := os.Stat(theFile); os.IsNotExist(err) {
		return nil
	}

	if err := clicumber.DeleteFile(theFile); err != nil {
		return fmt.Errorf("error deleting file %v", theFile)
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
		return fmt.Errorf("for key '%s' config contains unexpected value '%s'", keyPath, keyValue)
	} else if (condition == "does not contain") && matches {
		return fmt.Errorf("for key '%s' config contains value '%s', which it should not contain", keyPath, keyValue)
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
		return fmt.Errorf("config does not contain any value for key %s", keyPath)
	} else if (condition == "does not contain") && (keyValue != "<nil>") {
		return fmt.Errorf("config contains key %s with assigned value: %s", keyPath, keyValue)
	}

	return nil
}

func LoginToOcClusterSucceedsOrFails(expected string) error {
	cmd := "oc config use-context crc-admin"
	return clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)
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
	crcStart := crcCmd.CRC("start").ToString()
	cmd = fmt.Sprintf("%s -p '%s' %s", crcStart, pullSecretFile, extraBundleArgs)
	err := clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)

	return err
}

func StartCRCWithDefaultBundleWithStopNetworkTimeSynchronizationSucceedsOrFails(expected string) error {

	var cmd string
	var extraBundleArgs string

	if !bundleEmbedded {
		extraBundleArgs = fmt.Sprintf("-b %s", bundleLocation)
	}
	crcStart := crcCmd.CRC("start").WithDisableNTP().ToString()
	cmd = fmt.Sprintf("%s -p '%s' %s", crcStart, pullSecretFile, extraBundleArgs)
	err := clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)

	return err
}

func StartCRCWithCustomBundleSucceedsOrFails(expected string) error {
	crcStart := crcCmd.CRC("start").ToString()
	cmd := fmt.Sprintf("%s -p '%s' -b *.crcbundle", crcStart, pullSecretFile)
	return clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)
}

func StartCRCWithDefaultBundleAndNameServerSucceedsOrFails(nameserver string, expected string) error {

	var extraBundleArgs string
	if !bundleEmbedded {
		extraBundleArgs = fmt.Sprintf("-b %s", bundleLocation)
	}

	crcStart := crcCmd.CRC("start").ToString()
	cmd := fmt.Sprintf("%s -n %s -p '%s' %s", crcStart, nameserver, pullSecretFile, extraBundleArgs)
	return clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)
}

func SetConfigPropertyToValueSucceedsOrFails(property string, value string, expected string) error {
	if value == "current bundle" {
		if bundleEmbedded {
			value = filepath.Join(CRCHome, "cache", bundleName)
		} else {
			value = bundleLocation
		}
	}
	return crcCmd.SetConfigPropertyToValueSucceedsOrFails(property, value, expected)
}

func ExecuteCommand(command string) error {
	return crcCmd.CRC(command).Execute()
}

func ExecuteCommandWithExpectedExitStatus(command string, expectedExitStatus string) error {
	return crcCmd.CRC(command).ExecuteWithExpectedExit(expectedExitStatus)
}
