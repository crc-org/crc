package test_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func startInstance() {
	// cluster is not running
	if bundlePath == "" {
		Expect(RunCRCExpectSuccess("setup")).To(ContainSubstring("Your system is correctly setup for using CRC"))
	} else {
		Expect(RunCRCExpectSuccess("setup", "-b", bundlePath)).To(ContainSubstring("Your system is correctly setup for using CRC"))
	}
	It("start CRC", func() {
		Expect(RunCRCExpectSuccess("start", "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
	})
}
