package test_test

import (
	"os/exec"
	"strconv"

	// "github.com/crc-org/crc/test/extended/crc/cmd".
	"github.com/crc-org/crc/test/extended/crc/cmd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	sleepSeconds         = "30"
	maxDateDiffInSeconds = 10
)

var _ = Describe("Guest machine time should be in sync with host", Label("guest-timesync", "darwin"), func() {

	BeforeSuite(func() {
		It("Ensuring instance is running")
		if err := cmd.CheckCRCStatus("running"); err != nil {
			// cluster is not running
			startInstance()
		}
	})

	It("checks instance date is in sync with host after host suspension", func() {
		By("scheduling a suspensions and resume of the host")
		Expect(exec.Command("pmset", "relative", sleepSeconds).Run()).To(Succeed())
		Expect(exec.Command("pmset", "sleepnow").Run()).To(Succeed())
		By("getting dates from instance and host")
		out, err := exec.Command("date", "-u", "+%s").Output()
		Expect(err).To(BeEmpty())
		hostDate := string(out)
		instanceDate, err := SendCommandToVM("date -u +%s")
		Expect(err).To(BeEmpty())
		By("ensuring date matches on host and instance")
		hostDateAsInt, err := strconv.Atoi(hostDate)
		Expect(err).To(BeEmpty())
		instanceDateAsInt, err := strconv.Atoi(instanceDate)
		Expect(err).To(BeEmpty())
		dateDiff := instanceDateAsInt - hostDateAsInt
		Ω(dateDiff).Should(BeNumerically("<=", maxDateDiffInSeconds))
	})

	AfterSuite(func() {
		// delete instance cleanup...
		It("Cleanup instance")
		Expect(cmd.CRC("cleanup").Execute()).To(Succeed())
	})

})
