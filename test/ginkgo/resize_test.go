package test_test

import (
	//"bytes"
	//"encoding/json"
	//"fmt"
	//"io"
	//"os"
	//"os/exec"
	//"os/user"
	//"path/filepath"
	"runtime"
	//"strings"
	//"syscall"
	//"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	//"github.com/sirupsen/logrus"
)

var _ = Describe("start flags", func() {

	Describe("memory, cpus, disk", func() {

		It("setup CRC", func() {
			Expect(RunCRCExpectSuccess("setup")).To(ContainSubstring("Setup is complete"))
		})

		It("start CRC with custom memory, cpu, and disk", func() {
			Expect(RunCRCExpectSuccess("start", "-b", bundleLocation, "-p", pullSecretLocation, "--memory", "12000", "--cpus", "6", "--disk-size", "40")).To(ContainSubstring("Started the OpenShift cluster"))
		})

		switch os := runtime.GOOS; os {
		case "linux":

			It("check memory size", func() {
				out, err := RunOnHost("ssh", "-i", credPath, "core@192.168.130.11", "cat", "/proc/meminfo")
				Expect(err).ToNot(HaveOccurred())
				Expect(out).Should(MatchRegexp(`MemTotal:[\s]*11\d{6}`))
			})

			It("check number of cpus", func() {
				out, err := RunOnHost("ssh", "-i", credPath, "core@192.168.130.11", "cat", "/proc/cpuinfo")
				Expect(err).ToNot(HaveOccurred())
				Expect(out).Should(MatchRegexp(`processor[\s]*\:[\s]*5`))
			})
			It("check size of VM disk", func() {
				out, err := RunOnHostWithPrivilege("virsh", "vol-info", "crc.qcow2", "crc")
				Expect(err).ToNot(HaveOccurred())
				Expect(out).Should(MatchRegexp(`Capacity:[\s]*40\.00`))
			})
		case "windows":
			// testcase not designed
		case "darwin":
			// feature not implemented
		}
		It("stop CRC", func() {
			Expect(RunCRCExpectSuccess("stop", "-f")).To(ContainSubstring("Stopped the OpenShift cluster"))
		})

		// try bad things
		It("start CRC with too little memory", func() { // less than min = 9216
			Expect(RunCRCExpectFail("start", "--memory", "9000")).To(ContainSubstring("requires memory in MiB >= 9216"))
		})
		It("start CRC with too few cpus", func() { // fewer than min
			Expect(RunCRCExpectFail("start", "--cpus", "3")).To(ContainSubstring("")) // TODO
		})
		It("start CRC and shrink disk", func() { // bigger than default && smaller than current
			Expect(RunCRCExpectFail("start", "--disk-size", "35")).To(ContainSubstring("")) // TODO
		})
		It("start CRC and shrink disk", func() { // smaller than min = default = 31GiB
			Expect(RunCRCExpectFail("start", "--disk-size", "30")).To(ContainSubstring("")) // TODO
		})

		// start with default specs
		It("start CRC with memory size and cpu count", func() {
			Expect(RunCRCExpectSuccess("start", "-b", bundleLocation, "--memory", "9216", "--cpus", "4")).To(ContainSubstring("Started the OpenShift cluster"))
		})

		switch os := runtime.GOOS; os {
		case "linux":
			It("check memory size", func() {
				Expect(RunOnHost("ssh", "-i", credPath, "core@192.168.130.11", "cat", "/proc/meminfo")).Should(MatchRegexp(`MemTotal\:[\s]*8\d{5}`))
			})
			It("check number of cpus", func() {
				stdout, err := RunOnHost("ssh", "-i", credPath, "core@192.168.130.11", "cat", "/proc/cpuinfo")
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).Should(MatchRegexp(`processor[\s]\:[\s]3`))
				Expect(stdout).ShouldNot(MatchRegexp(`processor[\s]\:[\s]4`))
			})
		case "darwin":
			It("check memory size", func() {
				Expect(RunOnHost("ssh", "-i", credPath, "core@192.168.130.11", "cat", "/proc/meminfo")).Should(MatchRegexp(`MemTotal\:[\s]*8\d{5}`))
			})
			It("check number of cpus", func() {
				stdout, err := RunOnHost("-i", credPath, "core@192.168.130.11", "cat", "/proc/cpuinfo")
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).Should(MatchRegexp(`processor[\s]*\:[\s]*3`))
				Expect(stdout).ShouldNot(MatchRegexp(`processor[\s]*\:[\s]*4`))
			})
		case "windows":
			// case not designed
		}

		It("clean up", func() {
			RunCRCExpectSuccess("stop", "-f")
			RunCRCExpectSuccess("delete", "-f")
			RunCRCExpectSuccess("cleanup")

		})
	})
})
