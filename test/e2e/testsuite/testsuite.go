package testsuite

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/containers/common/pkg/strongunits"
	"github.com/crc-org/crc/v2/pkg/crc/ssh"
	"github.com/spf13/cast"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/machine"
	"github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/crc/version"
	crcCmd "github.com/crc-org/crc/v2/test/extended/crc/cmd"
	"github.com/crc-org/crc/v2/test/extended/util"
	"github.com/cucumber/godog"
	"github.com/spf13/pflag"
)

var (
	CRCExecutable      string
	userProvidedBundle bool
	bundleName         string
	bundleLocation     string
	pullSecretFile     string
	cleanupHome        bool
	testWithShell      string
	CRCVersion         string
	CRCMemory          string

	GodogTags string
)

func defaultCRCVersion() string {
	return fmt.Sprintf("%s+%s", version.GetCRCVersion(), version.GetCommitSha())
}

func ParseFlags() {

	pflag.StringVar(&util.TestDir, "test-dir", "out", "Path to the directory in which to execute the tests")
	pflag.StringVar(&testWithShell, "test-shell", "", "Specifies shell to be used for the testing.")

	pflag.StringVar(&bundleLocation, "bundle-location", "", "Path to the bundle to be used in tests")
	pflag.StringVar(&pullSecretFile, "pull-secret-file", "/path/to/pull-secret", "Path to the file containing pull secret")
	pflag.StringVar(&CRCExecutable, "crc-binary", "", "Path to the CRC executable to be tested")
	pflag.BoolVar(&cleanupHome, "cleanup-home", false, "Try to remove crc home folder before starting the suite") // TODO: default=true
	pflag.StringVar(&CRCVersion, "crc-version", defaultCRCVersion(), "Version of CRC to be tested")
	pflag.StringVar(&CRCMemory, "crc-memory", "", "Memory for CRC VM in MiB")
}

