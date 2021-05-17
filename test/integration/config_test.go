package test_test

import (
	"runtime"
	"time"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("config", func() {

	// 0 starts
	Describe("skipping a check during setup", func() {

		if runtime.GOOS == "linux" {

			It("stop libvirt service", func() {
				_, _, err := RunOnHostWithPrivilege("2s", "systemctl", "stop", "libvirtd")
				Expect(err).NotTo(HaveOccurred())
			})

			It("check libvirt service not running", func() {
				stdout, _, err := RunOnHostWithPrivilege("2s", "systemctl", "status", "libvirtd")
				Expect(err).To(HaveOccurred()) // exitcode 3, inactive
				Expect(stdout).To(ContainSubstring("inactive"))
			})

			It("set config to skip check that libvirt service is running", func() {
				Expect(RunCRCExpectSuccess("config", "set", "skip-check-libvirt-running", "true")).To(ContainSubstring("Successfully configured skip-check-libvirt-running to true"))
			})

			It("check if setup needs to be run", func() {
				_, err := RunCRCExpectFail("setup", "--check-only")
				Expect(err).NotTo(HaveOccurred())
			})

			It("run setup", func() {
				stderr, err := RunCRCExpectFail("setup")
				Expect(err).NotTo(HaveOccurred())
				Expect(stderr).To(ContainSubstring("failed to connect to the hypervisor"))
			})

			It("set config to not skip the check that libvirt service is running", func() {
				Expect(RunCRCExpectSuccess("config", "unset", "skip-check-libvirt-running")).To(ContainSubstring("Successfully unset configuration property 'skip-check-libvirt-running'"))
			})

			It("run setup", func() {
				Expect(RunCRCExpectSuccess("setup")).To(ContainSubstring("Your system is correctly setup for using CodeReady Containers"))
			})

			It("check libvirt service running", func() {
				stdout, _, err := RunOnHostWithPrivilege("2s", "systemctl", "status", "libvirtd")
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(ContainSubstring("active"))
			})

		}
	})

	// 3 starts
	Describe("Priority: flags >> config >> default", func() {

		It("setup CRC", func() {
			Expect(RunCRCExpectSuccess("setup")).To(ContainSubstring("Your system is correctly setup for using CodeReady Containers"))
		})

		It("set memory", func() {
			Expect(RunCRCExpectSuccess("config", "set", "memory", "10200")).To(ContainSubstring("Changes to configuration property 'memory' are only applied when the CRC instance is started."))
		})

		It("start CRC with memory flag", func() {
			if bundlePath == "embedded" {
				Expect(RunCRCExpectSuccess("start", "-p", pullSecretPath, "--memory", "11200")).To(ContainSubstring("Started the OpenShift cluster"))
			} else {
				Expect(RunCRCExpectSuccess("start", "-b", bundlePath, "-p", pullSecretPath, "--memory", "11200")).To(ContainSubstring("Started the OpenShift cluster"))
			}
			// memory amount should respect the flag
			ip := strings.TrimSpace(RunCRCExpectSuccess("ip"))
			out, err := SendCommandToVM("cat /proc/meminfo", ip, "22")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`MemTotal:[\s]*11\d{6}`))
		})

		It("stop CRC", func() {
			Expect(RunCRCExpectSuccess("stop", "-f")).To(MatchRegexp(`[Ss]topped the OpenShift cluster`))
		})

		It("start CRC and fall back on config settings", func() {
			if bundlePath == "embedded" {
				Expect(RunCRCExpectSuccess("start", "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
			} else {
				Expect(RunCRCExpectSuccess("start", "-b", bundlePath, "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
			}
			// memory amount should respect the flag
			ip := strings.TrimSpace(RunCRCExpectSuccess("ip"))
			out, err := SendCommandToVM("cat /proc/meminfo", ip, "22")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`MemTotal:[\s]*10\d{6}`))
		})

		It("stop CRC", func() {
			Expect(RunCRCExpectSuccess("stop", "-f")).To(MatchRegexp(`[Ss]topped the OpenShift cluster`))
		})

		It("unset memory", func() {
			Expect(RunCRCExpectSuccess("config", "unset", "memory")).To(ContainSubstring("Successfully unset configuration property 'memory'"))
		})

		It("start CRC and fall back on default settings", func() {
			if bundlePath == "embedded" {
				Expect(RunCRCExpectSuccess("start", "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
			} else {
				Expect(RunCRCExpectSuccess("start", "-b", bundlePath, "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
			}
			// memory amount should respect the flag
			ip := strings.TrimSpace(RunCRCExpectSuccess("ip"))
			out, err := SendCommandToVM("cat /proc/meminfo", ip, "22")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`MemTotal:[\s]*9\d{6}`))
		})

		It("cleanup and setup", func() {
			Expect(RunCRCExpectSuccess("cleanup")).To(ContainSubstring("Cleanup finished"))
			Expect(RunCRCExpectSuccess("setup")).To(ContainSubstring("Your system is correctly setup for using CodeReady Containers"))
		})

	})

	// 2 starts
	Describe("host-network-access", func() {

		if runtime.GOOS == "linux" {

			It("set network-mode to vsock", func() {
				Expect(RunCRCExpectSuccess("config", "set", "network-mode", "vsock")).To(ContainSubstring("Please run `crc cleanup` and `crc setup`"))
			})

			It("cleanup and setup, as instructed", func() {
				Expect(RunCRCExpectSuccess("cleanup")).To(ContainSubstring("Cleanup finished"))
				Expect(RunCRCExpectSuccess("setup")).To(ContainSubstring("Your system is correctly setup for using CodeReady Containers"))
			})

			It("set host-network-access to true", func() {
				Expect(RunCRCExpectSuccess("config", "set", "host-network-access", "true")).To(ContainSubstring("Successfully configured host-network-access to true"))
			})

			It("access host's network", func() {

				dChan := make(chan string, 1)
				go RunCRCDaemon(dChan)
				time.Sleep(10 * time.Second) // wait till daemon is running

				_ = RunPodmanExpectSuccess("build", "-t", "http-server:latest", "../testdata/host-network-access")
				_ = RunPodmanExpectSuccess("run", "--name", "crc-http-server", "-d", "-p", "1234:8080", "http-server:latest") // send to background to get exit code

				Expect(RunCRCExpectSuccess("start", "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))

				ip := strings.TrimSpace(RunCRCExpectSuccess("ip"))
				out, err := SendCommandToVM("curl host.crc.testing:1234", ip, "2222")
				Expect(err).NotTo(HaveOccurred())
				Expect(out).To(ContainSubstring("hello"))

				_ = RunPodmanExpectSuccess("kill", "crc-http-server") // kill the container
				Expect(err).NotTo(HaveOccurred())

				_ = RunPodmanExpectSuccess("image", "rm", "http-server", "-f") // remove image

				dChan <- "done" // stop the daemon
			})

			It("cleanup", func() {
				Expect(RunCRCExpectSuccess("cleanup")).To(ContainSubstring("Cleanup finished"))
			})

			It("unset network config", func() {
				Expect(RunCRCExpectSuccess("config", "unset", "host-network-access")).To(ContainSubstring("Successfully unset configuration property 'host-network-access'"))
				Expect(RunCRCExpectSuccess("config", "unset", "network-mode")).To(ContainSubstring("Successfully unset configuration property 'network-mode'"))
			})

		}
	})
})
