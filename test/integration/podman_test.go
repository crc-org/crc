package test_test

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/crc-org/crc/v2/test/extended/crc/cmd"
	"github.com/crc-org/crc/v2/test/extended/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("podman preset", Serial, Ordered, Label("podman-preset"), func() {

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
			Expect(RunCRCExpectSuccess("config", "set", "preset", "podman")).To(ContainSubstring("please run 'crc setup' before 'crc start'"))
		})

		It("setup CRC", func() {
			Expect(RunCRCExpectSuccess("setup")).To(ContainSubstring("Your system is correctly setup for using CRC"))
		})

		It("start CRC", func() {
			Expect(RunCRCExpectSuccess("start")).To(ContainSubstring("podman runtime is now running"))
		})

		It("podman-env", func() {
			// Do what 'eval $(crc podman-env) would do
			path := os.ExpandEnv("${HOME}/.crc/bin/podman:$PATH")
			csshk := os.ExpandEnv("${HOME}/.crc/machines/crc/id_ecdsa")
			dh := os.ExpandEnv("unix:///${HOME}/.crc/machines/crc/docker.sock")
			ch := "ssh://core@127.0.0.1:2222/run/user/1000/podman/podman.sock"
			if runtime.GOOS == "windows" {
				userHomeDir, _ := os.UserHomeDir()
				unexpandedPath := filepath.Join(userHomeDir, ".crc/bin/podman;${PATH}")
				path = os.ExpandEnv(unexpandedPath)
				csshk = filepath.Join(userHomeDir, ".crc/machines/crc/id_ecdsa")
				dh = "npipe:////./pipe/crc-podman"
			}
			if runtime.GOOS == "linux" {
				ch = "ssh://core@192.168.130.11:22/run/user/1000/podman/podman.sock"
			}

			os.Setenv("PATH", path)
			os.Setenv("CONTAINER_SSHKEY", csshk)
			os.Setenv("CONTAINER_HOST", ch)
			os.Setenv("DOCKER_HOST", dh)
		})

		It("version", func() {
			out, err := cmd.RunPodmanExpectSuccess("version")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).Should(MatchRegexp(`Version:[\s]*\d+\.\d+\.\d+`))
		})

		It("pull image", func() {
			_, err := cmd.RunPodmanExpectSuccess("pull", "fedora")
			Expect(err).NotTo(HaveOccurred())
		})

		It("run image", func() {
			_, err := cmd.RunPodmanExpectSuccess("run", "fedora")
			Expect(err).NotTo(HaveOccurred())
		})

		It("cleanup CRC", func() {
			Expect(RunCRCExpectSuccess("cleanup")).To(MatchRegexp("Cleanup finished"))
		})

		It("unset preset in config", func() {
			Expect(RunCRCExpectSuccess("config", "unset", "preset")).To(ContainSubstring("Successfully unset configuration property 'preset'"))
		})
	})
})
