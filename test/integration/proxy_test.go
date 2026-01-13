package test_test

import (
	"os"
	"os/exec"

	"github.com/crc-org/crc/v2/pkg/crc/adminhelper"
	crc "github.com/crc-org/crc/v2/test/extended/crc/cmd"
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

		// Two-phase proxy configuration:
		// Phase 1 (setup): Use 127.0.0.1 - directly accessible on host
		// Phase 2 (start): Use host.crc.testing - resolves to 127.0.0.1 on host, 192.168.127.1 in VM
		localhostProxy := "http://127.0.0.1:8888"
		hostProxy := "http://host.crc.testing:8888"
		noProxy := ".testing"

		// Start goproxy

		It("configure CRC for setup with localhost proxy", func() {
			// Use localhost proxy for setup phase - host can reach 127.0.0.1 directly
			Expect(RunCRCExpectSuccess("config", "set", "http-proxy", localhostProxy)).To(ContainSubstring("Successfully configured http-proxy"))
			Expect(RunCRCExpectSuccess("config", "set", "https-proxy", localhostProxy)).To(ContainSubstring("Successfully configured https-proxy"))
			Expect(RunCRCExpectSuccess("config", "set", "no-proxy", noProxy)).To(ContainSubstring("Successfully configured no-proxy"))
			Expect(RunCRCExpectSuccess("config", "set", "proxy-ca-file", util.CACertTempLocation)).To(ContainSubstring("Successfully configured proxy-ca-file"))
			Expect(RunCRCExpectSuccess("config", "set", "host-network-access", "true")).To(ContainSubstring("Changes to configuration property 'host-network-access' are only applied during 'crc setup'"))
		})

		It("setup CRC", func() {
			// Setup uses 127.0.0.1 proxy - works on host
			Expect(
				crcSuccess("setup")).
				To(ContainSubstring("Your system is correctly setup for using CRC"))
		})

		It("add host.crc.testing to host's /etc/hosts", func() {
			// After setup, adminhelper is available
			// Add host.crc.testing -> 127.0.0.1 so host can resolve it to localhost proxy
			err := adminhelper.AddToHostsFile("127.0.0.1", "host.crc.testing")
			Expect(err).NotTo(HaveOccurred())
		})

		It("reconfigure CRC proxy for start with host.crc.testing", func() {
			// Switch to host.crc.testing for start phase
			// Host resolves to 127.0.0.1 (from /etc/hosts we just added)
			// VM will resolve to 192.168.127.1 (we'll add to VM's /etc/hosts after start)
			Expect(RunCRCExpectSuccess("config", "set", "http-proxy", hostProxy)).To(ContainSubstring("Successfully configured http-proxy"))
			Expect(RunCRCExpectSuccess("config", "set", "https-proxy", hostProxy)).To(ContainSubstring("Successfully configured https-proxy"))
		})

		It("start CRC", func() {
			// default values: "--memory", "10752", "--cpus", "4", "disk-size", "31"
			Expect(
				crcSuccess("start", "-p", pullSecretPath)).
				To(ContainSubstring("Started the OpenShift cluster"))
		})

		It("wait for cluster in Running state", func() {
			err := crc.WaitForClusterInState("running")
			Expect(err).NotTo(HaveOccurred())
		})

		It("login to cluster using crc-admin context", func() {

			err := util.AddOCToPath()
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
			Expect(
				crcSuccess("stop", "-f")).
				To(MatchRegexp("[Ss]topped the instance"))
		})

		It("cleanup CRC", func() {
			Expect(
				crcSuccess("cleanup")).
				To(MatchRegexp("Cleanup finished"))
		})

	})
})