func InitializeTestSuite(tctx *godog.TestSuiteContext) {

	tctx.BeforeSuite(func() {

		err := util.PrepareForE2eTest()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		usr, _ := user.Current()
		util.CRCHome = filepath.Join(usr.HomeDir, ".crc")

		// init CRCExecutable if no location provided by user
		if CRCExecutable == "" {
			fmt.Println("Expecting the CRC executable to be in $HOME/go/bin.")
			usr, _ := user.Current()
			CRCExecutable = filepath.Join(usr.HomeDir, "go", "bin")
		}

		// Force debug logs
		err = os.Setenv("CRC_LOG_LEVEL", "debug")
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
			userProvidedBundle = false
			bundleName = constants.GetDefaultBundle(preset.OpenShift)
		} else {
			fmt.Println("Expecting the bundle provided by the user")
			userProvidedBundle = true
			_, bundleName = filepath.Split(bundleLocation)
		}

		if pullSecretFile == "" {
			fmt.Println("User must specify the pull secret file via --pull-secret-file flag.")
			os.Exit(1)
		}

		if cleanupHome {
			// remove $HOME/.crc
			err = util.RemoveCRCHome()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		if userProvidedBundle {
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

	tctx.AfterSuite(func() {

		err := crcCmd.DeleteCRC()
		if err != nil {
			fmt.Printf("Could not delete CRC VM: %s.", err)
		}

		err = util.LogMessage("info", "----- Cleaning Up -----")
		if err != nil {
			fmt.Println("error logging:", err)
		}

		err = util.CloseLog()
		if err != nil {
			fmt.Println("Error closing the log:", err)
		}
	})
}

func InitializeScenario(s *godog.ScenarioContext) {

	s.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {

		err := util.StartHostShellInstance(testWithShell)
		if err != nil {
			fmt.Println("error starting host shell instance:", err)
		}
		util.ClearScenarioVariables()

		err = util.CleanTestRunDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = util.LogMessage("info", fmt.Sprintf("----- Scenario: %s -----", sc.Name))
		if err != nil {
			fmt.Println("error logging:", err)
		}
		err = util.LogMessage("info", fmt.Sprintf("----- Scenario Outline: %s -----", sc.Name))
		if err != nil {
			fmt.Println("error logging:", err)
		}

		for _, tag := range sc.Tags {

			// copy data/config files to test dir
			if tag.Name == "@testdata" {
				err := util.CopyFilesToTestDir()
				if err != nil {
					os.Exit(1)
				}
			}

			// move host's date 13 months forward and turn timesync off
			if tag.Name == "@timesync" {
				err := util.ExecuteCommand("sudo timedatectl set-ntp off")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				err = util.ExecuteCommand("sudo date -s '13 month'")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				err = util.ExecuteCommandWithRetry(10, "1s", "virsh --readonly -c qemu:///system capabilities", "contains", "<capabilities>")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}

			if tag.Name == "@proxy" {
				// start container with squid proxy
				err := util.ExecuteCommand("podman run --name squid -d -p 3128:3128 quay.io/crcont/squid")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				err = util.ExecuteCommand("crc config set http-proxy http://host.crc.testing:3128")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				err = util.ExecuteCommand("crc config set https-proxy http://host.crc.testing:3128")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				err = util.ExecuteCommand("crc config set no-proxy .testing")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				err = util.ExecuteCommand("crc config set host-network-access true")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

			}

			if tag.Name == "@system_network" {
				err = util.ExecuteCommand("crc config set network-mode system")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}

		return ctx, nil
	})

	s.After(func(ctx context.Context, sc *godog.Scenario, _ error) (context.Context, error) {

		for _, tag := range sc.Tags {

			// delete testproj namespace after a Scenario that used it
			if tag.Name == "@needs_namespace" {
				err := util.ExecuteCommand("oc delete namespace testproj")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}

			// move host's date 13 months back and turn timesync on
			if tag.Name == "@timesync" {
				err := util.ExecuteCommand("sudo date -s '-13 month'")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				err = util.ExecuteCommand("sudo timedatectl set-ntp on")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}

			if tag.Name == "@cleanup" {

				// CRC instance cleanup
				err := util.ExecuteCommand("crc cleanup")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				// CRC config cleanup
				err = crcCmd.UnsetConfigPropertySucceedsOrFails("enable-cluster-monitoring", "succeeds") // unsetting property that is not set gives 0 exitcode, so this works
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				err = crcCmd.UnsetConfigPropertySucceedsOrFails("memory", "succeeds") // unsetting property that is not set gives 0 exitcode, so this works
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				err = crcCmd.UnsetConfigPropertySucceedsOrFails("preset", "succeeds") // unsetting property that is not set gives 0 exitcode, so this works
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				if runtime.GOOS == "linux" {
					err = crcCmd.UnsetConfigPropertySucceedsOrFails("network-mode", "succeeds") // unsetting property that is not set gives 0 exitcode, so this works
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
				}
			}

			if tag.Name == "@proxy" {

				err := util.ExecuteCommand("podman stop squid")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				err = util.ExecuteCommand("podman rm squid")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				err = crcCmd.UnsetConfigPropertySucceedsOrFails("http-proxy", "succeeds") // unsetting property that is not set gives 0 exitcode, so this works
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				err = crcCmd.UnsetConfigPropertySucceedsOrFails("https-proxy", "succeeds") // unsetting property that is not set gives 0 exitcode, so this works
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				err = crcCmd.UnsetConfigPropertySucceedsOrFails("no-proxy", "succeeds") // unsetting property that is not set gives 0 exitcode, so this works
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				// crc oc-env sets these three quietly if http(s)-proxy is set in crc config
				if err := os.Unsetenv("HTTP_PROXY"); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				if err := os.Unsetenv("HTTPS_PROXY"); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				if err := os.Unsetenv("NO_PROXY"); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}

			if tag.Name == "@system_network" {
				err := util.ExecuteCommand("crc config unset network-mode")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}

		}

		return ctx, nil
	})

	s.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
		st.Text = util.ProcessScenarioVariables(st.Text)
		return ctx, nil
	})

	// Executing commands
	s.Step(`^executing "(.*)"$`,
		util.ExecuteCommand)
	s.Step(`^executing "(.*)" (succeeds|fails)$`,
		util.ExecuteCommandSucceedsOrFails)

	// Command output verification
	s.Step(`^(stdout|stderr|exitcode) (?:should contain|contains) "(.*)"$`,
		util.CommandReturnShouldContain)
	s.Step(`^(stdout|stderr|exitcode) (?:should contain|contains)$`,
		util.CommandReturnShouldContainContent)
	s.Step(`^(stdout|stderr|exitcode) (?:should|does) not contain "(.*)"$`,
		util.CommandReturnShouldNotContain)
	s.Step(`^(stdout|stderr|exitcode) (?:should|does not) contain$`,
		util.CommandReturnShouldNotContainContent)

	s.Step(`^(stdout|stderr|exitcode) (?:should equal|equals) "(.*)"$`,
		util.CommandReturnShouldEqual)
	s.Step(`^(stdout|stderr|exitcode) (?:should equal|equals)$`,
		util.CommandReturnShouldEqualContent)
	s.Step(`^(stdout|stderr|exitcode) (?:should|does) not equal "(.*)"$`,
		util.CommandReturnShouldNotEqual)
	s.Step(`^(stdout|stderr|exitcode) (?:should|does) not equal$`,
		util.CommandReturnShouldNotEqualContent)

	s.Step(`^(stdout|stderr|exitcode) (?:should match|matches) "(.*)"$`,
		util.CommandReturnShouldMatch)
	s.Step(`^(stdout|stderr|exitcode) (?:should match|matches)`,
		util.CommandReturnShouldMatchContent)
	s.Step(`^(stdout|stderr|exitcode) (?:should|does) not match "(.*)"$`,
		util.CommandReturnShouldNotMatch)
	s.Step(`^(stdout|stderr|exitcode) (?:should|does) not match`,
		util.CommandReturnShouldNotMatchContent)

	s.Step(`^(stdout|stderr|exitcode) (?:should be|is) empty$`,
		util.CommandReturnShouldBeEmpty)
	s.Step(`^(stdout|stderr|exitcode) (?:should not be|is not) empty$`,
		util.CommandReturnShouldNotBeEmpty)

	s.Step(`^(stdout|stderr|exitcode) (?:should be|is) valid "([^"]*)"$`,
		util.ShouldBeInValidFormat)

	// Command output and execution: extra steps
	s.Step(`^with up to "(\d*)" retries with wait period of "(\d*(?:ms|s|m))" command "(.*)" output (should contain|contains|should not contain|does not contain) "(.*)"$`,
		util.ExecuteCommandWithRetry)
	s.Step(`^evaluating stdout of the previous command succeeds$`,
		util.ExecuteStdoutLineByLine)

	// Scenario variables
	// allows to set a scenario variable to the output values of minishift and oc commands
	// and then refer to it by $(NAME_OF_VARIABLE) directly in the text of feature file
	s.Step(`^setting scenario variable "(.*)" to the stdout from executing "(.*)"$`,
		util.SetScenarioVariableExecutingCommand)

	// Filesystem operations
	s.Step(`^creating directory "([^"]*)" succeeds$`,
		util.CreateDirectory)
	s.Step(`^creating file "([^"]*)" succeeds$`,
		util.CreateFile)
	s.Step(`^deleting directory "([^"]*)" succeeds$`,
		util.DeleteDirectory)
	s.Step(`^deleting file "([^"]*)" succeeds$`,
		util.DeleteFile)
	s.Step(`^directory "([^"]*)" should not exist$`,
		util.DirectoryShouldNotExist)
	s.Step(`^file "([^"]*)" should not exist$`,
		util.FileShouldNotExist)
	s.Step(`^file "([^"]*)" exists$`,
		util.FileExist)
	s.Step(`^file from "(.*)" is downloaded into location "(.*)"$`,
		util.DownloadFileIntoLocation)
	s.Step(`^writing text "([^"]*)" to file "([^"]*)" succeeds$`,
		util.WriteToFile)
	s.Step(`^removing (openshift) bundle from cache succeeds$`,
		RemoveBundleFromCache)

	// File content checks
	s.Step(`^content of file "([^"]*)" should contain "([^"]*)"$`,
		util.FileContentShouldContain)
	s.Step(`^content of file "([^"]*)" should not contain "([^"]*)"$`,
		util.FileContentShouldNotContain)
	s.Step(`^content of file "([^"]*)" should equal "([^"]*)"$`,
		util.FileContentShouldEqual)
	s.Step(`^content of file "([^"]*)" should not equal "([^"]*)"$`,
		util.FileContentShouldNotEqual)
	s.Step(`^content of file "([^"]*)" should match "([^"]*)"$`,
		util.FileContentShouldMatchRegex)
	s.Step(`^content of file "([^"]*)" should not match "([^"]*)"$`,
		util.FileContentShouldNotMatchRegex)
	s.Step(`^content of file "([^"]*)" (?:should be|is) valid "([^"]*)"$`,
		util.FileContentIsInValidFormat)

	// Config file content, JSON and YAML
	s.Step(`"(JSON|YAML)" config file "(.*)" (contains|does not contain) key "(.*)" with value matching "(.*)"$`,
		util.ConfigFileContainsKeyMatchingValue)
	s.Step(`"(JSON|YAML)" config file "(.*)" (contains|does not contain) key "(.*)"$`,
		util.ConfigFileContainsKey)

	// CRC related steps
	s.Step(`^removing CRC home directory succeeds$`,
		util.RemoveCRCHome)
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
	s.Step(`^getting config property "(.*)" (succeeds|fails)$`,
		crcCmd.GetConfigPropertySucceedsOrFails)
	s.Step(`^unsetting config property "(.*)" (succeeds|fails)$`,
		crcCmd.UnsetConfigPropertySucceedsOrFails)
	s.Step(`^login to the oc cluster (succeeds|fails)$`,
		util.LoginToOcClusterSucceedsOrFails)
	s.Step(`^setting kubeconfig context to "(.*)" (succeeds|fails)$`,
		SetKubeConfigContextSucceedsOrFails)
	s.Step(`^with up to "(\d+)" retries with wait period of "(\d*(?:ms|s|m))" http response from "(.*)" has status code "(\d+)"$`,
		CheckHTTPResponseWithRetry)
	s.Step(`^with up to "(\d+)" retries with wait period of "(\d*(?:ms|s|m))" command "(.*)" output (should match|matches|should not match|does not match) "(.*)"$`,
		CheckOutputMatchWithRetry)
	s.Step(`^checking that CRC is (running|stopped)$`,
		CheckCRCStatus)
	s.Step(`^checking the CRC status JSON output is valid$`,
		CheckCRCStatusJSONOutput)
	s.Step(`^execut(?:e|ing) crc (.*) command$`,
		ExecuteCRCCommand)
	s.Step(`^execut(?:e|ing) crc (.*) command (.*)$`,
		ExecuteCommandWithExpectedExitStatus)
	s.Step(`^execut(?:e|ing) single crc (.*) command (.*)$`,
		ExecuteSingleCommandWithExpectedExitStatus)
	s.Step(`^execut(?:e|ing) podman command (.*) (succeeds|fails)$`,
		ExecutingPodmanCommandSucceedsFails)
	s.Step(`^ensuring CRC cluster is running$`,
		EnsureCRCIsRunning)
	s.Step(`^ensuring oc command is available$`,
		EnsureOCCommandIsAvailable)
	s.Step(`^ensuring user is logged in (succeeds|fails)`,
		EnsureUserIsLoggedIntoClusterSucceedsOrFails)
	s.Step(`^podman command is available$`,
		PodmanCommandIsAvailable)
	s.Step(`^deleting a pod (succeeds|fails)$`,
		DeletingPodSucceedsOrFails)
	s.Step(`^pulling image "(.*)", logging in, and pushing local image to internal registry succeeds$`,
		PullLoginTagPushImageSucceeds)

	// CRC file operations
	s.Step(`^file "([^"]*)" exists in CRC home folder$`,
		FileExistsInCRCHome)
	s.Step(`"(JSON|YAML)" config file "(.*)" in CRC home folder (contains|does not contain) key "(.*)" with value matching "(.*)"$`,
		ConfigFileInCRCHomeContainsKeyMatchingValue)
	s.Step(`"(JSON|YAML)" config file "(.*)" in CRC home folder (contains|does not contain) key "(.*)"$`,
		ConfigFileInCRCHomeContainsKey)
	s.Step(`removing file "(.*)" from CRC home folder succeeds$`,
		DeleteFileFromCRCHome)
	s.Step(`^decode base64 file "(.*)" to "(.*)"$`,
		DecodeBase64File)
	s.Step(`^ensuring network mode user$`,
		EnsureUserNetworkmode)
	s.Step(`^ensuring microshift cluster is fully operational$`,
		EnsureMicroshiftClusterIsOperational)
	s.Step(`^kubeconfig is cleaned up$`,
		EnsureKubeConfigIsCleanedUp)
	s.Step(`^crc version has expected output$`,
		EnsureCrcVersionIsCorrect)
	s.Step(`^ensure service "(.*)" is accessible via NodePort with response body "(.*)"$`,
		EnsureApplicationIsAccessibleViaNodePort)
	s.Step(`^persistent volume of size "([^"]*)"GB exists$`,
		EnsureVMPartitionSizeCorrect)

	s.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {

		if usingPreexistingCluster() {
			// collecting diagnostics data is quite slow, and they
			// are not really useful when running the tests locally
			// against an already running cluster
			return ctx, nil
		}
		if err != nil {
			if err := util.RunDiagnose(filepath.Join("..", "test-results")); err != nil {
				fmt.Printf("Failed to collect diagnostic: %v\n", err)
			}
		}

		err = util.CloseHostShellInstance()
		if err != nil {
			fmt.Println("error closing host shell instance:", err)
		}

		return ctx, nil
	})
}

