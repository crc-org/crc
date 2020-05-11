package crcsuite

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v10"

	clicumber "github.com/code-ready/clicumber/testsuite"
	"github.com/code-ready/crc/pkg/crc/oc"
)

var (
	CRCHome        string
	CRCExecutable  string
	bundleEmbedded bool
	bundleName     string
	bundleURL      string
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
	s.Step(`^with up to "(\d+)" retries with wait period of "(\d*(?:ms|s|m))" all cluster operators are running$`,
		CheckClusterOperatorsWithRetry)
	s.Step(`^with up to "(\d+)" retries with wait period of "(\d*(?:ms|s|m))" http response from "(.*)" has status code "(\d+)"$`,
		CheckHTTPResponseWithRetry)
	s.Step(`^with up to "(\d+)" retries with wait period of "(\d*(?:ms|s|m))" command "(.*)" output (should match|matches|should not match|does not match) "(.*)"$`,
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

	// Monitoring
	s.Step(`^preparing and recording the environment succeeds$`,
		PrepareAndRecordEnvironment)
	s.Step(`^taking snapshot of the node every "(\d*(?:ms|s|m))" exactly "(\d+)" times succeeds$`,
		TakeTemperatureRepeatedlyAtIntervals)
	s.Step(`^packaging and uploading data succeeds$`,
		PackageAndUpload)

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

		if bundleURL == "embedded" {
			fmt.Println("Expecting the bundle to be embedded in the CRC executable.")
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
			os.Exit(1)
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

	// nolint:staticcheck
	s.BeforeFeature(func(this *messages.GherkinDocument) {
		// copy data/config files to test dir
		err := CopyFilesToTestDir()
		if err != nil {
			os.Exit(1)
		}

		if !bundleEmbedded {
			if _, err := os.Stat(bundleName); err != nil {
				if !os.IsNotExist(err) {
					fmt.Printf("Unexpected error obtaining the bundle %v.\n", bundleName)
					os.Exit(1)
				}
				// Obtain the bundle to current dir
				fmt.Println("Obtaining bundle...")
				bundle, err := DownloadBundle(bundleURL, ".")
				if err != nil {
					fmt.Printf("Failed to obtain CRC bundle, %v\n", err)
					os.Exit(1)
				}
				fmt.Println("Using bundle:", bundle)
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
		s, err := cluster.GetClusterOperatorsStatus(ocConfig)
		if err != nil {
			return err
		}
		if s.Available {
			return nil
		}
		time.Sleep(retryDuration)
	}

	return fmt.Errorf("Some cluster operators are still not running")
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

func PackageAndUpload() error {

	curTime := time.Now()
	curDir, _ := os.Getwd()
	pathbase := path.Dir(curDir) // currently in the out/test-run

	dateStamp := curTime.Format("2006-01-02")
	dataDir := fmt.Sprintf("%s%s", "data_", dateStamp)
	dataDirPath := path.Join(pathbase, "test-results", dataDir)

	// round up files to upload to Github
	var files []string
	err := filepath.Walk(dataDirPath, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}

	for i, file := range files {
		if i == 0 { // skip the folder itself
			continue
		}
		filename := filepath.Base(file) // only take the filename without path
		err := CreateGithubFile(file, "crc-data", filepath.Join(dataDir, filename), "Commit message")
		if err != nil {
			fmt.Printf("Github upload failed: %s", err)
			return err
		}
	}

	return nil
}

func PrepareAndRecordEnvironment() error {

	curTime := time.Now()
	curDir, _ := os.Getwd()
	pathbase := path.Dir(curDir) // currently in the out/test-run
	dateStamp := curTime.Format("2006-01-02")
	dataDir := fmt.Sprintf("%s%s", "data_", dateStamp)
	newDir := path.Join(pathbase, "test-results", dataDir)
	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		os.Mkdir(newDir, 0777)
	}
	cmd_crcVersion := fmt.Sprintf("crc version > %s", path.Join(newDir, "crc_version.txt"))
	cmd_systemHardware := fmt.Sprintf("sudo lshw > %s", path.Join(newDir, "system_hardware.txt"))
	cmd_systemInfo := fmt.Sprintf("uname -a > %s", path.Join(newDir, "system_info.txt"))
	cmd_hypervisorVersion := fmt.Sprintf("virsh --version > %s", path.Join(newDir, "hypervisor_version.txt"))

	cmds := []string{cmd_crcVersion,
		cmd_systemHardware,
		cmd_systemInfo,
		cmd_hypervisorVersion}

	for _, cmd := range cmds {
		exec_err := clicumber.ExecuteCommand(cmd)
		if exec_err != nil {
			fmt.Println("Failed to execute command %s", cmd)
			return exec_err
		}
	}

	return nil
}

func TakeTemperatureRepeatedlyAtIntervals(offTime string, repeatCount int) error {

	waitDuration, err := time.ParseDuration(offTime)
	if err != nil {
		return err
	}

	curDir, _ := os.Getwd()
	pathbase := path.Dir(curDir) // currently in the out/test-run

	for i := 0; i < repeatCount; i++ {
		curTime := time.Now()
		timeStamp := curTime.Format("2006-01-02T15:04:05")
		dateStamp := curTime.Format("2006-01-02")
		dataDir := fmt.Sprintf("%s%s", "data_", dateStamp)
		dumpFile := fmt.Sprintf("%s%s%s", "node-describe_", timeStamp, ".txt")
		dumpFileLocation := path.Join(pathbase, "test-results", dataDir, dumpFile)

		cmd := fmt.Sprintf("oc describe node > %s", dumpFileLocation)

		exec_err := clicumber.ExecuteCommand(cmd)
		if exec_err != nil {
			fmt.Println("Failed to execute command %s", cmd)
			return exec_err
		}
		time.Sleep(waitDuration)
	}

	return nil
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

func StartCRCWithDefaultBundleSucceedsOrFails(expected string) error {

	var cmd string
	var extraBundleArgs string

	if !bundleEmbedded {
		extraBundleArgs = fmt.Sprintf("-b %s", bundleName)
	}
	cmd = fmt.Sprintf("crc start -p '%s' %s --log-level debug", pullSecretFile, extraBundleArgs)
	err := clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)

	return err
}

func StartCRCWithDefaultBundleWithStopNetworkTimeSynchronizationSucceedsOrFails(expected string) error {

	var cmd string
	var extraBundleArgs string

	if !bundleEmbedded {
		extraBundleArgs = fmt.Sprintf("-b %s", bundleName)
	}
	cmd = fmt.Sprintf("CRC_DEBUG_ENABLE_STOP_NTP=true crc start -p '%s' %s --log-level debug", pullSecretFile, extraBundleArgs)
	err := clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)

	return err
}

func StartCRCWithDefaultBundleAndNameServerSucceedsOrFails(nameserver string, expected string) error {

	var extraBundleArgs string
	if !bundleEmbedded {
		extraBundleArgs = fmt.Sprintf("-b %s", bundleName)
	}

	cmd := fmt.Sprintf("crc start -n %s -p '%s' %s --log-level debug", nameserver, pullSecretFile, extraBundleArgs)
	return clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)
}

func StdoutContainsIfBundleEmbeddedOrNot(value string, expected string) error {
	if expected == "is" { // expect embedded
		if bundleEmbedded { // really embedded
			return clicumber.CommandReturnShouldContain("stdout", value)
		}
		return clicumber.CommandReturnShouldNotContain("stdout", value)
	}
	// expect not embedded
	if !bundleEmbedded { // really not embedded
		return clicumber.CommandReturnShouldContain("stdout", value)
	}
	return clicumber.CommandReturnShouldNotContain("stdout", value)
}

func SetConfigPropertyToValueSucceedsOrFails(property string, value string, expected string) error {

	if value == "current bundle" {

		if bundleEmbedded {
			value = filepath.Join(CRCHome, "cache", bundleName)
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
