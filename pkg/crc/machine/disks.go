package machine

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	logging "github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	crcPreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	crcssh "github.com/crc-org/crc/v2/pkg/crc/ssh"
	"golang.org/x/crypto/ssh"
)

func moveTopolvmPartition(ctx context.Context, shiftSize int, vm *virtualMachine, sshRunner *crcssh.Runner) error {
	pvPartition, err := getTopolvmPartition(sshRunner)
	if err != nil {
		return err
	}
	_, _, err = sshRunner.RunPrivileged("move topolvm partition to end of disk",
		fmt.Sprintf("echo '+%dG,' | sudo sfdisk --move-data %s -N %s --force", shiftSize, pvPartition[:len("/dev/.da")], pvPartition[len("/dev/.da"):]))
	var exitErr *ssh.ExitError
	if err != nil {
		if !errors.As(err, &exitErr) {
			return err
		}
		if exitErr.ExitStatus() != 1 {
			return err
		}
	}
	if err == nil {
		logging.Info("Restart VM after moving topolvm partition to end")
		if err := restartHost(ctx, vm, sshRunner); err != nil {
			return fmt.Errorf("Failed to move topolvm partition to increase root partition size: %w", err)
		}
	}
	return nil
}

func growPartition(sshRunner *crcssh.Runner, partition string) error {
	if _, _, err := sshRunner.RunPrivileged(fmt.Sprintf("Growing %s partition", partition), "/usr/bin/growpart", partition[:len("/dev/.da")], partition[len("/dev/.da"):]); err != nil {
		var exitErr *ssh.ExitError
		if !errors.As(err, &exitErr) {
			return err
		}
		if exitErr.ExitStatus() != 1 {
			return err
		}
		logging.Debugf("No free space after %s, nothing to do", partition)
		return nil
	}
	return nil
}

func ocpGetPVShiftSizeGiB(diskSize int, pvSize int) int {
	defaultPvSize := constants.GetDefaultPersistentVolumeSize(crcPreset.OpenShift)
	if pvSize > defaultPvSize {
		return (diskSize - constants.DefaultDiskSize) - (pvSize - defaultPvSize)
	}
	return diskSize - constants.DefaultDiskSize
}

func growRootFileSystem(ctx context.Context, startConfig types.StartConfig, vm *virtualMachine, sshRunner *crcssh.Runner) error {
	if startConfig.Preset == crcPreset.OpenShift {
		sizeToMove := ocpGetPVShiftSizeGiB(startConfig.DiskSize, startConfig.PersistentVolumeSize)
		if err := moveTopolvmPartition(ctx, sizeToMove, vm, sshRunner); err != nil {
			return err
		}
	}
	rootPart, err := getrootPartition(sshRunner, startConfig.Preset)
	if err != nil {
		return err
	}

	if err := growPersistentVolume(sshRunner, startConfig.Preset, startConfig.PersistentVolumeSize); err != nil {
		return fmt.Errorf("Unable to grow persistent volume partition: %w", err)
	}

	// with '/dev/[sv]da4' as input, run 'growpart /dev/[sv]da 4'
	if err := growPartition(sshRunner, rootPart); err != nil {
		return nil
	}
	return nil
}

func growPersistentVolume(sshRunner *crcssh.Runner, preset crcPreset.Preset, persistentVolumeSize int) error {
	if preset == crcPreset.Microshift {
		rootPart, err := getrootPartition(sshRunner, preset)
		if err != nil {
			return err
		}
		lvFullName := "rhel/root"
		if err := growLVForMicroshift(sshRunner, lvFullName, rootPart, persistentVolumeSize); err != nil {
			return err
		}
	}

	if preset == crcPreset.OpenShift {
		pvPartition, err := getTopolvmPartition(sshRunner)
		if err != nil {
			return err
		}
		if err := growPartition(sshRunner, pvPartition); err != nil {
			return err
		}
	}
	return nil
}

func getrootPartition(sshRunner *crcssh.Runner, preset crcPreset.Preset) (string, error) {
	if preset == crcPreset.Microshift {
		return runBlkidQuery(sshRunner, "-t", "TYPE=LVM2_member")
	}
	return runBlkidQuery(sshRunner, "--label", "root")
}

func getTopolvmPartition(sshRunner *crcssh.Runner) (string, error) {
	return runBlkidQuery(sshRunner, "-t", "PARTLABEL=topolvm")
}

func runBlkidQuery(sshRunner *crcssh.Runner, query ...string) (string, error) {
	cmd := []string{"/usr/sbin/blkid", "-o", "device"}
	cmd = append(cmd, query...)
	part, _, err := sshRunner.RunPrivileged("Get device id", cmd...)
	if err != nil {
		return "", err
	}
	parts := strings.Split(strings.TrimSpace(part), "\n")
	if len(parts) != 1 {
		return "", fmt.Errorf("Unexpected number of devices: %s", part)
	}
	part = strings.TrimSpace(parts[0])
	if !strings.HasPrefix(part, "/dev/vda") && !strings.HasPrefix(part, "/dev/sda") {
		return "", fmt.Errorf("Unexpected device: %s", part)
	}
	return part, nil
}

func growLVForMicroshift(sshRunner *crcssh.Runner, lvFullName string, rootPart string, persistentVolumeSize int) error {
	if _, _, err := sshRunner.RunPrivileged("Resizing the physical volume(PV)", "/usr/sbin/pvresize", "--devices", rootPart, rootPart); err != nil {
		return err
	}

	// Get the size of volume group
	sizeVG, _, err := sshRunner.RunPrivileged("Get the volume group size", "/usr/sbin/vgs", "--noheadings", "--nosuffix", "--units", "b", "-o", "vg_size", "--devices", rootPart)
	if err != nil {
		return err
	}
	vgSize, err := strconv.Atoi(strings.TrimSpace(sizeVG))
	if err != nil {
		return err
	}

	// Get the size of root lv
	sizeLV, _, err := sshRunner.RunPrivileged("Get the size of root logical volume", "/usr/sbin/lvs", "-S", fmt.Sprintf("lv_full_name=%s", lvFullName), "--noheadings", "--nosuffix", "--units", "b", "-o", "lv_size", "--devices", rootPart)
	if err != nil {
		return err
	}
	lvSize, err := strconv.Atoi(strings.TrimSpace(sizeLV))
	if err != nil {
		return err
	}

	GB := 1073741824
	vgFree := persistentVolumeSize * GB
	expectedLVSize := vgSize - vgFree
	sizeToIncrease := expectedLVSize - lvSize
	lvPath := fmt.Sprintf("/dev/%s", lvFullName)
	if sizeToIncrease > 1 {
		logging.Info("Extending and resizing '/dev/rhel/root' logical volume")
		if _, _, err := sshRunner.RunPrivileged("Extending and resizing the logical volume(LV)", "/usr/sbin/lvextend", "-L", fmt.Sprintf("+%db", sizeToIncrease), lvPath, "--devices", rootPart); err != nil {
			return err
		}
	}
	return nil
}
