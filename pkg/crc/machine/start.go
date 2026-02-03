package machine

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"go.podman.io/common/pkg/strongunits"

	"github.com/crc-org/crc/v2/pkg/crc/cloudinit"
	"github.com/crc-org/crc/v2/pkg/crc/cluster"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	crcerrors "github.com/crc-org/crc/v2/pkg/crc/errors"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/macadam"
	"github.com/crc-org/crc/v2/pkg/crc/machine/bundle"
	"github.com/crc-org/crc/v2/pkg/crc/machine/config"
	"github.com/crc-org/crc/v2/pkg/crc/machine/state"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	"github.com/crc-org/crc/v2/pkg/crc/network"
	"github.com/crc-org/crc/v2/pkg/crc/oc"
	crcPreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/crc/services"
	"github.com/crc-org/crc/v2/pkg/crc/services/dns"
	crcssh "github.com/crc-org/crc/v2/pkg/crc/ssh"
	"github.com/crc-org/crc/v2/pkg/crc/systemd"
	"github.com/crc-org/crc/v2/pkg/crc/telemetry"
	"github.com/crc-org/crc/v2/pkg/crc/validation"
	crcos "github.com/crc-org/crc/v2/pkg/os"
	"github.com/docker/go-units"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

const minimumMemoryForMonitoring = strongunits.MiB(14336)

func getCrcBundleInfo(ctx context.Context, preset crcPreset.Preset, bundleName, bundlePath string, enableBundleQuayFallback bool) (*bundle.CrcBundleInfo, error) {
	bundleInfo, err := bundle.Use(bundleName)
	if err == nil {
		logging.Infof("Loading bundle: %s...", bundleName)
		return bundleInfo, nil
	}
	logging.Debugf("Failed to load bundle %s: %v", bundleName, err)
	logging.Infof("Downloading bundle: %s...", bundleName)
	bundlePath, err = bundle.Download(ctx, preset, bundlePath, enableBundleQuayFallback)
	if err != nil {
		return nil, err
	}
	logging.Infof("Extracting bundle: %s...", bundleName)
	if _, err := bundle.Extract(ctx, bundlePath); err != nil {
		return nil, err
	}
	return bundle.Use(bundleName)
}

// updateVMConfig is no longer needed with macadam as VM configuration
// is set during initialization and cannot be changed afterwards
// If configuration changes are needed, the VM must be deleted and recreated

func growRootFileSystem(sshRunner *crcssh.Runner, preset crcPreset.Preset, persistentVolumeSize int) error {
	rootPart, err := getrootPartition(sshRunner, preset)
	if err != nil {
		return err
	}

	// with '/dev/[sv]da4' as input, run 'growpart /dev/[sv]da 4'
	if _, _, err := sshRunner.RunPrivileged(fmt.Sprintf("Growing %s partition", rootPart), "/usr/bin/growpart", rootPart[:len("/dev/.da")], rootPart[len("/dev/.da"):]); err != nil {
		var exitErr *ssh.ExitError
		if !errors.As(err, &exitErr) {
			return err
		}
		if exitErr.ExitStatus() != 1 {
			return err
		}
		logging.Debugf("No free space after %s, nothing to do", rootPart)
		return nil
	}

	if preset == crcPreset.Microshift {
		lvFullName := "rhel/root"
		if err := growLVForMicroshift(sshRunner, lvFullName, rootPart, persistentVolumeSize); err != nil {
			return err
		}
	}

	logging.Infof("Resizing %s filesystem", rootPart)
	rootFS := "/sysroot"
	if _, _, err := sshRunner.RunPrivileged(fmt.Sprintf("Remounting %s read/write", rootFS), "mount -o remount,rw", rootFS); err != nil {
		return err
	}
	if _, _, err = sshRunner.RunPrivileged(fmt.Sprintf("Growing %s filesystem", rootFS), "xfs_growfs", rootFS); err != nil {
		return err
	}

	return nil
}

