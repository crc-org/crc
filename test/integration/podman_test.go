package test_test

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/crc-org/crc/v2/test/extended/crc/cmd"
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

		It("podman-env", func() {
			// Do what 'eval $(crc podman-env) would do
			path := os.ExpandEnv("${HOME}/.crc/bin/podman:$PATH")
			csshk := os.ExpandEnv("${HOME}/.crc/machines/crc/id_ed25519")
			dh := os.ExpandEnv("unix:///${HOME}/.crc/machines/crc/docker.sock")
			ch := "ssh://core@127.0.0.1:2222/run/user/1000/podman/podman.sock"
			if runtime.GOOS == "windows" {
				userHomeDir, _ := os.UserHomeDir()
				unexpandedPath := filepath.Join(userHomeDir, ".crc/bin/podman;${PATH}")
				path = os.ExpandEnv(unexpandedPath)
				csshk = filepath.Join(userHomeDir, ".crc/machines/crc/id_ed25519")
				dh = "npipe:////./pipe/crc-podman"
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
