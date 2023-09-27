package test_test

import (
	"os/exec"
	"runtime"

	"github.com/crc-org/crc/v2/test/extended/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("vary VM parameters: memory cpus, disk", Serial, Ordered, Label("openshift-preset", "vm-resize"), func() {

	// runs 1x after all the It blocks (specs) inside this Describe node
	AfterAll(func() {

		// cleanup CRC
		Expect(RunCRCExpectSuccess("cleanup")).To(MatchRegexp("Cleanup finished"))

		// remove config file crc.json
		err := util.RemoveCRCConfig()
		Expect(err).NotTo(HaveOccurred())

	})

	Describe("use default values", Serial, Ordered, func() {

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
				Expect(RunCRCExpectSuccess("start", "--memory", "12000", "--cpus", "5", "--disk-size", "40", "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
			} else {
				Expect(RunCRCExpectSuccess("start", "--memory", "12000", "--cpus", "5", "--disk-size", "40", "-b", bundlePath, "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
			}
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

		It("check VM's memory size", func() {
			out, err := util.SendCommandToVM("cat /proc/meminfo")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`MemTotal:[\s]*11\d{6}`))
		})

		It("check VM's number of cpus", func() {
			out, err := util.SendCommandToVM("cat /proc/cpuinfo")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`processor[\s]*\:[\s]*4`))
			Expect(out).ShouldNot(MatchRegexp(`processor[\s]*\:[\s]*5`))
		})

		It("check VM's disk size", func() {
			out, err := util.SendCommandToVM("df -h | grep sysroot")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`.*40G[\s].*[\s]/sysroot`))
		})

		It("stop CRC", func() {
			Expect(RunCRCExpectSuccess("stop", "-f")).To(MatchRegexp("[Ss]topped the instance"))
		})

	})

	Describe("use custom values", Serial, Ordered, func() {

		It("start CRC", func() {
			Expect(RunCRCExpectSuccess("start", "--memory", "13000", "--cpus", "6", "--disk-size", "50")).To(ContainSubstring("Started the OpenShift cluster"))
		})

		It("check VM's memory size", func() {
			out, err := util.SendCommandToVM("cat /proc/meminfo")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`MemTotal:[\s]*12\d{6}`))
		})

		It("check VM's number of cpus", func() {
			out, err := util.SendCommandToVM("cat /proc/cpuinfo")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`processor[\s]*\:[\s]*5`))
			Expect(out).ShouldNot(MatchRegexp(`processor[\s]*\:[\s]*6`))
		})

		It("check VM's disk size", func() {
			out, err := util.SendCommandToVM("df -h | grep sysroot")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`.*50G[\s].*[\s]/sysroot`))
		})

		It("stop CRC", func() {
			Expect(RunCRCExpectSuccess("stop", "-f")).To(MatchRegexp("[Ss]topped the instance"))
		})
	})

	Describe("use flawed values", Serial, Ordered, func() {

		It("start CRC with sub-minimum memory", func() { // less than min = 9216
			Expect(RunCRCExpectFail("start", "--memory", "9000")).To(ContainSubstring("requires memory in MiB >= 9216"))
		})
		It("start CRC with sub-minimum cpus", func() { // fewer than min
			Expect(RunCRCExpectFail("start", "--cpus", "3")).To(ContainSubstring("requires CPUs >= 4"))
		})
		It("start CRC with smaller disk", func() { // bigger than default && smaller than current
			if runtime.GOOS == "windows" {
				Expect(RunCRCExpectFail("start", "--disk-size", "35")).To(ContainSubstring("Failed to set disk size to"))
			} else {
				Expect(RunCRCExpectFail("start", "--disk-size", "35")).To(ContainSubstring("current disk image capacity is bigger than the requested size"))
			}
		})
		It("start CRC with sub-minimum disk", func() { // smaller than min = default = 31GiB
			Expect(RunCRCExpectFail("start", "--disk-size", "30")).To(ContainSubstring("requires disk size in GiB >= 31")) // TODO: message should be different on macOS!
		})
	})

	Describe("use default values again", Serial, Ordered, func() {

		It("start CRC", func() {
			Expect(RunCRCExpectSuccess("start")).To(ContainSubstring("Started the OpenShift cluster")) // default values: "--memory", "9216", "--cpus", "4", "disk-size", "31"
		})

		It("check VM's memory size", func() {
			out, err := util.SendCommandToVM("cat /proc/meminfo")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`MemTotal:[\s]*9\d{6}`)) // there should be a check if cluster needs >9216MiB; it isn't there and mem gets scaled down regardless
		})

		It("check VM's number of cpus", func() {
			out, err := util.SendCommandToVM("cat /proc/cpuinfo")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`processor[\s]*\:[\s]*3`))
			Expect(out).ShouldNot(MatchRegexp(`processor[\s]*\:[\s]*4`))
		})

		if runtime.GOOS != "darwin" {
			It("check VM's disk size", func() {
				out, err := util.SendCommandToVM("df -h | grep sysroot")
				Expect(err).NotTo(HaveOccurred())
				Expect(out).Should(MatchRegexp(`.*50G[\s].*[\s]/sysroot`))
			})
		}

		It("clean up", func() {
			RunCRCExpectSuccess("stop", "-f")
			RunCRCExpectSuccess("delete", "-f")
			RunCRCExpectSuccess("cleanup")

		})
	})

})