func getrootPartition(sshRunner *crcssh.Runner, preset crcPreset.Preset) (string, error) {
	diskType := "xfs"
	if preset == crcPreset.Microshift {
		diskType = "LVM2_member"
	}
	part, _, err := sshRunner.RunPrivileged("Get device id", "/usr/sbin/blkid", "-t", fmt.Sprintf("TYPE=%s", diskType), "-o", "device")
	if err != nil {
		return "", err
	}
	parts := strings.Split(strings.TrimSpace(part), "\n")
	if len(parts) != 1 {
		return "", fmt.Errorf("Unexpected number of devices: %s", part)
	}
	rootPart := strings.TrimSpace(parts[0])
	if !strings.HasPrefix(rootPart, "/dev/vda") && !strings.HasPrefix(rootPart, "/dev/sda") {
		return "", fmt.Errorf("Unexpected root device: %s", rootPart)
	}
	return rootPart, nil
}

func growLVForMicroshift(sshRunner crcos.CommandRunner, lvFullName string, rootPart string, persistentVolumeSize int) error {
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

func configureSharedDirs(_ *virtualMachine, _ *crcssh.Runner) error {
	// TODO: Implement shared directory configuration for macadam
	// For now, shared directories are not supported with macadam
	logging.Debug("Shared directory configuration not yet implemented for macadam")
	return nil
}

func (client *client) Start(ctx context.Context, startConfig types.StartConfig) (*types.StartResult, error) {
	telemetry.SetCPUs(ctx, startConfig.CPUs)
	telemetry.SetMemory(ctx, uint64(startConfig.Memory.ToBytes()))
	telemetry.SetDiskSize(ctx, uint64(startConfig.DiskSize.ToBytes()))

	if err := client.validateStartConfig(startConfig); err != nil {
		return nil, err
	}

	// Pre-VM start
	exists, err := client.Exists()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot determine if VM exists")
	}

	bundleNameFromURI, err := bundle.GetBundleNameFromURI(startConfig.BundlePath)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting bundle name")
	}
	bundleName := bundle.GetBundleNameWithoutExtension(bundleNameFromURI)
	crcBundleMetadata, err := getCrcBundleInfo(ctx, startConfig.Preset, bundleName, startConfig.BundlePath, startConfig.EnableBundleQuayFallback)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting bundle metadata")
	}

	if err := validation.BundleMismatchWithPresetMetadata(startConfig.Preset, crcBundleMetadata); err != nil {
		return nil, err
	}

	if !exists {
		telemetry.SetStartType(ctx, telemetry.CreationStartType)

		// Ask early for pull secret if it hasn't been requested yet
		_, err = startConfig.PullSecret.Value()
		if err != nil {
			return nil, errors.Wrap(err, "Failed to ask for pull secret")
		}

		logging.Infof("Creating CRC VM for %s %s...", startConfig.Preset.ForDisplay(), crcBundleMetadata.GetVersion())

		sharedDirs := []string{}
		if homeDir, err := os.UserHomeDir(); err == nil {
			sharedDirs = append(sharedDirs, homeDir)
		}

		machineConfig := config.MachineConfig{
			Name:              client.name,
			BundleName:        bundleName,
			CPUs:              startConfig.CPUs,
			Memory:            startConfig.Memory,
			DiskSize:          startConfig.DiskSize,
			NetworkMode:       client.networkMode(),
			ImageSourcePath:   crcBundleMetadata.GetDiskImagePath(),
			ImageFormat:       crcBundleMetadata.GetDiskImageFormat(),
			SSHKeyPath:        crcBundleMetadata.GetSSHKeyPath(),
			SharedDirs:        sharedDirs,
			SharedDirPassword: startConfig.SharedDirPassword,
			SharedDirUsername: startConfig.SharedDirUsername,
		}
		if crcBundleMetadata.IsOpenShift() {
			machineConfig.KubeConfig = crcBundleMetadata.GetKubeConfigPath()
		}
		if err := createHost(machineConfig, crcBundleMetadata.GetBundleType(), startConfig.PullSecret); err != nil {
			return nil, errors.Wrap(err, "Error creating machine")
		}
	} else {
		telemetry.SetStartType(ctx, telemetry.StartStartType)
	}

	vm, err := loadVirtualMachine(client.name, client.useVSock())
	if err != nil {
		return nil, errors.Wrap(err, "Error loading machine")
	}
	defer vm.Close()

	currentBundleName := vm.bundle.GetBundleName()
	if currentBundleName != bundleName {
		logging.Debugf("Bundle '%s' was requested, but the existing VM is using '%s'",
			bundleName, currentBundleName)
		return nil, fmt.Errorf("Bundle '%s' was requested, but the existing VM is using '%s'. Please delete your existing cluster and start again",
			bundleName,
			currentBundleName)
	}
	vmState, err := vm.State()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting the machine state")
	}
	if vmState == state.Running {
		logging.Infof("A CRC VM for %s %s is already running", startConfig.Preset.ForDisplay(), vm.bundle.GetVersion())
		clusterConfig, err := getClusterConfig(vm.bundle)
		if err != nil {
			return nil, errors.Wrap(err, "Cannot create cluster configuration")
		}

		telemetry.SetStartType(ctx, telemetry.AlreadyRunningStartType)
		return &types.StartResult{
			Status:         vmState,
			ClusterConfig:  *clusterConfig,
			KubeletStarted: true,
		}, nil
	}

	if _, err := bundle.Use(currentBundleName); err != nil {
		return nil, err
	}

	logging.Infof("Starting CRC VM for %s %s...", startConfig.Preset, vm.bundle.GetVersion())

	// Note: With macadam, VM configuration is set during init and cannot be updated afterwards
	// If config changes are needed, the VM must be recreated

	if err := startHost(ctx, vm.name); err != nil {
		return nil, errors.Wrap(err, "Error starting machine")
	}

	// Post-VM start
	vmState, err = vm.State()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting the state")
	}
	if vmState != state.Running {
		return nil, errors.Wrap(err, "CRC VM is not running")
	}

	instanceIP, err := vm.IP()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting the IP")
	}
	logging.Infof("CRC instance is running with IP %s", instanceIP)

	// Configure internal DNS if using vsock/user-mode networking
	if client.useVSock() {
		gvClient, err := getGVProxyClient(vm.name)
		if err != nil {
			return nil, errors.Wrap(err, "Error getting gvproxy client")
		}
		if err := enableInternalDNS(gvClient); err != nil {
			logging.Warnf("Failed to configure internal DNS: %v", err)
			// Don't fail startup if DNS configuration fails, just warn
		}
		if err := exposePorts(gvClient, startConfig.Preset, startConfig.IngressHTTPPort, startConfig.IngressHTTPSPort); err != nil {
			logging.Warnf("Failed to expose ports: %v", err)
			// Don't fail startup if port exposure fails, just warn
		}
	}

	sshRunner, err := vm.SSHRunner()
	if err != nil {
		return nil, errors.Wrap(err, "Error creating the ssh client")
	}
	defer sshRunner.Close()

	logging.Debug("Waiting until ssh is available")
	if err := sshRunner.WaitForConnectivity(ctx, 300*time.Second); err != nil {
		return nil, errors.Wrap(err, "Failed to connect to the CRC VM with SSH -- virtual machine might be unreachable")
	}
	logging.Info("CRC VM is running")

	if startConfig.EmergencyLogin {
		if err := enableEmergencyLogin(sshRunner); err != nil {
			return nil, errors.Wrap(err, "Error enabling emergency login")
		}
	} else {
		if err := disableEmergencyLogin(sshRunner); err != nil {
			return nil, errors.Wrap(err, "Error deleting the password for core user")
		}
	}

	// Post VM start immediately update SSH key and copy kubeconfig to instance
	// dir and VM
	if err := updateSSHKeyPair(sshRunner); err != nil {
		return nil, errors.Wrap(err, "Error updating public key")
	}

	// Trigger disk resize, this will be a no-op if no disk size change is needed
	if err := growRootFileSystem(sshRunner, startConfig.Preset, startConfig.PersistentVolumeSize); err != nil {
		return nil, errors.Wrap(err, "Error updating filesystem size")
	}

	// Start network time synchronization if `CRC_DEBUG_ENABLE_STOP_NTP` is not set
	if stopNtp, _ := strconv.ParseBool(os.Getenv("CRC_DEBUG_ENABLE_STOP_NTP")); stopNtp {
		logging.Info("Stopping network time synchronization in CRC VM")
		if _, _, err := sshRunner.RunPrivileged("Turning off the ntp server", "timedatectl set-ntp off"); err != nil {
			return nil, errors.Wrap(err, "Failed to stop network time synchronization")
		}
		if _, _, err := sshRunner.RunPrivileged("Manual mode for chrony config", "tee /etc/chrony.conf <<< \"manual\""); err != nil {
			return nil, errors.Wrap(err, "Failed to update manual mode for /etc/chrony.conf")
		}
		logging.Info("Setting clock to vm clock (UTC timezone)")
		dateCmd := fmt.Sprintf("date -s '%s'", time.Now().Format(time.UnixDate))
		if _, _, err := sshRunner.RunPrivileged("Setting clock same as host", dateCmd); err != nil {
			return nil, errors.Wrap(err, "Failed to set clock to same as host")
		}
	}

	// Add nameserver to VM if provided by User
	if startConfig.NameServer != "" {
		if err = addNameServerToInstance(sshRunner, startConfig.NameServer); err != nil {
			return nil, errors.Wrap(err, "Failed to add nameserver to the VM")
		}
	}
	if startConfig.EnableSharedDirs {
		if err := configureSharedDirs(vm, sshRunner); err != nil {
			return nil, err
		}
	}

	if _, _, err := sshRunner.RunPrivileged("make root Podman socket accessible", "chmod 777 /run/podman/ /run/podman/podman.sock"); err != nil {
		return nil, errors.Wrap(err, "Failed to change permissions to root podman socket")
	}

	proxyConfig, err := getProxyConfig(vm.bundle)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting proxy configuration")
	}

	proxyConfig.ApplyToEnvironment()
	proxyConfig.AddNoProxy(instanceIP)

	// Create servicePostStartConfig for DNS checks and DNS start.
	servicePostStartConfig := services.ServicePostStartConfig{
		Name: client.name,
		// TODO: would prefer passing in a more generic type
		SSHRunner: sshRunner,
		IP:        instanceIP,
		// TODO: should be more finegrained
		BundleMetadata:  *vm.bundle,
		NetworkMode:     client.networkMode(),
		ModifyHostsFile: client.modifyHostsFile(),
	}

	// Run the DNS server inside the VM
	if err := dns.RunPostStart(servicePostStartConfig); err != nil {
		return nil, errors.Wrap(err, "Error running post start")
	}

	// Check DNS lookup before starting the kubelet
	logging.Info("Check internal and public DNS query...")
	if !client.useVSock() {
		if queryOutput, err := dns.CheckCRCLocalDNSReachable(ctx, servicePostStartConfig); err != nil {
			return nil, errors.Wrapf(err, "Failed internal DNS query: %s", queryOutput)
		}
	}

	if queryOutput, err := dns.CheckCRCPublicDNSReachable(servicePostStartConfig); err != nil {
		logging.Warnf("Failed public DNS query from the cluster: %v : %s", err, queryOutput)
	}

	// Check DNS lookup from host to VM
	logging.Info("Check DNS query from host...")
	if err := dns.CheckCRCLocalDNSReachableFromHost(servicePostStartConfig); err != nil {
		if !client.useVSock() {
			msg := "Failed to query DNS from host"
			if !servicePostStartConfig.ModifyHostsFile {
				msg += " (modify-hosts-file=false). Ensure your system DNS/hosts entries resolve the CRC domains."
			}
			return nil, errors.Wrap(err, msg)
		}
		logging.Warn(fmt.Sprintf("Failed to query DNS from host: %v", err))
	}

	if vm.bundle.IsMicroshift() {
		// **************************
		//  END OF MICROSHIFT START CODE
		// **************************
		// Start the microshift and copy the generated kubeconfig file
		ocConfig := oc.UseOCWithSSH(sshRunner)
		ocConfig.Context = "microshift"
		ocConfig.Cluster = "microshift"

		if err := startMicroshift(ctx, sshRunner, ocConfig, startConfig.PullSecret); err != nil {
			return nil, err
		}

		if client.useVSock() {
			if err := ensureRoutesControllerIsRunning(sshRunner, ocConfig); err != nil {
				return nil, err
			}
		}
		logging.Info("Adding microshift context to kubeconfig...")
		if err := mergeKubeConfigFile(constants.KubeconfigFilePath); err != nil {
			return nil, err
		}

		return &types.StartResult{
			ClusterConfig: types.ClusterConfig{ClusterType: startConfig.Preset},
			Status:        vmState,
		}, nil
	}

	// Check the certs validity inside the vm
	logging.Info("Verifying validity of the kubelet certificates...")
	certsExpired, err := cluster.CheckCertsValidity(sshRunner)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to check certificate validity")
	}

	ocConfig := oc.UseOCWithSSH(sshRunner)

	if err := cluster.WaitForAPIServer(ctx, ocConfig); err != nil {
		return nil, errors.Wrap(err, "Error waiting for apiserver")
	}

	if err := cluster.ApproveCSRAndWaitForCertsRenewal(ctx, sshRunner, ocConfig, certsExpired[cluster.KubeletClientCert], certsExpired[cluster.KubeletServerCert], certsExpired[cluster.AggregatorClientCert]); err != nil {
		logBundleDate(vm.bundle)
		return nil, errors.Wrap(err, "Failed to renew TLS certificates: please check if a newer CRC release is available")
	}

	if err := cluster.DeleteMCOLeaderLease(ctx, ocConfig); err != nil {
		return nil, err
	}
	systemdRunner := systemd.NewInstanceSystemdCommander(sshRunner)

	if err := cluster.EnsurePullSecretPresentInTheCluster(ctx, systemdRunner, ocConfig, startConfig.PullSecret); err != nil {
		return nil, errors.Wrap(err, "Failed to update cluster pull secret")
	}

	if err := cluster.EnsureSSHKeyPresentInTheCluster(ctx, ocConfig, constants.GetPublicKeyPath()); err != nil {
		return nil, errors.Wrap(err, "Failed to update ssh public key to machine config")
	}

	if err := cluster.EnsureClusterIDIsNotEmpty(ctx, systemdRunner, ocConfig); err != nil {
		return nil, errors.Wrap(err, "Failed to update cluster ID")
	}

	if client.useVSock() {
		if err := ensureRoutesControllerIsRunning(sshRunner, ocConfig); err != nil {
			return nil, err
		}
	}

	if client.monitoringEnabled() {
		logging.Info("Enabling cluster monitoring operator...")
		if err := cluster.StartMonitoring(ocConfig); err != nil {
			return nil, errors.Wrap(err, "Cannot start monitoring stack")
		}
	}

	if err := copyKubeconfigFileFromVMToHost(ctx, systemdRunner, sshRunner, constants.KubeconfigFilePath); err != nil {
		return nil, errors.Wrap(err, "Failed to update kubeconfig file")
	}

	logging.Infof("Starting %s instance... [waiting for the cluster to stabilize]", startConfig.Preset)
	if err := cluster.WaitForClusterStable(ctx, instanceIP, constants.KubeconfigFilePath, proxyConfig); err != nil {
		logging.Warnf("Cluster is not ready: %v", err)
	}

	if err := cluster.WaitForPullSecretPresentOnInstanceDisk(ctx, sshRunner); err != nil {
		return nil, errors.Wrap(err, "Failed to update pull secret on the disk")
	}

	/*waitForProxyPropagation(ctx, ocConfig, proxyConfig)*/

	clusterConfig, err := getClusterConfig(vm.bundle)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get cluster configuration")
	}

	logging.Info("Adding crc-admin and crc-developer contexts to kubeconfig...")
	if err := writeKubeconfig(instanceIP, clusterConfig, startConfig.IngressHTTPSPort); err != nil {
		logging.Errorf("Cannot update kubeconfig: %v", err)
	}

	return &types.StartResult{
		KubeletStarted: true,
		ClusterConfig:  *clusterConfig,
		Status:         vmState,
	}, nil
}

