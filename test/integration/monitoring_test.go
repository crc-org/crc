package integration_test

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("run openshift with monitoring stack", func() {
	BeforeAll(cleanUp())

	It("setup CRC", func() {
		RunCRCExpectSuccess("config", "set", "memory", "16000")
		RunCRCExpectSuccess("config", "set", "enable-cluster-monitoring", "true")
		Expect(RunCRCExpectSuccess("setup")).To(ContainSubstring("Your system is correctly setup for using CodeReady Containers"))
	})

	It("start CRC", func() {
		if bundlePath == "embedded" {
			Expect(RunCRCExpectSuccess("start", "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
		} else {
			Expect(RunCRCExpectSuccess("start", "-b", bundlePath, "-p", pullSecretPath)).To(ContainSubstring("Started the OpenShift cluster"))
		}
	})

	It("collect pods usage", func() {
		stdout, _, err := Exec(exec.Command("oc", "adm", "top", "pods", "-A"), nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(ioutil.WriteFile(filepath.Join(outputDirectory, "oc-adm-top-pods.txt"), []byte(stdout), 0600)).To(Succeed())
	})
})
