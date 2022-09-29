package test_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Certificate renewal", Label("openshift-preset", "cert-renewal"), func() {

	Describe("11 months in the future", func() {

		It("setup CRC", func() {
			if bundlePath == "" {
				Expect(RunCRCExpectSuccess("setup")).To(ContainSubstring("Your system is correctly setup for using CRC"))
			} else {
				Expect(RunCRCExpectSuccess("setup", "-b", bundlePath)).To(ContainSubstring("Your system is correctly setup for using CRC"))
			}
		})

		It("Prewind clock by 11 months", func() {
			_, err := exec.Command("bash", "-c", "sudo timedatectl set-ntp off").Output()
			Expect(err).NotTo(HaveOccurred())

			_, err = exec.Command("bash", "-c", "sudo date -s '11 month'").Output()
			Expect(err).NotTo(HaveOccurred())

			contains, err := CheckOutputContainsWithRetry(10, "1s", "virsh --readonly -c qemu:///system capabilities", "<capabilities>")
			Expect(err).NotTo(HaveOccurred())
			Expect(contains).To(BeTrue())
		})

		It("start CRC", func() {
			// default values: "--memory", "9216", "--cpus", "4", "disk-size", "31"
			if bundlePath == "" {
				Expect(RunCRCAddEnvExpectSuccess([]string{"CRC_DEBUG_ENABLE_STOP_NTP=true"}, "start", "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
			} else {
				Expect(RunCRCAddEnvExpectSuccess([]string{"CRC_DEBUG_ENABLE_STOP_NTP=true"}, "start", "-b", bundlePath, "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
			}
		})

		It("check cluster", func() {
			err := exec.Command("bash", "-c", "eval $(crc oc-env)").Run()
			Expect(err).NotTo(HaveOccurred())

		})

		It("delete CRC", func() {
			_ = RunCRCExpectSuccess("delete", "-f")
		})

		It("clean up", func() {
			RunCRCExpectSuccess("stop", "-f")
			RunCRCExpectSuccess("delete", "-f")
			RunCRCExpectSuccess("cleanup")

		})

	})

	Describe("13 months in the future", func() {

		It("Prewind clock by another 2 months", func() {
			_, err := exec.Command("bash", "-c", "sudo timedatectl set-ntp off").Output()
			Expect(err).NotTo(HaveOccurred())

			_, err = exec.Command("bash", "-c", "sudo date -s '2 month'").Output()
			Expect(err).NotTo(HaveOccurred())

			contains, err := CheckOutputContainsWithRetry(10, "1s", "virsh --readonly -c qemu:///system capabilities", "<capabilities>")
			Expect(err).NotTo(HaveOccurred())
			Expect(contains).To(BeTrue())
		})

		It("start CRC", func() {
			// default values: "--memory", "9216", "--cpus", "4", "disk-size", "31"
			if bundlePath == "" {
				Expect(RunCRCAddEnvExpectSuccess([]string{"CRC_DEBUG_ENABLE_STOP_NTP=true"}, "start", "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
			} else {
				Expect(RunCRCAddEnvExpectSuccess([]string{"CRC_DEBUG_ENABLE_STOP_NTP=true"}, "start", "-b", bundlePath, "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
			}
		})

		It("check cluster", func() {
			err := exec.Command("bash", "-c", "eval $(crc oc-env)").Run()
			Expect(err).NotTo(HaveOccurred())

		})

		It("clean up", func() {
			RunCRCExpectSuccess("stop", "-f")
			RunCRCExpectSuccess("delete", "-f")
			RunCRCExpectSuccess("cleanup")

			err := exec.Command("bash", "-c", "sudo date -s '-13 month'").Run()
			Expect(err).NotTo(HaveOccurred())

			err = exec.Command("bash", "-c", "sudo timedatectl set-ntp on").Run()
			Expect(err).NotTo(HaveOccurred())

		})

	})

})
