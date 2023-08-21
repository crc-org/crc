package machine

import (
	"context"
	"fmt"

	"github.com/crc-org/crc/v2/pkg/crc/cluster"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine/state"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	"github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/pkg/errors"
)

func (client *client) Status() (*types.ClusterStatusResult, error) {
	vm, err := loadVirtualMachine(client.name, client.useVSock())
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

	vmStatus, err := vm.State()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get machine state")
	}

	clusterStatusResult := &types.ClusterStatusResult{
		CrcStatus: vmStatus,
	}
	switch {
	case vm.bundle.IsPodman():
		clusterStatusResult.PodmanVersion = vm.bundle.GetPodmanVersion()
		clusterStatusResult.Preset = preset.Podman
	case vm.bundle.IsMicroshift():
		clusterStatusResult.OpenshiftStatus = types.OpenshiftStopped
		clusterStatusResult.OpenshiftVersion = vm.bundle.GetOpenshiftVersion()
		clusterStatusResult.Preset = preset.Microshift
	default:
		clusterStatusResult.OpenshiftStatus = types.OpenshiftStopped
		clusterStatusResult.OpenshiftVersion = vm.bundle.GetOpenshiftVersion()
		clusterStatusResult.Preset = preset.OpenShift
	}

	if vmStatus != state.Running {
		return clusterStatusResult, nil
	}

	ip, err := vm.IP()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting ip")
	}

	diskSize, diskUse := client.getDiskDetails(vm)
	clusterStatusResult.CrcStatus = state.Running
	clusterStatusResult.DiskUse = diskUse
	clusterStatusResult.DiskSize = diskSize

	switch {
	case vm.bundle.IsMicroshift():
		clusterStatusResult.OpenshiftStatus = getMicroShiftStatus(context.Background(), ip)
	case vm.bundle.IsOpenShift():
		clusterStatusResult.OpenshiftStatus = getOpenShiftStatus(context.Background(), ip)
	}

	ramSize, ramUse := client.getRAMStatus(vm)
	clusterStatusResult.RAMSize = ramSize
	clusterStatusResult.RAMUse = ramUse

	return clusterStatusResult, nil
}

func (client *client) GetClusterLoad() (*types.ClusterLoadResult, error) {
	vm, err := loadVirtualMachine(client.name, client.useVSock())
	if err != nil {
		if errors.Is(err, errMissingHost(client.name)) {
			return &types.ClusterLoadResult{
				RAMUse:  -1,
				RAMSize: -1,
				CPUUse:  nil,
			}, nil
		}
		return nil, errors.Wrap(err, fmt.Sprintf("Cannot load '%s' virtual machine", client.name))
	}
	defer vm.Close()

	vmStatus, err := vm.State()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get machine state")
	}
	if vmStatus != state.Running {
		return &types.ClusterLoadResult{
			RAMUse:  -1,
			RAMSize: -1,
			CPUUse:  nil,
		}, nil
	}

	ramSize, ramUse := client.getRAMStatus(vm)
	cpuUsage := client.getCPUStatus(vm)

	return &types.ClusterLoadResult{
		RAMUse:  ramUse,
		RAMSize: ramSize,
		CPUUse:  cpuUsage,
	}, nil
}

func (client *client) getDiskDetails(vm *virtualMachine) (int64, int64) {
	disk, err, _ := client.diskDetails.Memoize("disks", func() (interface{}, error) {
		sshRunner, err := vm.SSHRunner()
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
	return getStatus(status)
}

func getMicroShiftStatus(ctx context.Context, ip string) types.OpenshiftStatus {
	status, err := cluster.GetClusterNodeStatus(ctx, ip, constants.KubeconfigFilePath)
	if err != nil {
		logging.Debugf("failed to get microshift node status: %v", err)
		return types.OpenshiftUnreachable
	}
	return getStatus(status)
}

func getStatus(status *cluster.Status) types.OpenshiftStatus {
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

func (client *client) getRAMStatus(vm *virtualMachine) (int64, int64) {
	ram, err, _ := client.ramDetails.Memoize("ram", func() (interface{}, error) {
		sshRunner, err := vm.SSHRunner()
		if err != nil {
			return nil, errors.Wrap(err, "Error creating the ssh client")
		}
		defer sshRunner.Close()
		ramSize, ramUse, err := cluster.GetRAMUsage(sshRunner)
		if err != nil {
			return nil, err
		}
		return []int64{ramSize, ramUse}, nil
	})

	if err != nil {
		logging.Debugf("Cannot get RAM usage: %v", err)
		return -1, -1
	}

	return ram.([]int64)[0], ram.([]int64)[1]
}

func (client *client) getCPUStatus(vm *virtualMachine) []int64 {
	sshRunner, err := vm.SSHRunner()
	if err != nil {
		logging.Debugf("Cannot get SSH runner: %v", err)
		return []int64{}
	}
	defer sshRunner.Close()

	cpuUsage, err := cluster.GetCPUUsage(sshRunner)
	if err != nil {
		logging.Debugf("Cannot get CPU usage: %v", err)
		return nil
	}

	return cpuUsage

}