func (client *client) IsRunning() (bool, error) {
	// Check if VM exists first
	exists, err := vmExists(client.name)
	if err != nil {
		return false, errors.Wrap(err, "Cannot check if machine exists")
	}
	if !exists {
		return false, nil
	}

	// get the actual state
	vmState, err := getVMState(client.name)
	if err != nil {
		// but reports not started on error
		return false, errors.Wrap(err, "Error getting the state")
	}
	if vmState != state.Running {
		return false, nil
	}
	return true, nil
}

func (client *client) validateStartConfig(startConfig types.StartConfig) error {
	if client.monitoringEnabled() && startConfig.Memory < minimumMemoryForMonitoring {
		return fmt.Errorf("Too little memory (%s) allocated to the virtual machine to start the monitoring stack, %s is the minimum",
			units.BytesSize(float64(startConfig.Memory.ToBytes())),
			units.BytesSize(float64(minimumMemoryForMonitoring.ToBytes())))
	}
	return nil
}

func createHost(machineConfig config.MachineConfig, preset crcPreset.Preset, pullSecret cluster.PullSecretLoader) error {
	logging.Info("Generating new SSH key pair...")
	if err := crcssh.GenerateSSHKey(constants.GetPrivateKeyPath()); err != nil {
		return fmt.Errorf("Error generating ssh key pair: %v", err)
	}

	// Read the public key for cloud-init
	pubKeyBytes, err := os.ReadFile(constants.GetPublicKeyPath())
	if err != nil {
		return fmt.Errorf("Error reading public key: %v", err)
	}
	publicKey := strings.TrimSpace(string(pubKeyBytes))

	// Generate passwords for OpenShift/OKD
	var kubeAdminPassword, developerPassword string
	if preset == crcPreset.OpenShift || preset == crcPreset.OKD {
		if err := cluster.GenerateUserPassword(constants.GetKubeAdminPasswordPath(), "kubeadmin"); err != nil {
			return errors.Wrap(err, "Error generating new kubeadmin password")
		}
		kubeAdminPassBytes, err := os.ReadFile(constants.GetKubeAdminPasswordPath())
		if err != nil {
			return errors.Wrap(err, "Error reading kubeadmin password")
		}
		kubeAdminPassword = strings.TrimSpace(string(kubeAdminPassBytes))

		developerPassword = constants.DefaultDeveloperPassword
		if err = os.WriteFile(constants.GetDeveloperPasswordPath(), []byte(developerPassword), 0o600); err != nil {
			return errors.Wrap(err, "Error writing developer password")
		}
	}

	// pull secret is not present in VM or is invalid
	content, err := pullSecret.Value()
	if err != nil {
		return errors.Wrap(err, "Error getting pull secret")
	}

	// Generate cloud-init user-data
	logging.Debug("Generating cloud-init user-data...")
	cloudInitOpts := cloudinit.UserDataOptions{
		PublicKey:         publicKey,
		PullSecret:        content,
		KubeAdminPassword: kubeAdminPassword,
		DeveloperPassword: developerPassword,
	}

	userDataPath, err := cloudinit.GenerateUserData(machineConfig.Name, cloudInitOpts)
	if err != nil {
		return fmt.Errorf("Error generating cloud-init user-data: %v", err)
	}

	// Prepare VM options for macadam
	vmOpts := macadam.VMOptions{
		DiskImagePath:   machineConfig.ImageSourcePath,
		DiskSize:        uint64(machineConfig.DiskSize),
		Memory:          uint64(machineConfig.Memory),
		Name:            machineConfig.Name,
		Username:        "core",
		SSHIdentityPath: constants.GetPrivateKeyPath(),
		CPUs:            uint64(machineConfig.CPUs),
		CloudInitPath:   userDataPath,
	}

	// Initialize macadam and create the VM
	logging.Debug("Creating machine with macadam...")
	m := macadam.UseMacadam()
	stdout, stderr, err := m.InitVM(vmOpts)
	if err != nil {
		return fmt.Errorf("Error in macadam during machine creation: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Save bundle metadata to config.json so we can retrieve it later
	if err := saveBundleMetadataToConfig(machineConfig.Name, machineConfig.BundleName); err != nil {
		logging.Warnf("Failed to save bundle metadata: %v", err)
		// Non-fatal error, continue
	}

	logging.Debug("Machine successfully created")
	return nil
}

func startHost(ctx context.Context, vmName string) error {
	m := getMacadamClient()
	_, stdErr, err := m.StartVM(vmName)
	fmt.Println("stdErr", stdErr)
	// TODO: Ignoring error for now, we need to handle this better
	// https://github.com/cfergeau/podman/pull/24
	if err != nil && !strings.Contains(stdErr, "ssh error: ssh: handshake failed:") {
		return fmt.Errorf("Error in driver during machine start: %s", err)
	}

	logging.Debug("Waiting for machine to be running, this may take a few minutes...")
	if err := crcerrors.Retry(ctx, 3*time.Minute, func() error {
		vmState, err := getVMState(vmName)
		if err != nil {
			return err
		}
		if vmState != state.Running {
			return fmt.Errorf("machine not running yet, current state: %s", vmState)
		}
		return nil
	}, 3*time.Second); err != nil {
		return fmt.Errorf("Error waiting for machine to be running: %s", err)
	}

	logging.Debug("Machine is up and running!")

	return nil
}

func addNameServerToInstance(sshRunner *crcssh.Runner, ns string) error {
	nameserver := network.NameServer{IPAddress: ns}
	nameservers := []network.NameServer{nameserver}
	exist, err := network.HasGivenNameserversConfigured(sshRunner, nameserver)
	if err != nil {
		return err
	}
	if !exist {
		logging.Infof("Adding %s as nameserver to the instance...", nameserver.IPAddress)
		return network.AddNameserversToInstance(sshRunner, nameservers)
	}
	return nil
}

func enableEmergencyLogin(sshRunner *crcssh.Runner) error {
	if crcos.FileExists(constants.PasswdFilePath) {
		return nil
	}
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))] //nolint:gosec
	}
	if err := os.WriteFile(constants.PasswdFilePath, b, 0o600); err != nil {
		return err
	}
	logging.Infof("Emergency login password for core user is stored to %s", constants.PasswdFilePath)
	_, _, err := sshRunner.Run(fmt.Sprintf("sudo passwd core -f --unlock && echo %s | sudo passwd core --stdin", b))
	return err
}

