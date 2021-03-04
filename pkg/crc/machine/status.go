package machine

import (
	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
	crcssh "github.com/code-ready/crc/pkg/crc/ssh"
	"github.com/code-ready/machine/libmachine/state"
	"github.com/pkg/errors"
)

func (client *client) Status() (*ClusterStatusResult, error) {
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

	_, crcBundleMetadata, err := getBundleMetadataFromDriver(host.Driver)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading bundle metadata")
	}

	if vmStatus != state.Running {
		return &ClusterStatusResult{
			CrcStatus:        vmStatus,
			OpenshiftStatus:  "Stopped",
			OpenshiftVersion: crcBundleMetadata.GetOpenshiftVersion(),
		}, nil
	}

	ip, err := getIP(host, client.useVSock())
	if err != nil {
		return nil, errors.Wrap(err, "Error getting ip")
	}
	sshRunner, err := crcssh.CreateRunner(ip, getSSHPort(client.useVSock()), constants.GetPrivateKeyPath(), constants.GetRsaPrivateKeyPath(), crcBundleMetadata.GetSSHKeyPath())
	if err != nil {
		return nil, errors.Wrap(err, "Error creating the ssh client")
	}
	defer sshRunner.Close()
	// check if all the clusteroperators are running
	diskSize, diskUse, err := cluster.GetRootPartitionUsage(sshRunner)
	if err != nil {
		logging.Debugf("Cannot get root partition usage: %v", err)
	}
	return &ClusterStatusResult{
		CrcStatus:        state.Running,
		OpenshiftStatus:  getOpenShiftStatus(sshRunner, client.monitoringEnabled()),
		OpenshiftVersion: crcBundleMetadata.GetOpenshiftVersion(),
		DiskUse:          diskUse,
		DiskSize:         diskSize,
	}, nil
}

func getOpenShiftStatus(sshRunner *crcssh.Runner, monitoringEnabled bool) string {
	status, err := cluster.GetClusterOperatorsStatus(oc.UseOCWithSSH(sshRunner), monitoringEnabled)
	if err != nil {
		logging.Debugf("cannot get OpenShift status: %v", err)
		return "Unreachable"
	}
	switch {
	case status.Progressing:
		return "Starting"
	case status.Degraded:
		return "Degraded"
	case status.Available:
		return "Running"
	}
	return "Stopped"
}
