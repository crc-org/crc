package machine

import (
	"context"
	"fmt"

	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/machine/state"
	"github.com/code-ready/crc/pkg/crc/machine/types"
	crcssh "github.com/code-ready/crc/pkg/crc/ssh"
	libmachinestate "github.com/code-ready/machine/libmachine/state"
	"github.com/pkg/errors"
)

func (client *client) Status() (*types.ClusterStatusResult, error) {
	vm, err := loadVirtualMachine(client.name)
	if err != nil {
		if errors.Is(err, errMissingHost(client.name)) {
			return &types.ClusterStatusResult{
				CrcStatus:       state.Stopped,
				OpenshiftStatus: types.OpenshiftStopped,
			}, nil
		}
		return nil, errors.Wrap(err, fmt.Sprintf("Cannot load '%s' virtual machine", client.name))
	}
	defer vm.Close()

	vmStatus, err := vm.Driver.GetState()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get machine state")
	}

	if vmStatus != libmachinestate.Running {
		clusterStatusResult := &types.ClusterStatusResult{
			CrcStatus: state.FromMachine(vmStatus),
		}
		if vm.bundle.IsOpenShift() {
			clusterStatusResult.OpenshiftStatus = types.OpenshiftStopped
			clusterStatusResult.OpenshiftVersion = vm.bundle.GetOpenshiftVersion()
		} else {
			clusterStatusResult.PodmanVersion = vm.bundle.GetPodmanVersion()
		}
		return clusterStatusResult, nil
	}

	ip, err := getIP(vm.Host, client.useVSock())
	if err != nil {
		return nil, errors.Wrap(err, "Error getting ip")
	}

	diskSize, diskUse := client.getDiskDetails(ip, vm.bundle)
	clusterStatusResult := &types.ClusterStatusResult{
		CrcStatus: state.Running,
		DiskUse:   diskUse,
		DiskSize:  diskSize,
	}
	if vm.bundle.IsOpenShift() {
		clusterStatusResult.OpenshiftStatus = getOpenShiftStatus(context.Background(), ip)
		clusterStatusResult.OpenshiftVersion = vm.bundle.GetOpenshiftVersion()
	} else {
		clusterStatusResult.PodmanVersion = vm.bundle.GetPodmanVersion()
	}
	return clusterStatusResult, nil
}

func (client *client) getDiskDetails(ip string, bundle *bundle.CrcBundleInfo) (int64, int64) {
	disk, err, _ := client.diskDetails.Memoize("disks", func() (interface{}, error) {
		sshRunner, err := crcssh.CreateRunner(ip, getSSHPort(client.useVSock()), constants.GetPrivateKeyPath(), constants.GetRsaPrivateKeyPath(), bundle.GetSSHKeyPath())
		if err != nil {
			return nil, errors.Wrap(err, "Error creating the ssh client")
		}
		defer sshRunner.Close()
		diskSize, diskUse, err := cluster.GetRootPartitionUsage(sshRunner)
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

func getOpenShiftStatus(ctx context.Context, ip string) types.OpenshiftStatus {
	status, err := cluster.GetClusterOperatorsStatus(ctx, ip, constants.KubeconfigFilePath)
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