func disableEmergencyLogin(sshRunner *crcssh.Runner) error {
	defer os.Remove(constants.PasswdFilePath)
	_, _, err := sshRunner.RunPrivileged("disable core user password", "passwd", "--lock", "core")
	return err
}

func updateSSHKeyPair(sshRunner *crcssh.Runner) error {
	// Read generated public key
	publicKey, err := os.ReadFile(constants.GetPublicKeyPath())
	if err != nil {
		return err
	}

	authorizedKeys, _, err := sshRunner.Run("cat /home/core/.ssh/authorized_keys")
	if err == nil && strings.TrimSpace(authorizedKeys) == strings.TrimSpace(string(publicKey)) {
		return nil
	}

	logging.Info("Updating authorized keys...")
	err = sshRunner.CopyData(publicKey, "/home/core/.ssh/authorized_keys", 0o644)
	if err != nil {
		return err
	}

	/* This is specific to the podman bundle, but is required to drop the 'default' ssh key */
	_, _, _ = sshRunner.Run("rm", "/home/core/.ssh/authorized_keys.d/ignition")
	return nil
}

func logBundleDate(crcBundleMetadata *bundle.CrcBundleInfo) {
	if buildTime, err := crcBundleMetadata.GetBundleBuildTime(); err == nil {
		bundleAgeDays := time.Since(buildTime).Hours() / 24
		if bundleAgeDays >= 30 {
			/* Initial bundle certificates are only valid for 30 days */
			logging.Debugf("Bundle has been generated %d days ago", int(bundleAgeDays))
		}
	}
}

