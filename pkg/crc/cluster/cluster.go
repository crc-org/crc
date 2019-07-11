package cluster

import (
	"fmt"
	"github.com/code-ready/machine/libmachine/drivers"
	"strings"
	"time"
)

// CheckCertsValidity checks if the cluster certs have expired
func CheckCertsValidity(driver drivers.Driver) error {
	certExpiryDateCmd := `date --date="$(sudo openssl x509 -in /var/lib/kubelet/pki/kubelet-client-current.pem -noout -enddate | cut -d= -f 2)" --iso-8601=seconds`
	output, err := drivers.RunSSHCommandFromDriver(driver, certExpiryDateCmd)
	if err != nil {
		return err
	}
	certExpiryDate, err := time.Parse(time.RFC3339, strings.TrimSpace(output))
	if err != nil {
		return err
	}
	if time.Now().After(certExpiryDate) {
		return fmt.Errorf("Certs have expired, they were valid till: %s", certExpiryDate.Format(time.RFC822))
	}
	return nil
}
