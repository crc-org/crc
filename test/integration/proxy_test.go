package test_test

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/crc-org/crc/v2/test/extended/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("", Serial, Ordered, Label("openshift-preset", "goproxy"), func() {

	// runs 1x after all the It blocks (specs) inside this Describe node
	AfterAll(func() {

		// cleanup CRC
		Expect(RunCRCExpectSuccess("cleanup")).To(MatchRegexp("Cleanup finished"))

		// remove config file crc.json
		err := util.RemoveCRCConfig()
		Expect(err).NotTo(HaveOccurred())

		// HTTP_PROXY and HTTPS_PROXY vars were set implicitly
		// unset them at the end
		err = os.Unsetenv("HTTPS_PROXY")
		Expect(err).NotTo(HaveOccurred())

		err = os.Unsetenv("HTTP_PROXY")
		Expect(err).NotTo(HaveOccurred())

	})

	go util.RunProxy()

	Describe("Behind proxy", Serial, Ordered, func() {

		networkMode := "user"
		httpProxy := "http://127.0.0.1:8888"
		httpsProxy := "http://127.0.0.1:8888"
		noProxy := ".testing"

		// Start goproxy

		It("configure CRC", func() {
			Expect(RunCRCExpectSuccess("config", "set", "network-mode", networkMode), ContainSubstring("Network mode"))
			Expect(RunCRCExpectSuccess("config", "set", "http-proxy", httpProxy), ContainSubstring("Successfully configured http-proxy"))
			Expect(RunCRCExpectSuccess("config", "set", "https-proxy", httpsProxy), ContainSubstring("Successfully configured https-proxy"))
			Expect(RunCRCExpectSuccess("config", "set", "no-proxy", noProxy), ContainSubstring("Successfully configured no-proxy"))
			Expect(RunCRCExpectSuccess("config", "set", "proxy-ca-file", util.CACertTempLocation), ContainSubstring("Successfully configured proxy-ca-file"))
		})

		It("setup CRC", func() {
			if bundlePath == "" {
				Expect(RunCRCExpectSuccess("setup")).To(ContainSubstring("Your system is correctly setup for using CRC"))
			} else {
				Expect(RunCRCExpectSuccess("setup", "-b", bundlePath)).To(ContainSubstring("Your system is correctly setup for using CRC"))
			}
		})

		It("start CRC", func() {
			// default values: "--memory", "9216", "--cpus", "4", "disk-size", "31"
			if bundlePath == "" {
				Expect(RunCRCExpectSuccess("start", "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
			} else {
				Expect(RunCRCExpectSuccess("start", "-b", bundlePath, "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
			}
		})

		It("login to cluster using crc-admin context", func() {

			err := AddOCToPath()
			Expect(err).NotTo(HaveOccurred())

			cmd := exec.Command("oc", "config", "use-context", "crc-admin")
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			cmd = exec.Command("oc", "whoami")
			out, err := cmd.Output()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).To(ContainSubstring("kubeadmin"))

			cmd = exec.Command("oc", "get", "co")
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())
		})

		It("stop CRC", func() {
			Expect(RunCRCExpectSuccess("stop", "-f")).To(MatchRegexp("[Ss]topped the instance"))
		})

		It("cleanup CRC", func() {
			Expect(RunCRCExpectSuccess("cleanup")).To(MatchRegexp("Cleanup finished"))
		})

	})
})