func usingPreexistingCluster() bool {
	return strings.Contains(GodogTags, "~@startstop")
}

func WaitForClusterInState(state string) error {
	return crcCmd.WaitForClusterInState(state)
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
		execErr := util.ExecuteCommand(command)
		if execErr == nil {
			if strings.Contains(expected, " not ") {
				matchErr = util.CommandReturnShouldNotMatch("stdout", expectedOutput)
			} else {
				matchErr = util.CommandReturnShouldMatch("stdout", expectedOutput)
			}
			if matchErr == nil {
				return nil
			}
		}
		time.Sleep(retryDuration)
	}

	return matchErr
}

func EnsureCrcVersionIsCorrect() error {
	err := util.ExecuteCommand("crc version -ojson")
	if err != nil {
		return fmt.Errorf("could not execute 'crc version -ojson' command: %v", err)
	}
	crcVersionJSONOutput := util.GetLastCommandOutput("stdout")
	type CrcVersionOutput struct {
		Version           string `json:"version"`
		Commit            string `json:"commit"`
		OpenShiftVersion  string `json:"openshiftVersion"`
		MicroShiftVersion string `json:"microshiftVersion"`
	}
	var crcVersionOutput CrcVersionOutput
	err = json.Unmarshal([]byte(crcVersionJSONOutput), &crcVersionOutput)
	if err != nil {
		return fmt.Errorf("error in unmarshalling crc version output json: %v", err)
	}
	if crcVersionOutput.Version != version.GetCRCVersion() {
		_, err := time.Parse("06.01.02", crcVersionOutput.Version)
		if err != nil {
			return fmt.Errorf("crc version doesn't match, expected '%s', actual '%s'", version.GetCRCVersion(), crcVersionOutput.Version)
		}
	}
	if crcVersionOutput.Commit != version.GetCommitSha() {
		return fmt.Errorf("crc version commit sha don't match, expected '%s', actual '%s'", version.GetCommitSha(), crcVersionOutput.Commit)
	}
	if len(crcVersionOutput.OpenShiftVersion) == 0 {
		return fmt.Errorf("expected OpenShift version to be set in crc version output")
	}
	if len(crcVersionOutput.MicroShiftVersion) == 0 {
		return fmt.Errorf("expected MicroShift version to be set in crc version output")
	}
	return nil
}

