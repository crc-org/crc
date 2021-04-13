package integration_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

type versionAnswer struct {
	Embedded bool `json:"embedded"`
}

var (
	userHome       string
	bundlePath     string
	pullSecretPath string
)

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)

	junitReporter := reporters.NewJUnitReporter(filepath.Join("out", "integration.xml"))
	RunSpecsWithDefaultAndCustomReporters(t, "Test Suite", []Reporter{junitReporter})
}

var _ = BeforeSuite(func() {
	usr, err := user.Current()
	Expect(err).NotTo(HaveOccurred())
	userHome = usr.HomeDir

	// find out if bundle embedded in the binary
	raw := RunCRCExpectSuccess("version", "-o", "json")

	var versionInfo versionAnswer
	Expect(json.Unmarshal([]byte(raw), &versionInfo)).NotTo(HaveOccurred())

	Expect(ioutil.WriteFile(filepath.Join("out", "crc-version.json"), []byte(raw), 0600)).To(Succeed())

	// bundle location
	bundlePath = "embedded"
	if !versionInfo.Embedded {
		bundlePath = os.Getenv("BUNDLE_PATH")
		if bundlePath != "embedded" {
			Expect(bundlePath).To(BeAnExistingFile(), "$BUNDLE_PATH should be a valid .crcbundle file")
		}
	}

	// pull-secret location
	pullSecretPath = os.Getenv("PULL_SECRET_PATH")
	Expect(pullSecretPath).To(BeAnExistingFile(), "$PULL_SECRET_PATH should be a pull secret file")
})

func BeforeAll(fn func()) {
	first := true
	BeforeEach(func() {
		if first {
			fn()
			first = false
		}
	})
}

func cleanUp() func() {
	return func() {
		_, _ = NewCRCCommand("delete", "-f").Exec()
		_, _ = NewCRCCommand("cleanup").Exec()
		_ = os.Remove(filepath.Join(userHome, ".crc", "crc.json"))
	}
}
