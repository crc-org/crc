package test_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/crc-org/crc/v2/pkg/crc/adminhelper"
	crc "github.com/crc-org/crc/v2/test/extended/crc/cmd"
	"github.com/crc-org/crc/v2/test/extended/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// runCRCWithLiveOutput runs a CRC command and streams output to os.Stdout/Stderr in real-time.
// This helps diagnose hangs by showing progress as it happens.
// Note: We use os.Stdout/Stderr directly instead of GinkgoWriter because GinkgoWriter
// buffers output until the test completes, which defeats the purpose of real-time output.
func runCRCWithLiveOutput(args ...string) error {
	cmd := exec.Command("crc", args...)
	cmd.Env = append(os.Environ(),
		"CRC_DISABLE_UPDATE_CHECK=true",
		"CRC_LOG_LEVEL=debug",
	)

	// Direct stdout and stderr to terminal for real-time output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	_, _ = fmt.Fprintf(os.Stdout, "[%s] Starting: crc %v\n", time.Now().Format(time.RFC3339), args)

	err := cmd.Run()

	_, _ = fmt.Fprintf(os.Stdout, "[%s] Finished: crc %v (err=%v)\n", time.Now().Format(time.RFC3339), args, err)
	return err
}

var _ = Describe("", Serial, Ordered, Label("openshift-preset", "goproxy"), func() {

	var proxyServer *http.Server
	var proxyCleanup func()

	// Start proxy server before any tests run
	BeforeAll(func() {
		// Create proxy server - errors are caught here in the main goroutine
		var err error
		proxyServer, proxyCleanup, err = util.NewProxy()
		Expect(err).NotTo(HaveOccurred())

		ln, err := net.Listen("tcp", "127.0.0.1:8888")
		Expect(err).NotTo(HaveOccurred())

		errChan := make(chan error, 1)

		go func() {
			if err := proxyServer.Serve(ln); err != nil && err != http.ErrServerClosed {
				errChan <- err
			}
		}()

		// Wait for proxy to be ready
		Eventually(func() error {
			select {
			case err := <-errChan:
				return err
			default:
			}
			conn, err := net.Dial("tcp", "127.0.0.1:8888")
			if err == nil {
				conn.Close()
			}
			return err
		}).Should(Succeed())
	})

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

		// Stop the proxy server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if proxyServer != nil {
			err = proxyServer.Shutdown(ctx)
			Expect(err).NotTo(HaveOccurred())
		}

		// Close the log file
		if proxyCleanup != nil {
			proxyCleanup()
		}
	})

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
			// Use live output so we can see progress and diagnose hangs
			args := []string{"setup"}
			if len(bundlePath) > 0 {
				args = append(args, "-b", bundlePath)
			}
			err := runCRCWithLiveOutput(args...)
			Expect(err).NotTo(HaveOccurred())
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
			// Use live output so we can see progress and diagnose hangs
			args := []string{"start", "-p", pullSecretPath}
			if len(bundlePath) > 0 {
				args = append(args, "-b", bundlePath)
			}
			err := runCRCWithLiveOutput(args...)
			Expect(err).NotTo(HaveOccurred())
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
