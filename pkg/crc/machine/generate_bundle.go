package machine

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/oc"
	crcssh "github.com/code-ready/crc/pkg/crc/ssh"
	"github.com/code-ready/machine/libmachine/state"
	"github.com/pkg/errors"
)

func (client *client) GenerateBundle() error {
	bundleMetadata, sshRunner, err := loadVM(client)
	if err != nil {
		return err
	}
	defer sshRunner.Close()

	ocConfig := oc.UseOCWithSSH(sshRunner)
	if err := cluster.RemovePullSecretFromCluster(ocConfig, sshRunner); err != nil {
		return errors.Wrap(err, "Error loading bundle metadata")
	}

	// Stop the cluster
	currentState, err := client.Stop()
	if err != nil {
		return err
	}
	if currentState != state.Stopped {
		return fmt.Errorf("VM is not stopped, current state is %s", currentState.String())
	}

	tmpBaseDir, err := ioutil.TempDir(constants.MachineCacheDir, "crc_custom_bundle")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpBaseDir)

	// Create the custom bundle directory which is used as top level directory for tarball during compression
	customBundleName := bundle.GetCustomBundleName(bundleMetadata.GetBundleName())
	customBundleNameWithoutExtension := bundle.GetBundleNameWithoutExtension(customBundleName)

	copier, err := bundle.NewCopier(bundleMetadata, tmpBaseDir, customBundleNameWithoutExtension)
	if err != nil {
		return err
	}
	defer copier.Cleanup() //nolint
	customBundleDir := copier.CachedPath()

	if err := copier.CopyKubeConfig(); err != nil {
		return err
	}

	if err := copier.CopyPrivateSSHKey(constants.GetPrivateKeyPath()); err != nil {
		return err
	}

	if err := copier.CopyFilesFromFileList(); err != nil {
		return err
	}

	// Copy disk image
	logging.Infof("Copying the disk image to %s", customBundleNameWithoutExtension)
	logging.Debugf("Absolute path of custom bundle directory: %s", customBundleDir)
	diskPath, diskFormat, err := copyDiskImage(customBundleDir)
	if err != nil {
		return err
	}

	if err := copier.SetDiskImage(diskPath, diskFormat); err != nil {
		return err
	}

	if err := copier.GenerateBundle(customBundleNameWithoutExtension); err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	logging.Infof("Bundle is generated in %s", filepath.Join(cwd, customBundleName))
	logging.Infof("You need to perform 'crc delete' and 'crc start -b %s' to use this bundle", filepath.Join(cwd, customBundleName))
	return nil
}

func loadVM(client *client) (*bundle.CrcBundleInfo, *crcssh.Runner, error) {
	libMachineAPIClient, cleanup := createLibMachineClient()
	defer cleanup()

	host, err := libMachineAPIClient.Load(client.name)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Cannot load machine")
	}

	currentState, err := host.Driver.GetState()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Cannot get machine state")
	}
	if currentState != state.Running {
		return nil, nil, errors.New("machine is not running")
	}

	crcBundleMetadata, err := getBundleMetadataFromDriver(host.Driver)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error loading bundle metadata")
	}

	instanceIP, err := getIP(host, client.useVSock())
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error getting the IP")
	}
	sshRunner, err := crcssh.CreateRunner(instanceIP, getSSHPort(client.useVSock()), crcBundleMetadata.GetSSHKeyPath(), constants.GetPrivateKeyPath(), constants.GetRsaPrivateKeyPath())
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error creating the ssh client")
	}

	return crcBundleMetadata, sshRunner, nil
}
