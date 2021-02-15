package cluster

import (
	"fmt"
	"time"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
)

// WaitForClusterStable checks that the cluster is running a number of consecutive times
func WaitForClusterStable(ocConfig oc.Config, monitoringEnabled bool) error {
	startTime := time.Now()

	retryDuration := 30 * time.Second
	retryCount := 20 // 10 minutes

	numConsecutive := 3
	var count int // holds num of consecutive matches

	for i := 0; i < retryCount; i++ {
		status, err := GetClusterOperatorsStatus(ocConfig, monitoringEnabled)
		if err == nil {
			// update counter for consecutive matches
			if status.IsReady() {
				count++
			} else {
				count = 0
			}
			// break if done
			if count == numConsecutive {
				logging.Debugf("Cluster took %s to stabilize", time.Since(startTime))
				return nil
			}
		} else {
			count = 0
		}
		time.Sleep(retryDuration)
	}

	return fmt.Errorf("cluster operators are still not stable after %s", time.Since(startTime))
}
