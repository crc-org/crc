package test_test

import (
	"time"

	"github.com/crc-org/crc/v2/test/extended/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("podman-remote", Serial, Ordered, Label("microshift-preset"), func() {
	filename := "time-consume.txt"
	// runs 1x after all the It blocks (specs) inside this Describe node
	AfterAll(func() {

		// cleanup CRC
		Expect(RunCRCExpectSuccess("cleanup")).To(MatchRegexp("Cleanup finished"))

		// remove config file crc.json
		err := util.RemoveCRCConfig()
		Expect(err).NotTo(HaveOccurred())

	})

	Describe("basic use", Serial, Ordered, func() {

		It("write to config", func() {
			Expect(RunCRCExpectSuccess("config", "set", "preset", "microshift")).To(ContainSubstring("please run 'crc setup' before 'crc start'"))
		})

		It("setup CRC", func() {
			Expect(
				crcSuccess("setup")).
				To(ContainSubstring("Your system is correctly setup for using CRC"))
		})

		It("start CRC", func() {
			// default values: "--memory", "10752", "--cpus", "4", "disk-size", "31"
			start := time.Now()
			message := crcSuccess("start", "-p", pullSecretPath)
			duration := time.Since(start)
			Expect(message).To(ContainSubstring("Started the MicroShift cluster"))
			data := "crc start: " + duration.String() + "\n"
			writeDataToFile(filename, data)
		})

		It("cleanup CRC", func() {
			Expect(
				crcSuccess("cleanup")).
				To(MatchRegexp("Cleanup finished"))
		})

		It("unset preset in config", func() {
			Expect(
				crcSuccess("config", "unset", "preset")).
				To(ContainSubstring("Successfully unset configuration property 'preset'"))
		})
	})
})
