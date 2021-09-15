package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"
)

// WaitForClusterStable checks that the cluster is running a number of consecutive times
func WaitForClusterStable(ctx context.Context, ip string, kubeconfigFilePath string, proxy *network.ProxyConfig) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	startTime := time.Now()

	retryDuration := 30 * time.Second
	retryCount := 20 // 10 minutes

	if proxy.IsEnabled() {
		// In case proxy is enabled increase the retry count
		// to 10 and this will add addition 5 mins.
		retryCount += 10
	}

	numConsecutive := 3
	var count int // holds num of consecutive matches

	for i := 0; i < retryCount; i++ {
		status, err := GetClusterOperatorsStatus(ctx, ip, kubeconfigFilePath)
		if err == nil {
			// update counter for consecutive matches
			if status.IsReady() {
				count++
				if count == 1 {
					logging.Info("All operators are available. Ensuring stability...")
				} else {
					logging.Infof("Operators are stable (%d/%d)...", count, numConsecutive)
				}
			} else {
				logging.Info(status.String())
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
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryDuration):
		}
	}

	return fmt.Errorf("cluster operators are still not stable after %s", time.Since(startTime))
}
