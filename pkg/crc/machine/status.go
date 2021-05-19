package machine

import (
	"context"
	"time"

	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/machine/types"
	crcssh "github.com/code-ready/crc/pkg/crc/ssh"
	"github.com/code-ready/machine/libmachine/state"
	"github.com/pkg/errors"
)

func (client *client) Status() (*types.ClusterStatusResult, error) {
	libMachineAPIClient, cleanup := createLibMachineClient()
	defer cleanup()

	_, err := libMachineAPIClient.Exists(client.name)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot check if machine exists")
	}

	host, err := libMachineAPIClient.Load(client.name)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot load machine")
	}
	vmStatus, err := host.Driver.GetState()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get machine state")
	}

	crcBundleMetadata, err := getBundleMetadataFromDriver(host.Driver)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading bundle metadata")
	}

	if vmStatus != state.Running {
		return &types.ClusterStatusResult{
			CrcStatus:        vmStatus,
			OpenshiftStatus:  types.OpenshiftStopped,
			OpenshiftVersion: crcBundleMetadata.GetOpenshiftVersion(),
		}, nil
	}

	ip, err := getIP(host, client.useVSock())
	if err != nil {
		return nil, errors.Wrap(err, "Error getting ip")
	}

	diskSize, diskUse := client.getDiskDetails(ip, crcBundleMetadata)
	return &types.ClusterStatusResult{
		CrcStatus:        state.Running,
		OpenshiftStatus:  getOpenShiftStatus(context.Background(), ip, crcBundleMetadata),
		OpenshiftVersion: crcBundleMetadata.GetOpenshiftVersion(),
		DiskUse:          diskUse,
		DiskSize:         diskSize,
	}, nil
}

func (client *client) getDiskDetails(ip string, bundle *bundle.CrcBundleInfo) (int64, int64) {
	disk, err, _ := client.diskDetails.Memoize("disks", func() (interface{}, error) {
		sshRunner, err := crcssh.CreateRunner(ip, getSSHPort(client.useVSock()), constants.GetPrivateKeyPath(), constants.GetRsaPrivateKeyPath(), bundle.GetSSHKeyPath())
		if err != nil {
			return nil, errors.Wrap(err, "Error creating the ssh client")
		}
		defer sshRunner.Close()
		sshRunnerWithTimeout := sshRunner.WithTimeout(5 * time.Second)
		diskSize, diskUse, err := cluster.GetRootPartitionUsage(sshRunnerWithTimeout)
		if err != nil {
			return nil, err
		}
		return []int64{diskSize, diskUse}, nil
	})
	if err != nil {
		logging.Debugf("Cannot get root partition usage: %v", err)
		return 0, 0
	}
	return disk.([]int64)[0], disk.([]int64)[1]
}

func getOpenShiftStatus(ctx context.Context, ip string, bundle *bundle.CrcBundleInfo) types.OpenshiftStatus {
	status, err := cluster.GetClusterOperatorsStatus(ctx, ip, bundle)
	if err != nil {
		logging.Debugf("cannot get OpenShift status: %v", err)
		return types.OpenshiftUnreachable
	}
	switch {
	case status.Progressing:
		return types.OpenshiftStarting
	case status.Degraded:
		return types.OpenshiftDegraded
	case status.Available:
		return types.OpenshiftRunning
	}
	return types.OpenshiftStopped
}
