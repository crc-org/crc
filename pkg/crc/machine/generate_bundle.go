package machine

import (
	"context"
	"os"
	"path/filepath"

	"github.com/crc-org/crc/v2/pkg/crc/cluster"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine/bundle"
	"github.com/crc-org/crc/v2/pkg/crc/oc"
	crcssh "github.com/crc-org/crc/v2/pkg/crc/ssh"
	"github.com/crc-org/machine/libmachine/state"
	"github.com/pkg/errors"
)

func (client *client) GenerateBundle(forceStop bool) error {
	bundleMetadata, sshRunner, err := loadVM(client)
	if err != nil {
		return err
	}
	defer sshRunner.Close()

	if bundleMetadata.IsOpenShift() {
		ocConfig := oc.UseOCWithSSH(sshRunner)
		if err := cluster.RemovePullSecretFromCluster(context.Background(), ocConfig, sshRunner); err != nil {
			return errors.Wrap(err, "Error removing pull secret from cluster")
		}

		if err := cluster.RemoveOldRenderedMachineConfig(ocConfig); err != nil {
			return errors.Wrap(err, "Error removing old rendered machine configs")
		}
	}

	// Stop the cluster
	if _, err := client.Stop(); err != nil {
		if forceStop {
			if err := client.PowerOff(); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	running, err := client.IsRunning()
	if err != nil {
		return err
	}
	if running {
		return errors.New("VM is still running")
	}

	tmpBaseDir, err := os.MkdirTemp(constants.MachineCacheDir, "crc_custom_bundle")
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
	vm, err := loadVirtualMachine(client.name, client.useVSock())
	if err != nil {
		return nil, nil, errors.Wrap(err, "Cannot load machine")
	}
	defer vm.Close()

	currentState, err := vm.Driver.GetState()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Cannot get machine state")
	}
	if currentState != state.Running {
		return nil, nil, errors.New("machine is not running")
	}

	sshRunner, err := vm.SSHRunner()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error creating the ssh client")
	}

	return vm.bundle, sshRunner, nil
}