func CheckCRCStatus(state string) error {
	if state == "running" {
		// crc start can finish successfully, even when
		// status for cluster is still starting. It is expected
		// the cluster got stabilized at most within 10 minutes
		return crcCmd.WaitForClusterInState(state)
	}
	return crcCmd.CheckCRCStatus(state)
}

func CheckCRCStatusJSONOutput() error {
	err := util.ExecuteCommand("crc status -ojson")
	if err != nil {
		return err
	}
	crcStatusJSONOutput := util.GetLastCommandOutput("stdout")
	var crcStatusJSONOutputObj map[string]interface{}
	if err := json.Unmarshal([]byte(crcStatusJSONOutput), &crcStatusJSONOutputObj); err != nil {
		return err
	}
	crcStatusSuccess := crcStatusJSONOutputObj["success"]
	if crcStatusSuccess != true {
		return fmt.Errorf("failure in asserting 'success' field of crc status json output, expected : true, actual : %t", crcStatusSuccess)
	}
	crcStatus := crcStatusJSONOutputObj["crcStatus"]
	if crcStatus != "Running" {
		return fmt.Errorf("failure in asserting 'crcStatus' field of crc status json output, expected : 'Running', actual : %s", crcStatus)
	}
	crcCacheDir := crcStatusJSONOutputObj["cacheDir"]
	if crcCacheDir != filepath.Join(util.CRCHome, "cache") {
		return fmt.Errorf("failure is asserting 'cacheDir' field of crc status json output, expected : %s, actual %s", filepath.Join(util.CRCHome, "cache"), crcCacheDir)
	}
	crcOpenShiftStatus := crcStatusJSONOutputObj["openshiftStatus"]
	if crcOpenShiftStatus != "Running" {
		return fmt.Errorf("failure in asserting 'openshiftStatus' field of crc status json output, expected : 'Running', actual : %s", crcOpenShiftStatus)
	}
	crcOpenShiftVersion := crcStatusJSONOutputObj["openshiftVersion"]
	if !strings.HasPrefix(cast.ToString(crcOpenShiftVersion), "4.") {
		return fmt.Errorf("failure in asserting 'openshiftVersion' field of crc status json output, expected with prefix: '4.', actual : %s", crcOpenShiftVersion)
	}
	crcPresetStr := crcStatusJSONOutputObj["preset"]
	crcPreset, err := preset.ParsePresetE(crcPresetStr.(string))
	if err != nil {
		return fmt.Errorf("failure in asserting 'preset' field of crc status json output, %v", err)
	}
	crcDiskSize := crcStatusJSONOutputObj["diskSize"]
	if strongunits.B(cast.ToUint64(crcDiskSize)) > strongunits.GiB(constants.DefaultDiskSize).ToBytes() {
		return fmt.Errorf("failure in asserting 'diskSize' field of crc status json output, expected less than or equal to %d bytes, actual : %d bytes", strongunits.GiB(constants.DefaultDiskSize).ToBytes(), strongunits.B(cast.ToUint64(crcDiskSize)))
	}
	crcRAMSize := crcStatusJSONOutputObj["ramSize"]
	if strongunits.B(cast.ToUint64(crcRAMSize)) > constants.GetDefaultMemory(crcPreset).ToBytes() {
		return fmt.Errorf("failure in asserting 'ramSize' field of crc status json output, expected less than or equal to %d bytes, actual : %d bytes", constants.GetDefaultMemory(crcPreset).ToBytes(), cast.ToUint64(crcRAMSize))
	}
	return nil
}

