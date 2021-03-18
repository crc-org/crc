package test_test

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
	"testing"

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

var credPath string
var userHome string
var versionInfo VersionAnswer

var bundlePath string
var pullSecretPath string

func TestTest(t *testing.T) {

	RegisterFailHandler(Fail)

	junitReporter := reporters.NewJUnitReporter(filepath.Join("out", "integration.xml"))
	RunSpecsWithDefaultAndCustomReporters(t, "Test Suite", []Reporter{junitReporter})

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
		bundlePath = os.Getenv("BUNDLE_PATH") // this env var should contain location of bundle
		if bundlePath == "" {
			logrus.Infof("Error: You need to set BUNDLE_PATH because your binary does not contain a bundle.")
			logrus.Infof("%v", err)
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(bundlePath).To(BeAnExistingFile()) // not checking if it's an actual bundle
	} else {
		bundlePath = "embedded"
	}

	// pull-secret location
	pullSecretPath = os.Getenv("PULL_SECRET_PATH") // this env var should contain location of pull-secret file
	if err != nil {
		logrus.Infof("Error: You need to set PULL_SECRET_PATH to find CRC useful.")
		logrus.Infof("%v", err)
		Expect(err).NotTo(HaveOccurred())
	}
	Expect(pullSecretPath).To(BeAnExistingFile()) // not checking if it's a valid pull secret file

})
