package test_test

import (
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test config command", func() {
	/*
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
					Expect(stdout).To(ContainSubstring("Stopped Virtualization daemon"))
				})

				It("set config to skip check that libvirt service is running", func() {
					Expect(RunCRCExpectSuccess("config", "set", "skip-check-libvirt-running", "true")).To(ContainSubstring("Successfully configured skip-check-libvirt-running to true"))
				})

				It("check if setup needs to be run", func() {
					out, err := RunCRCExpectFail("setup", "--check-only")
					Expect(err).NotTo(HaveOccurred())
					Expect(out).To(ContainSubstring("Failed to run 'virsh capabilities'"))
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
				out, err := SendCommandToVM("cat /proc/meminfo")
				Expect(err).NotTo(HaveOccurred())
				Expect(out).Should(MatchRegexp(`MemTotal:[\s]*11\d{6}`))
			})

			It("stop CRC", func() {
				Expect(RunCRCExpectSuccess("stop", "-f")).To(ContainSubstring("Stopped the OpenShift cluster"))
			})

			It("start CRC and fall back on config settings", func() {
				if bundlePath == "embedded" {
					Expect(RunCRCExpectSuccess("start", "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
				} else {
					Expect(RunCRCExpectSuccess("start", "-b", bundlePath, "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
				}
				// memory amount should respect the flag
				out, err := SendCommandToVM("cat /proc/meminfo")
				Expect(err).NotTo(HaveOccurred())
				Expect(out).Should(MatchRegexp(`MemTotal:[\s]*10\d{6}`))
			})

			It("stop CRC", func() {
				Expect(RunCRCExpectSuccess("stop", "-f")).To(ContainSubstring("Stopped the OpenShift cluster"))
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
				out, err := SendCommandToVM("cat /proc/meminfo")
				Expect(err).NotTo(HaveOccurred())
				Expect(out).Should(MatchRegexp(`MemTotal:[\s]*9\d{6}`))
			})

			It("cleanup and setup", func() {
				Expect(RunCRCExpectSuccess("cleanup")).To(ContainSubstring("finished successfully"))
				Expect(RunCRCExpectSuccess("setup")).To(ContainSubstring("Your system is correctly setup for using CodeReady Containers"))
			})

		})
	*/
	// 2 starts
	Describe("Check host-network-access works", func() {

		if runtime.GOOS == "linux" {

			dChan := make(chan string, 1) // keeps track of when deamon is needed

			It("set network-mode to vsock", func() {
				Expect(RunCRCExpectSuccess("config", "set", "network-mode", "vsock")).To(ContainSubstring("Please run `crc cleanup` and `crc setup`"))
			})

			It("cleanup and setup, as instructed", func() {
				Expect(RunCRCExpectSuccess("cleanup")).To(ContainSubstring("Cleanup finished"))
				Expect(RunCRCExpectSuccess("setup")).To(ContainSubstring("Your system is correctly setup for using CodeReady Containers"))
			})

			/*
				It("start CRC", func() {
					Expect(RunCRCExpectSuccess("start", "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
				})

				It("fail to reach the server from VM", func() {
					_, err := SendCommandToVM("curl host.crc.testing:8000")
					Expect(err).To(HaveOccurred())
				})

				It("stop CRC", func() {
					Expect(RunCRCExpectSuccess("stop", "-f")).To(ContainSubstring("Stopped the OpenShift cluster"))
				})
			*/

			It("set host-network-access to true", func() {
				Expect(RunCRCExpectSuccess("config", "set", "host-network-access", "true")).To(ContainSubstring("Successfully configured host-network-access to true"))
			})

			It("access host's network", func() {

				go RunCRCDaemon(dChan)
				time.Sleep(10 * time.Second) // wait till daemon is running

				// ideally, we could do this with podman-remote, but not yet

				_, _, err := RunOnHost("120s", "podman", "build", "-t", "http-server:latest", "/home/jsliacan/github/code-ready/crc/test/testdata/host-network-access")
				Expect(err).NotTo(HaveOccurred())

				_, _, err = RunOnHost("10s", "podman", "run", "--name", "crc-http-server", "-d", "-p", "1234:8080", "http-server:latest") // send to background
				Expect(err).NotTo(HaveOccurred())

				Expect(RunCRCExpectSuccess("start", "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
				out, err := SendCommandToVM("curl host.crc.testing:1234", "127.0.0.1", "2222")
				Expect(err).NotTo(HaveOccurred())
				Expect(out).To(ContainSubstring("hello"))

				_, _, err = RunOnHost("10s", "podman", "kill", "crc-http-server") // kill the container
				Expect(err).NotTo(HaveOccurred())

				_, _, err = RunOnHost("10s", "podman", "image", "rm", "http-server", "-f") // remove image
				Expect(err).NotTo(HaveOccurred())

				dChan <- "done"
			})

			It("cleanup", func() {
				Expect(RunCRCExpectSuccess("cleanup")).To(ContainSubstring("Cleanup finished"))
			})

		}
	})
})