func DeleteFileFromCRCHome(fileName string) error {

	theFile := filepath.Join(util.CRCHome, fileName)

	if _, err := os.Stat(theFile); os.IsNotExist(err) {
		return nil
	}

	if err := util.DeleteFile(theFile); err != nil {
		return fmt.Errorf("error deleting file %v", theFile)
	}
	return nil
}

func FileExistsInCRCHome(fileName string) error {

	theFile := filepath.Join(util.CRCHome, fileName)

	_, err := os.Stat(theFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exists, error: %v ", theFile, err)
	}

	return err
}

func RemoveBundleFromCache() error {

	var p = preset.OpenShift

	theBundle := util.GetBundlePath(p)
	theFolder := strings.TrimSuffix(theBundle, ".crcbundle")

	// remove the unpacked folder (if present)
	err := os.RemoveAll(theFolder)
	if err != nil {
		return err
	}

	// remove the bundle file (if present)
	err = os.RemoveAll(theBundle)

	if err != nil {
		return err
	}

	return nil
}

func ConfigFileInCRCHomeContainsKeyMatchingValue(format string, configFile string, condition string, keyPath string, expectedValue string) error {

	if expectedValue == "current bundle" {
		if !userProvidedBundle {
			return ConfigFileInCRCHomeContainsKey("JSON", "crc.json", "does not contain", "bundle")
		}
		expectedValue = fmt.Sprintf(".*%s", bundleName)
	}
	configPath := filepath.Join(util.CRCHome, configFile)

	config, err := util.GetFileContent(configPath)
	if err != nil {
		return err
	}

	keyValue, err := util.GetConfigKeyValue([]byte(config), format, keyPath)
	if err != nil {
		return err
	}

	matches, err := util.PerformRegexMatch(expectedValue, keyValue)
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

	configPath := filepath.Join(util.CRCHome, configFile)

	config, err := util.GetFileContent(configPath)
	if err != nil {
		return err
	}

	keyValue, err := util.GetConfigKeyValue([]byte(config), format, keyPath)
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

func SetKubeConfigContextSucceedsOrFails(context, expected string) error {
	cmd := fmt.Sprintf("oc config use-context %s", context)
	return util.ExecuteCommandSucceedsOrFails(cmd, expected)
}

func StartCRCWithDefaultBundleSucceedsOrFails(expected string) error {

	var cmd string
	var extraBundleArgs string

	if userProvidedBundle {
		extraBundleArgs = fmt.Sprintf("-b %s", bundleLocation)
	}
	crcStart := crcCmd.CRC("start").ToString()
	cmd = fmt.Sprintf("%s -p '%s' %s", crcStart, pullSecretFile, extraBundleArgs)
	err := util.ExecuteCommandSucceedsOrFails(cmd, expected)

	return err
}

func StartCRCWithDefaultBundleWithStopNetworkTimeSynchronizationSucceedsOrFails(expected string) error {

	var cmd string
	var extraBundleArgs string

	if userProvidedBundle {
		extraBundleArgs = fmt.Sprintf("-b %s", bundleLocation)
	}
	crcStart := crcCmd.CRC("start").WithDisableNTP().ToString()
	cmd = fmt.Sprintf("%s -p '%s' %s", crcStart, pullSecretFile, extraBundleArgs)
	err := util.ExecuteCommandSucceedsOrFails(cmd, expected)

	return err
}

func StartCRCWithCustomBundleSucceedsOrFails(expected string) error {
	crcStart := crcCmd.CRC("start").ToString()
	cmd := fmt.Sprintf("%s -p '%s' -b *.crcbundle", crcStart, pullSecretFile)
	return util.ExecuteCommandSucceedsOrFails(cmd, expected)
}

func StartCRCWithDefaultBundleAndNameServerSucceedsOrFails(nameserver string, expected string) error {

	var extraBundleArgs string
	if userProvidedBundle {
		extraBundleArgs = fmt.Sprintf("-b %s", bundleLocation)
	}

	crcStart := crcCmd.CRC("start").ToString()
	cmd := fmt.Sprintf("%s -n %s -p '%s' %s", crcStart, nameserver, pullSecretFile, extraBundleArgs)
	return util.ExecuteCommandSucceedsOrFails(cmd, expected)
}

func EnsureCRCIsRunning() error {
	if usingPreexistingCluster() {
		return nil
	}

	err := crcCmd.CheckCRCStatus("running")

	// if cluster is not in a Running state
	if err != nil {
		// make sure cluster doesn't exist in unexpected state
		err = ExecuteCRCCommand("cleanup")
		if err != nil {
			return err
		}

		// set up and start the cluster with lots of memory, if specified
		if CRCMemory != "" {
			err = SetConfigPropertyToValueSucceedsOrFails("memory", CRCMemory, "succeeds")
			if err != nil {
				return err
			}
		}

		err = ExecuteSingleCommandWithExpectedExitStatus("setup", "succeeds") // uses the right bundle argument if needed
		if err != nil {
			return err
		}

		if runtime.GOOS == "windows" {
			err = StartCRCWithDefaultBundleAndNameServerSucceedsOrFails("10.75.5.25", "succeeds")
		} else {
			err = StartCRCWithDefaultBundleSucceedsOrFails("succeeds")
		}
		if err != nil {
			return err
		}

		// We're not testing if the cluster comes up fast enough, just need it Running
		err = crcCmd.WaitForClusterInState("running")
		if err != nil {
			err = crcCmd.WaitForClusterInState("running")
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func EnsureUserIsLoggedIntoClusterSucceedsOrFails(expected string) error {

	if err := setOcEnv(); err != nil {
		return err
	}

	err := util.LoginToOcCluster([]string{})
	if expected == "succeeds" && err != nil && strings.Contains(err.Error(), "The server uses a certificate signed by unknown authority") {
		// do some logging

		err1 := util.ExecuteCommand("oc config view --raw -o jsonpath=\"{.clusters[?(@.name=='api-crc-testing:6443')].cluster.certificate-authority-data}\" > ca.base64")
		if err1 != nil {
			fmt.Println(err1)
		}
		err1 = DecodeBase64File("ca.base64", "ca.crt")
		if err1 != nil {
			fmt.Println(err1)
		}
		err1 = util.ExecuteCommand("echo QUIT | openssl s_client -connect api.crc.testing:6443 | openssl x509 -out server.crt")
		if err1 != nil {
			fmt.Println(err1)
		}
		err1 = util.ExecuteCommand("openssl verify -CAfile ca.crt server.crt")
		if err1 != nil {
			fmt.Println(err1)
		}

		// login with ignorance
		err = util.LoginToOcCluster([]string{"--insecure-skip-tls-verify"})
	}
	return err
}

func EnsureOCCommandIsAvailable() error {
	err := setOcEnv()
	if err != nil {
		return err
	}
	return setPodmanEnv()
}

func setOcEnv() error {
	if runtime.GOOS == "windows" {
		return util.ExecuteCommandSucceedsOrFails("crc oc-env | Invoke-Expression", "succeeds")
	}
	return util.ExecuteCommandSucceedsOrFails("eval $(crc oc-env)", "succeeds")
}

func setPodmanEnv() error {
	if runtime.GOOS == "windows" {
		return util.ExecuteCommandSucceedsOrFails("crc podman-env | Invoke-Expression", "succeeds")
	}
	return util.ExecuteCommandSucceedsOrFails("eval $(crc podman-env)", "succeeds")
}

func SetConfigPropertyToValueSucceedsOrFails(property string, value string, expected string) error {
	// Since network-mode is only supported on Linux, we skip this property test for non-linux platforms
	if property == "network-mode" && runtime.GOOS != "linux" {
		return nil
	}
	if value == "current bundle" {
		if !userProvidedBundle {
			value = filepath.Join(util.CRCHome, "cache", bundleName)
		} else {
			value = bundleLocation
		}
	}
	return crcCmd.SetConfigPropertyToValueSucceedsOrFails(property, value, expected)
}

func ExecuteCRCCommand(command string) error {
	return crcCmd.CRC(command).Execute()
}

func ExecuteCommandWithExpectedExitStatus(command string, expectedExitStatus string) error {
	if command == "setup" && userProvidedBundle {
		command = fmt.Sprintf("%s -b %s", command, bundleLocation)
	}
	return crcCmd.CRC(command).ExecuteWithExpectedExit(expectedExitStatus)
}

func ExecuteSingleCommandWithExpectedExitStatus(command string, expectedExitStatus string) error {
	if command == "setup" && userProvidedBundle {
		command = fmt.Sprintf("%s -b %s", command, bundleLocation)
	}
	return crcCmd.CRC(command).ExecuteSingleWithExpectedExit(expectedExitStatus)
}

func DeletingPodSucceedsOrFails(expected string) error {
	var err error
	if runtime.GOOS == "windows" {
		_ = util.ExecuteCommandSucceedsOrFails("$Env:POD = $(oc get pod -o jsonpath=\"{.items[0].metadata.name}\")", expected)
		err = util.ExecuteCommandSucceedsOrFails("oc delete pod $Env:POD --now", expected)
	} else {
		_ = util.ExecuteCommandSucceedsOrFails("POD=$(oc get pod -o jsonpath=\"{.items[0].metadata.name}\")", expected)
		err = util.ExecuteCommandSucceedsOrFails("oc delete pod $POD --now", expected)
	}
	return err
}

func PodmanCommandIsAvailable() error {

	// Do what 'eval $(crc podman-env) would do
	path := os.ExpandEnv("${HOME}/.crc/bin/podman:$PATH")
	csshk := os.ExpandEnv("${HOME}/.crc/machines/crc/id_ed25519")
	dh := os.ExpandEnv("unix:///${HOME}/.crc/machines/crc/docker.sock")
	ch := "ssh://core@127.0.0.1:2222/run/user/1000/podman/podman.sock"
	if runtime.GOOS == "windows" {
		userHomeDir, _ := os.UserHomeDir()
		unexpandedPath := filepath.Join(userHomeDir, ".crc/bin/podman;${PATH}")
		path = os.ExpandEnv(unexpandedPath)
		csshk = filepath.Join(userHomeDir, ".crc/machines/crc/id_ed25519")
		dh = "npipe:////./pipe/crc-podman"
	}

	os.Setenv("PATH", path)
	os.Setenv("CONTAINER_SSHKEY", csshk)
	os.Setenv("CONTAINER_HOST", ch)
	os.Setenv("DOCKER_HOST", dh)

	return nil

}

func ExecutingPodmanCommandSucceedsFails(command string, expected string) error {

	var err error
	switch expected {
	case "succeeds":
		_, err = crcCmd.RunPodmanExpectSuccess(strings.Split(command[1:len(command)-1], " ")...)
	case "fails":
		_, err = crcCmd.RunPodmanExpectFail(strings.Split(command[1:len(command)-1], " ")...)
	}

	return err
}

func PullLoginTagPushImageSucceeds(image string) error {
	_, err := crcCmd.RunPodmanExpectSuccess("pull", image)
	if err != nil {
		return err
	}

	err = util.ExecuteCommand("oc whoami -t")
	if err != nil {
		return err
	}

	token := util.GetLastCommandOutput("stdout")
	fmt.Println(token)
	_, err = crcCmd.RunPodmanExpectSuccess("login", "-u", "kubeadmin", "-p", token, "default-route-openshift-image-registry.apps-crc.testing", "--tls-verify=false") // $(oc whoami -t)
	if err != nil {
		return err
	}

	_, err = crcCmd.RunPodmanExpectSuccess("tag", image, "default-route-openshift-image-registry.apps-crc.testing/testproj/hello:test")
	if err != nil {
		return err
	}

	_, err = crcCmd.RunPodmanExpectSuccess("push", "default-route-openshift-image-registry.apps-crc.testing/testproj/hello:test", "--remove-signatures", "--tls-verify=false")
	if err != nil {
		return err
	}

	return nil
}

// Decode a file encoded with base64
func DecodeBase64File(inputFile, outputFile string) error {
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = fmt.Sprintf("certutil.exe -decode %s %s", inputFile, outputFile)
	} else {
		cmd = fmt.Sprintf("base64 -d -i %s > %s", inputFile, outputFile)
	}
	return util.ExecuteCommandSucceedsOrFails(cmd, "succeeds")
}

func EnsureUserNetworkmode() error {
	if runtime.GOOS == "linux" {
		return crcCmd.SetConfigPropertyToValueSucceedsOrFails(
			"network-mode", "user", "succeeds")
	}
	return nil
}

func EnsureKubeConfigIsCleanedUp() error {
	kubeConfig, cfg, err := machine.GetGlobalKubeConfig()
	if err != nil {
		return err
	}
	if len(kubeConfig) == 0 {
		return fmt.Errorf("unable to load kube config file while verifying kube config cleanup")
	}
	if cfg.CurrentContext != "" {
		return fmt.Errorf("kube config's current context not cleaned up. [expected : \"\", actual : %s]", cfg.CurrentContext)
	}
	crcClusterDomain := fmt.Sprintf("https://api%s:6443", constants.ClusterDomain)
	for name, cluster := range cfg.Clusters {
		if cluster.Server == crcClusterDomain {
			return fmt.Errorf("kube config's cluster %s is not cleaned up, it still contains a cluster with %s domain [expected : \"\", actual : %s]", name, crcClusterDomain, crcClusterDomain)
		}
	}
	return nil
}

func EnsureApplicationIsAccessibleViaNodePort(svcName string, expectedResponseBody string) error {
	crcIPCmdErr := util.ExecuteCommand("crc ip")
	if crcIPCmdErr != nil {
		return fmt.Errorf("crc ip command failed: %s", crcIPCmdErr.Error())
	}
	crcIP := util.GetLastCommandOutput("stdout")
	serviceNodePortErr := util.ExecuteCommand(fmt.Sprintf("oc get svc %s -o jsonpath='{.spec.ports[0].nodePort}'", svcName))
	if serviceNodePortErr != nil {
		return fmt.Errorf("oc get svc command failed: %s", serviceNodePortErr.Error())
	}
	serviceNodePort := util.GetLastCommandOutput("stdout")
	applicationURL := fmt.Sprintf("http://%s:%s", crcIP, serviceNodePort)
	req, err := http.NewRequest(http.MethodGet, applicationURL, nil)
	if err != nil {
		return fmt.Errorf("unable to create http request for: %s, %s", applicationURL, err.Error())
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to access application via NodePort Url: %s", applicationURL)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unable to access application via NodePort Url: %s, expected response code 200, got %d", applicationURL, resp.StatusCode)
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to access application via NodePort Url: %s", applicationURL)
	}
	if string(responseBody) != expectedResponseBody {
		return fmt.Errorf("unexpected response from url : %s, expected : %s, actual : %s", applicationURL, expectedResponseBody, string(responseBody))
	}
	return nil
}

// This function will wait until the microshift cluster got operational
func EnsureMicroshiftClusterIsOperational() error {
	// First wait until crc report the cluster as running
	err := crcCmd.WaitForClusterInState("running")
	if err != nil {
		return err
	}
	// Define the services to declare the cluster operational
	services := map[string]string{
		".*dns-default.*2/2.*Running.*":    "oc get pods -n openshift-dns",
		".*ovnkube-master.*4/4.*Running.*": "oc get pods -n openshift-ovn-kubernetes",
		".*ovnkube-node.*1/1.*Running.*":   "oc get pods -n openshift-ovn-kubernetes"}

	for operationalState, getPodCommand := range services {
		var operational = false
		for !operational {
			if err := util.ExecuteCommandSucceedsOrFails(getPodCommand, "succeeds"); err != nil {
				return err
			}
			operational = (nil == util.CommandReturnShouldMatch(
				"stdout",
				operationalState))
		}
	}

	return nil
}

func EnsureVMPartitionSizeCorrect(expectedPVSizeStr string) error {
	expectedPVSize, err := strconv.Atoi(expectedPVSizeStr)
	if err != nil {
		return fmt.Errorf("invalid expected persistent volume size provided in test input")
	}
	err = util.ExecuteCommand("crc ip")
	if err != nil {
		return fmt.Errorf("error in determining crc vm's ip address: %v", err)
	}
	crcIP := util.GetLastCommandOutput("stdout")
	runner, err := ssh.CreateRunner(crcIP, 2222, filepath.Join(util.CRCHome, "machines", "crc", "id_ed25519"))
	if err != nil {
		return fmt.Errorf("error creating ssh runner: %v", err)
	}
	out, _, err := runner.Run("lsblk -oTYPE,SIZE -n")
	if err != nil {
		return fmt.Errorf("error in executing command in crc vm: %v", err)
	}

	actualPVSize, err := deserializeListBlockDeviceCommandOutputToExtractPVSize(out)
	if err != nil {
		return err
	}
	if actualPVSize != expectedPVSize {
		return fmt.Errorf("expecting persistent volume size to be %d, got %d", expectedPVSize, actualPVSize)
	}
	return nil
}

func deserializeListBlockDeviceCommandOutputToExtractPVSize(lsblkOutput string) (int, error) {
	type BlockDevice struct {
		DeviceType string
		Size       string
	}
	blockDevices := make([]BlockDevice, 0)
	lines := strings.Split(lsblkOutput, "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		blockDevices = append(blockDevices, BlockDevice{
			DeviceType: fields[0],
			Size:       fields[1],
		})
	}

	var lvmSize int
	lvmBlockDeviceIndex := slices.IndexFunc(blockDevices, func(b BlockDevice) bool {
		return b.DeviceType == "lvm"
	})
	if lvmBlockDeviceIndex == -1 {
		return -1, fmt.Errorf("expecting lsblk output to contain a lvm device, got no device with type lvm")
	}
	_, err := fmt.Sscanf(blockDevices[lvmBlockDeviceIndex].Size, "%dG", &lvmSize)
	if err != nil {
		return -1, fmt.Errorf("error in scanning lvm device size: %v", err)
	}

	var diskSize = math.MinInt64
	for _, blockDevice := range blockDevices {
		if blockDevice.DeviceType == "disk" {
			diskSizeValue, err := strconv.ParseFloat(strings.TrimSuffix(blockDevice.Size, "G"), 64)
			if err != nil {
				return -1, fmt.Errorf("error in parsing disk size: %v", err)
			}
			if int(diskSizeValue) > diskSize {
				diskSize = int(diskSizeValue)
			}
		}
	}
	return diskSize - (lvmSize + 1), nil
}
