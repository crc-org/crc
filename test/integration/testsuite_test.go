package test_test

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/crc-org/crc/v2/test/extended/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegaformat "github.com/onsi/gomega/format"
	"github.com/sirupsen/logrus"
)

type VersionAnswer struct {
	Version          string `json:"version"`
	Commit           string `json:"commit"`
	OpenshiftVersion string `json:"openshiftVersion"`
}

var credPath string
var userHome string
var versionInfo VersionAnswer

var bundlePath string
var ginkgoOpts string
var pullSecretPath string

func TestTest(t *testing.T) {

	RegisterFailHandler(Fail)

	// disable error/output strings truncation
	gomegaformat.MaxLength = 0

	// fetch the current (reporter) config
	suiteConfig, reporterConfig := GinkgoConfiguration()

	err := os.MkdirAll("out", 0775)
	if err != nil {
		logrus.Infof("failed to create directory: %v", err)
	}
	reporterConfig.JUnitReport = filepath.Join("out", "integration.xml")

	RunSpecs(t, "Integration", suiteConfig, reporterConfig)

}

var _ = BeforeSuite(func() {

	// set userHome
	usr, err := user.Current()
	Expect(err).NotTo(HaveOccurred())
	userHome = usr.HomeDir
	util.CRCHome = filepath.Join(userHome, ".crc")

	// cleanup CRC
	Expect(RunCRCExpectSuccess("cleanup")).To(MatchRegexp("Cleanup finished"))

	// remove config file crc.json
	err = util.RemoveCRCConfig()
	Expect(err).NotTo(HaveOccurred())

	// set credPath
	credPath = filepath.Join(userHome, ".crc", "machines", "crc", "id_rsa")

	// find out if bundle embedded in the binary
	raw := RunCRCExpectSuccess("version", "-o", "json")
	err = json.Unmarshal([]byte(raw), &versionInfo)

	Expect(err).NotTo(HaveOccurred())

	bundlePath = os.Getenv("BUNDLE_PATH") // this env var should contain location of bundle
	if bundlePath != "" {
		Expect(bundlePath).To(BeAnExistingFile())
	}

	ginkgoOpts = os.Getenv("GINKGO_OPTS")
	if err != nil {

		logrus.Infof("Error: Could not read GINKGO_OPTS.")
		logrus.Infof("%v", err)
	}
	Expect(err).NotTo(HaveOccurred())

	pullSecretPath = os.Getenv("PULL_SECRET_PATH") // this env var should contain location of pull-secret file
	if err != nil {

		logrus.Infof("Error: You need to set PULL_SECRET_PATH to find CRC useful.")
		logrus.Infof("%v", err)
	}
	Expect(err).NotTo(HaveOccurred())

})
