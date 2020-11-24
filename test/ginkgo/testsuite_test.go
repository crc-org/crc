// Before you run integration tests, you need to set
// PULL_SECRET_FILE environment variable to point to your
// pull-secret file. In case you are running a non-release binary
// (with standalone bundle), you also need to set BUNDLE_LOCATION
// variable to point to the location of the bundle you are using.

package test_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

type VersionAnswer struct {
	Version          string `json:"version"`
	Commit           string `json:"commit"`
	OpenshiftVersion string `json:"openshiftVersion"`
	Embedded         bool   `json:"embedded"`
}

type StatusAnswer struct {
	CRCStatus        string `json:"crcStatus"`
	OpenshiftStatus  string `json:"openshiftStatus"`
	OpenshiftVersion string `json:"openshiftVersion"`
	DiskUsage        int    `json:"diskUsage"`
	DiskSize         int    `json:"diskSize"`
	CacheUsage       int    `json:"cacheUsage"`
	CacheDir         string `json:"cacheDir"`
}

var credPath string
var userHome string
var versionInfo VersionAnswer

var bundleLocation string
var pullSecretLocation string

func TestTest(t *testing.T) {

	RegisterFailHandler(Fail)

	junitfile := fmt.Sprintf("junit_%s.xml", time.Now().Format("2006-01-02 3:4:5"))
	junitReporter := reporters.NewJUnitReporter(filepath.Join("results", junitfile))
	RunSpecsWithDefaultAndCustomReporters(t, "Test Suite", []Reporter{junitReporter})
	//RunSpecs(t, "Test Suite")
}

var _ = BeforeSuite(func() {

	// set userHome
	usr, err := user.Current()
	if err != nil {
		Expect(err).NotTo(HaveOccurred())
	}
	userHome = usr.HomeDir

	// set credPath
	credPath = filepath.Join(userHome, ".crc", "machines", "crc", "id_rsa")

	// find out if bundle embedded in the binary
	raw := RunCRCExpectSuccess("version", "-o", "json")
	err = json.Unmarshal([]byte(raw), &versionInfo)

	Expect(err).NotTo(HaveOccurred())

	// bundle location
	if !versionInfo.Embedded {
		bundleLocation = os.Getenv("BUNDLE_LOCATION") // this env var should contain location of bundle
		if bundleLocation == "" {
			logrus.Infof("Error: You need to set BUNDLE_LOCATION because your binary does not contain a bundle.")
			logrus.Infof("%s", err)
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(bundleLocation).To(BeAnExistingFile()) // not checking if it's an actual bundle
	} else {
		bundleLocation = "embedded"
	}

	// pull-secret location
	pullSecretLocation = os.Getenv("PULL_SECRET_FILE") // this env var should contain location of pull-secret file
	if err != nil {
		logrus.Infof("Error: You need to set PULL_SECRET_FILE to find CRC useful.")
		logrus.Infof("%s", err)
		Expect(err).NotTo(HaveOccurred())
	}
	Expect(pullSecretLocation).To(BeAnExistingFile()) // not checking if it's a valid pull secret file

})
