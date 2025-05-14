package machine

import (
	"context"
	"fmt"

	"github.com/spf13/cast"

	"github.com/containers/common/pkg/strongunits"

	"github.com/crc-org/crc/v2/pkg/crc/cluster"
	"github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine/state"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	"github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/pkg/errors"
)

type openShiftStatusSupplierFunc func(context.Context, string) types.OpenshiftStatus

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

	ip := ""
	if vmStatus == state.Running {
		ip, err = vm.IP()
		if err != nil {
			return nil, errors.Wrap(err, "Error getting ip")
		}
	}

	ramSize, ramUse := client.getRAMStatus(vm)
	diskSize, diskUse := client.getDiskDetails(vm)
	pvSize, pvUse := client.getPVCSize(vm)
	var openShiftStatusSupplier = getOpenShiftStatus
	if vm.Bundle().IsMicroshift() {
		openShiftStatusSupplier = getMicroShiftStatus
	}

	return createClusterStatusResult(vmStatus, vm.Bundle().GetBundleType(), vm.Bundle().GetVersion(), ip, diskSize, diskUse, ramSize, ramUse, pvUse, pvSize, openShiftStatusSupplier)
}

func createClusterStatusResult(vmStatus state.State, bundleType preset.Preset, vmBundleVersion, vmIP string, diskSize, diskUse, ramSize, ramUse strongunits.B, pvUse, pvSize strongunits.B, openShiftStatusSupplier openShiftStatusSupplierFunc) (*types.ClusterStatusResult, error) {
	clusterStatusResult := &types.ClusterStatusResult{
		CrcStatus:        vmStatus,
		OpenshiftVersion: vmBundleVersion,
		OpenshiftStatus:  types.OpenshiftStopped,
	}
	switch bundleType {
	case preset.Microshift:
		clusterStatusResult.Preset = preset.Microshift
	case preset.OKD:
		clusterStatusResult.Preset = preset.OKD
	default:
		clusterStatusResult.Preset = preset.OpenShift
	}

	if vmStatus != state.Running {
		return clusterStatusResult, nil
	}

	clusterStatusResult.CrcStatus = state.Running
	clusterStatusResult.DiskUse = diskUse
	clusterStatusResult.DiskSize = diskSize
	clusterStatusResult.RAMSize = ramSize
	clusterStatusResult.RAMUse = ramUse
	clusterStatusResult.OpenshiftStatus = openShiftStatusSupplier(context.Background(), vmIP)

	if bundleType == preset.Microshift {
		clusterStatusResult.PersistentVolumeUse = pvUse
		clusterStatusResult.PersistentVolumeSize = pvSize
	}

	return clusterStatusResult, nil
}

func (client *client) GetClusterLoad() (*types.ClusterLoadResult, error) {
	vm, err := loadVirtualMachine(client.name, client.useVSock())
	if err != nil {
		if errors.Is(err, errMissingHost(client.name)) {
			return &types.ClusterLoadResult{
				RAMUse:  0,
				RAMSize: 0,
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
			RAMUse:  0,
			RAMSize: 0,
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

func (client *client) getDiskDetails(vm VirtualMachine) (strongunits.B, strongunits.B) {
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
		return []strongunits.B{diskSize, diskUse}, nil
	})
	if err != nil {
		logging.Debugf("Cannot get root partition usage: %v", err)
		return 0, 0
	}
	return disk.([]strongunits.B)[0], disk.([]strongunits.B)[1]
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

func (client *client) getRAMStatus(vm VirtualMachine) (strongunits.B, strongunits.B) {
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
		return []uint64{cast.ToUint64(ramSize), cast.ToUint64(ramUse)}, nil
	})

	if err != nil {
		logging.Debugf("Cannot get RAM usage: %v", err)
		return 0, 0
	}

	used, total := ram.([]uint64)[0], ram.([]uint64)[1]
	return strongunits.B(used), strongunits.B(total)
}

func (client *client) getCPUStatus(vm VirtualMachine) []int64 {
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

func (client *client) getPVCSize(vm VirtualMachine) (strongunits.B, strongunits.B) {
	sshRunner, err := vm.SSHRunner()
	if err != nil {
		logging.Debugf("Cannot get SSH runner: %v", err)
		return strongunits.B(0), strongunits.B(0)
	}
	total := client.config.Get(config.PersistentVolumeSize)
	defer sshRunner.Close()
	used, err := cluster.GetPVCUsage(sshRunner)
	if err != nil {
		logging.Debugf("Cannot get PVC usage: %v", err)
		return strongunits.B(0), strongunits.B(0)
	}
	return used, strongunits.GiB(cast.ToUint64(total.AsInt())).ToBytes()
}
