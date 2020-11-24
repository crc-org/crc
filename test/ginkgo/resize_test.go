package test_test

import (
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("start flags", func() {

	Describe("memory, cpus, disk", func() {

		It("setup CRC", func() {
			Expect(RunCRCExpectSuccess("setup")).To(ContainSubstring("Setup is complete"))
		})

		It("start CRC with custom memory, cpu, and disk", func() {
			Expect(RunCRCExpectSuccess("start", "-b", bundleLocation, "-p", pullSecretLocation, "--memory", "12000", "--cpus", "6", "--disk-size", "40")).To(ContainSubstring("Started the OpenShift cluster"))
		})

		It("check memory size", func() {
			out, err := SendCommandToVM("cat /proc/meminfo")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`MemTotal:[\s]*11\d{6}`))
		})

		It("check number of cpus", func() {
			out, err := SendCommandToVM("cat /proc/cpuinfo")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`processor[\s]*\:[\s]*5`))
		})

		// only check disk size on linux and windows
		if os := runtime.GOOS; os == "linux" || os == "windows" {

			It("check size of VM disk", func() {
				out, err := SendCommandToVM("df -h")
				Expect(err).NotTo(HaveOccurred())
				Expect(out).Should(MatchRegexp(`.*coreos-luks-root-nocrypt[\s]*40G`))
			})
		} else { // darwin
			It("check size of VM disk", func() {
				out, err := SendCommandToVM("df -h")
				Expect(err).NotTo(HaveOccurred())
				Expect(out).Should(MatchRegexp(`.*coreos-luks-root-nocrypt[\s]*31G`)) // default
			})
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
			Expect(RunCRCExpectFail("start", "--disk-size", "35")).To(ContainSubstring("")) // TODO: diff between darwin & the rest
		})
		It("start CRC and shrink disk", func() { // smaller than min = default = 31GiB
			Expect(RunCRCExpectFail("start", "--disk-size", "30")).To(ContainSubstring("")) // TODO: diff between darwin & the rest
		})

		// start with default specs
		It("start CRC with memory size and cpu count", func() {
			Expect(RunCRCExpectSuccess("start", "-b", bundleLocation, "--memory", "9216", "--cpus", "4")).To(ContainSubstring("Started the OpenShift cluster"))
		})

		It("check memory size", func() {
			out, err := SendCommandToVM("cat /proc/meminfo")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`MemTotal:[\s]*8\d{6}`))
		})

		It("check number of cpus", func() {
			out, err := SendCommandToVM("cat /proc/cpuinfo")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`processor[\s]*\:[\s]*3`))
			Expect(out).ShouldNot(MatchRegexp(`processor[\s]*\:[\s]*4`))
		})

		// only check disk size on linux and windows
		if os := runtime.GOOS; os == "linux" || os == "windows" {
			It("check size of VM disk", func() {
				out, err := SendCommandToVM("df -h")
				Expect(err).NotTo(HaveOccurred())
				Expect(out).Should(MatchRegexp(`.*coreos-luks-root-nocrypt[\s]*40G`)) // no change
			})
		}

		It("clean up", func() {
			RunCRCExpectSuccess("stop", "-f")
			RunCRCExpectSuccess("delete", "-f")
			RunCRCExpectSuccess("cleanup")

		})
	})
})
