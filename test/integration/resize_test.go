package test_test

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/crc-org/crc/v2/test/extended/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("vary VM parameters: memory cpus, disk", Serial, Ordered, Label("openshift-preset", "vm-resize"), func() {
	filename := "time-consume.txt"
	// runs 1x after all the It blocks (specs) inside this Describe node
	AfterAll(func() {

		// cleanup CRC
		Expect(RunCRCExpectSuccess("cleanup")).To(MatchRegexp("Cleanup finished"))

		// remove config file crc.json
		err := util.RemoveCRCConfig()
		Expect(err).NotTo(HaveOccurred())

		_, err = os.Stat(filename)
		if err != nil {
			fmt.Println("Failed to gathering time-consume data")
		}

	})

	Describe("use default values", Serial, Ordered, func() {

		It("setup CRC", func() {
			Expect(
				crcSuccess("setup")).
				To(ContainSubstring("Your system is correctly setup for using CRC"))
		})

		It("start CRC", func() {
			start := time.Now()
			message := crcSuccess("start", "--memory", "12000", "--cpus", "5", "--disk-size", "40", "-p", pullSecretPath)
			duration := time.Since(start)
			Expect(message).To(ContainSubstring("Started the OpenShift cluster"))

			data := "crc start(default): " + duration.String() + "\n"
			writeDataToFile(filename, data)
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
			start := time.Now()
			message := crcSuccess("stop", "-f")
			duration := time.Since(start)
			Expect(message).To(MatchRegexp("[Ss]topped the instance"))

			data := "crc stop(default):" + duration.String() + "\n"
			writeDataToFile(filename, data)
		})

	})

	Describe("use custom values", Serial, Ordered, func() {

		It("start CRC", func() {
			start := time.Now()
			message := crcSuccess("start", "--memory", "13000", "--cpus", "6", "--disk-size", "50")
			duration := time.Since(start)
			Expect(message).To(ContainSubstring("Started the OpenShift cluster"))

			data := "crc start(custom):" + duration.String() + "\n"
			writeDataToFile(filename, data)
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
			start := time.Now()
			message := crcSuccess("stop", "-f")
			duration := time.Since(start)
			Expect(message).To(MatchRegexp("[Ss]topped the instance"))

			data := "crc stop(custom):" + duration.String() + "\n"
			writeDataToFile(filename, data)
		})
	})

	Describe("use flawed values", Serial, Ordered, func() {

		It("start CRC with sub-minimum memory", func() { // less than min = 10752
			Expect(
				crcFails("start", "--memory", "9000")).
				To(ContainSubstring("requires memory in MiB >= 10752"))
		})
		It("start CRC with sub-minimum cpus", func() { // fewer than min
			Expect(
				crcFails("start", "--cpus", "3")).
				To(ContainSubstring("requires CPUs >= 4"))
		})
		It("start CRC with smaller disk", func() { // bigger than default && smaller than current
			diskSizeOutput := "current disk image capacity is bigger than the requested size"
			if runtime.GOOS == "windows" {
				diskSizeOutput = "Failed to set disk size to"
			}
			Expect(
				crcFails("start", "--disk-size", "35")).
				To(ContainSubstring(diskSizeOutput))
		})
		It("start CRC with sub-minimum disk", func() { // smaller than min = default = 31GiB
			Expect(
				crcFails("start", "--disk-size", "30")).
				To(ContainSubstring("requires disk size in GiB >= 31"))
		})
	})

	Describe("use default values again", Serial, Ordered, func() {

		It("start CRC", func() {
			start := time.Now()
			message := crcSuccess("start")
			duration := time.Since(start)
			Expect(message).To(ContainSubstring("Started the OpenShift cluster"))

			data := "crc start(default2):" + duration.String() + "\n"
			writeDataToFile(filename, data)
		})

		It("check VM's memory size", func() {
			out, err := util.SendCommandToVM("cat /proc/meminfo")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`MemTotal:[\s]*1\d{6}`)) // there should be a check if cluster needs >10752MiB; it isn't there and mem gets scaled down regardless
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
			start := time.Now()
			RunCRCExpectSuccess("stop", "-f")
			duration := time.Since(start)
			RunCRCExpectSuccess("delete", "-f")
			RunCRCExpectSuccess("cleanup")

			data := "crc stop(default2):" + duration.String() + "\n"
			writeDataToFile(filename, data)

		})
	})

})
