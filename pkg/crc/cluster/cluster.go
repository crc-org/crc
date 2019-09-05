package cluster

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/machine/libmachine/drivers"
)

func WaitForSsh(driver drivers.Driver) error {
	checkSshConnectivity := func() error {
		_, err := drivers.RunSSHCommandFromDriver(driver, "exit 0")
		if err != nil {
			return &errors.RetriableError{Err: err}
		}
		return nil
	}

	return errors.RetryAfter(10, checkSshConnectivity, time.Second)
}

// CheckCertsValidityUsingBundleBuildTime check if the cluster certs going to expire soon.
func CheckCertsValidityUsingBundleBuildTime(buildTime time.Time) (bool, int) {
	certExpiryDate := buildTime.AddDate(0, 1, 0)
	// Warn user if the cert expiry going to happen starting of the 7 days
	timeAfter7Days := time.Now().AddDate(0, 0, 7)
	return timeAfter7Days.After(certExpiryDate), int(time.Until(certExpiryDate).Hours()) / 24
}

// CheckCertsValidity checks if the cluster certs have expired or going to expire in next 7 days
func CheckCertsValidity(driver drivers.Driver) (bool, int, error) {
	certExpiryDate, err := getcertExipryDateFromVM(driver)
	if err != nil {
		return false, 0, err
	}
	if time.Now().After(certExpiryDate) {
		return false, 0, fmt.Errorf("Certs have expired, they were valid till: %s", certExpiryDate.Format(time.RFC822))
	}

	// Warn user if the cert expiry going to happen starting of the 7 days
	timeAfter7Days := time.Now().AddDate(0, 0, 7)
	if timeAfter7Days.After(certExpiryDate) {
		return true, int(time.Until(certExpiryDate).Hours()) / 24, nil
	}
	return false, 0, nil
}

func getcertExipryDateFromVM(driver drivers.Driver) (time.Time, error) {
	certExpiryDate := time.Time{}
	certExpiryDateCmd := `date --date="$(sudo openssl x509 -in /var/lib/kubelet/pki/kubelet-client-current.pem -noout -enddate | cut -d= -f 2)" --iso-8601=seconds`
	output, err := drivers.RunSSHCommandFromDriver(driver, certExpiryDateCmd)
	if err != nil {
		return certExpiryDate, err
	}
	certExpiryDate, err = time.Parse(time.RFC3339, strings.TrimSpace(output))
	if err != nil {
		return certExpiryDate, err
	}
	return certExpiryDate, nil
}

// Return size of disk, used space in bytes and the mountpoint
func GetRootPartitionUsage(driver drivers.Driver) (int64, int64, error) {
	cmd := "df -B1 --output=size,used,target /sysroot | tail -1"

	out, err := drivers.RunSSHCommandFromDriver(driver, cmd)

	if err != nil {
		return 0, 0, err
	}
	diskDetails := strings.Split(strings.TrimSpace(out), " ")
	diskSize, err := strconv.ParseInt(diskDetails[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	diskUsage, err := strconv.ParseInt(diskDetails[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return diskSize, diskUsage, nil
}
