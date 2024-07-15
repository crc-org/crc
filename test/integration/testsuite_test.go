package test_test

import (
	"encoding/json"
	"flag"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/crc-org/crc/v2/test/extended/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegaformat "github.com/onsi/gomega/format"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type VersionAnswer struct {
	Version           string `json:"version"`
	Commit            string `json:"commit"`
	OpenshiftVersion  string `json:"openshiftVersion"`
	MicroshiftVersion string `json:"microshiftVersion"`
}

var credPath string
var userHome string
var versionInfo VersionAnswer

var bundlePath string
var pullSecretPath string

func TestMain(m *testing.M) {
	RegisterFlags(flag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	os.Exit(m.Run())
}

func RegisterFlags(flags *flag.FlagSet) {
	flags.StringVar(&bundlePath, "bundle-path", "", "Path to the bundle to be used in tests.")
	flags.StringVar(&pullSecretPath, "pull-secret-path", "", "Path to the file containing pull secret.")
}

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

	// start shell instance
	err = util.StartHostShellInstance("")
	Expect(err).NotTo(HaveOccurred())

	// set credPath
	credPath = filepath.Join(userHome, ".crc", "machines", "crc", "id_rsa")

	// find out if bundle embedded in the binary
	raw := RunCRCExpectSuccess("version", "-o", "json")
	err = json.Unmarshal([]byte(raw), &versionInfo)

	Expect(err).NotTo(HaveOccurred())

	if len(bundlePath) != 0 {
		Expect(bundlePath).To(BeAnExistingFile())
	}

	if len(pullSecretPath) == 0 {
		logrus.Infof("Error: You need to set PULL_SECRET_PATH for CRC to function properly.")
	}
	Expect(pullSecretPath).NotTo(BeEmpty())
})