func ensureRoutesControllerIsRunning(sshRunner *crcssh.Runner, ocConfig oc.Config) error {
	// Check if the bundle have `/opt/crc/routes-controller.yaml` file and if it has
	// then use it to create the resource for the routes controller.
	_, _, err := sshRunner.Run("ls", "/opt/crc/routes-controller.yaml")
	if err != nil {
		return err
	}
	_, _, err = ocConfig.RunOcCommand("apply", "-f", "/opt/crc/routes-controller.yaml")
	return err
}

func copyKubeconfigFileFromVMToHost(ctx context.Context, systemdRunner *systemd.Commander, sshRunner *crcssh.Runner, kubeconfigFilePath string) error {
	logging.Info("Waiting for the updated kubeconfig file to be available on the host...")
	if err := cluster.WaitForServiceSuccessfullyFinished(ctx, systemdRunner, "ocp-cluster-ca.service", 180*time.Second, 2*time.Second); err != nil {
		return err
	}
	if err := sshRunner.CopyFileFromVM("/opt/kubeconfig", kubeconfigFilePath, 0o600); err != nil {
		return err
	}
	return nil
}

func startMicroshift(ctx context.Context, sshRunner *crcssh.Runner, ocConfig oc.Config, pullSec cluster.PullSecretLoader) error {
	logging.Infof("Starting Microshift service... [takes around 1min]")
	if err := ensurePullSecretPresentInVM(sshRunner, pullSec); err != nil {
		return err
	}
	if _, _, err := sshRunner.RunPrivileged("Starting microshift service", "systemctl", "start", "microshift"); err != nil {
		return err
	}
	if err := sshRunner.CopyFileFromVM(fmt.Sprintf("/var/lib/microshift/resources/kubeadmin/api%s/kubeconfig", constants.ClusterDomain), constants.KubeconfigFilePath, 0o600); err != nil {
		return err
	}
	if err := sshRunner.CopyFile(constants.KubeconfigFilePath, "/opt/kubeconfig", 0o644); err != nil {
		return err
	}

	return cluster.WaitForAPIServer(ctx, ocConfig)
}

func ensurePullSecretPresentInVM(sshRunner *crcssh.Runner, pullSec cluster.PullSecretLoader) error {
	if pullSecret, _, err := sshRunner.RunPrivate("sudo", "cat", "/etc/crio/openshift-pull-secret"); err == nil {
		if err := validation.ImagePullSecret(pullSecret); err == nil {
			return nil
		}
	}
	// pull secret is not present in VM or is invalid
	content, err := pullSec.Value()
	if err != nil {
		return err
	}
	return sshRunner.CopyDataPrivileged([]byte(content), "/etc/crio/openshift-pull-secret", 0o600)
}
