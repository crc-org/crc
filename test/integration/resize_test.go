package test_test

import (
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("vary VM parameters: memory cpus, disk", func() {

	Describe("use default values", func() {

		It("setup CRC", func() {
			Expect(RunCRCExpectSuccess("setup")).To(ContainSubstring("Setup is complete"))
		})

		It("start CRC", func() {
			// default values: "--memory", "9216", "--cpus", "4", "disk-size", "31"
			if bundlePath == "embedded" {
				Expect(RunCRCExpectSuccess("start", "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
			} else {
				Expect(RunCRCExpectSuccess("start", "-b", bundlePath, "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
			}
		})

		It("check VM's memory size", func() {
			out, err := SendCommandToVM("cat /proc/meminfo")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`MemTotal:[\s]*9\d{6}`))
		})

		It("check VM's number of cpus", func() {
			out, err := SendCommandToVM("cat /proc/cpuinfo")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`processor[\s]*\:[\s]*3`))
			Expect(out).ShouldNot(MatchRegexp(`processor[\s]*\:[\s]*4`))
		})

		It("check VM's disk size", func() {
			out, err := SendCommandToVM("df -h | grep sysroot")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`.*31G[\s].*[\s]/sysroot`))
		})

		It("stop CRC", func() {
			Expect(RunCRCExpectSuccess("stop", "-f")).To(ContainSubstring("Stopped the OpenShift cluster"))
		})

	})

	Describe("use custom values", func() {

		It("start CRC", func() {
			if runtime.GOOS == "darwin" {
				Expect(RunCRCExpectFail("start", "--memory", "12000", "--cpus", "6", "--disk-size", "40")).To(ContainSubstring("Disk resizing is not supported on macOS"))
				Expect(RunCRCExpectSuccess("start", "--memory", "12000", "--cpus", "6")).To(ContainSubstring("Started the OpenShift cluster"))
			} else {
				Expect(RunCRCExpectSuccess("start", "--memory", "12000", "--cpus", "6", "--disk-size", "40")).To(ContainSubstring("Started the OpenShift cluster"))
			}
		})

		It("check VM's memory size", func() {
			out, err := SendCommandToVM("cat /proc/meminfo")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`MemTotal:[\s]*11\d{6}`))
		})

		It("check VM's number of cpus", func() {
			out, err := SendCommandToVM("cat /proc/cpuinfo")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`processor[\s]*\:[\s]*5`))
			Expect(out).ShouldNot(MatchRegexp(`processor[\s]*\:[\s]*6`))
		})

		It("check VM's disk size", func() {
			out, err := SendCommandToVM("df -h | grep sysroot")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`.*40G[\s].*[\s]/sysroot`))
		})

		It("stop CRC", func() {
			Expect(RunCRCExpectSuccess("stop", "-f")).To(ContainSubstring("Stopped the OpenShift cluster"))
		})
	})

	Describe("use flawed values", func() {

		It("start CRC with sub-minimum memory", func() { // less than min = 9216
			Expect(RunCRCExpectFail("start", "--memory", "9000")).To(ContainSubstring("requires memory in MiB >= 9216"))
		})
		It("start CRC with sub-minimum cpus", func() { // fewer than min
			Expect(RunCRCExpectFail("start", "--cpus", "3")).To(ContainSubstring("requires CPUs >= 4"))
		})
		It("start CRC with smaller disk", func() { // bigger than default && smaller than current
			switch runtime.GOOS {
			case "darwin":
				Expect(RunCRCExpectFail("start", "--disk-size", "35")).To(ContainSubstring("Disk resizing is not supported on macOS"))
			case "linux":
				Expect(RunCRCExpectFail("start", "--disk-size", "35")).To(ContainSubstring("current disk image capacity is bigger than the requested size"))
			case "windows":
				Expect(RunCRCExpectFail("start", "--disk-size", "35")).To(ContainSubstring("Failed to set disk size to"))
			}
		})
		It("start CRC with sub-minimum disk", func() { // smaller than min = default = 31GiB
			Expect(RunCRCExpectFail("start", "--disk-size", "30")).To(ContainSubstring("requires disk size in GiB >= 31")) // TODO: message should be different on macOS!
		})
	})

	Describe("use default values again", func() {

		It("start CRC", func() {
			Expect(RunCRCExpectSuccess("start")).To(ContainSubstring("Started the OpenShift cluster")) // default values: "--memory", "9216", "--cpus", "4", "disk-size", "31"
		})

		It("check VM's memory size", func() {
			out, err := SendCommandToVM("cat /proc/meminfo")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`MemTotal:[\s]*9\d{6}`)) // there should be a check if cluster needs >9216MiB; it isn't there and mem gets scaled down regardless
		})

		It("check VM's number of cpus", func() {
			out, err := SendCommandToVM("cat /proc/cpuinfo")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`processor[\s]*\:[\s]*3`))
			Expect(out).ShouldNot(MatchRegexp(`processor[\s]*\:[\s]*4`))
		})

		if runtime.GOOS != "darwin" {
			It("check VM's disk size", func() {
				out, err := SendCommandToVM("df -h | grep sysroot")
				Expect(err).NotTo(HaveOccurred())
				Expect(out).Should(MatchRegexp(`.*40G[\s].*[\s]/sysroot`))
			})
		}

		It("clean up", func() {
			RunCRCExpectSuccess("stop", "-f")
			RunCRCExpectSuccess("delete", "-f")
			RunCRCExpectSuccess("cleanup")

		})
	})
})
