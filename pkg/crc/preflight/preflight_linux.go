package preflight

// StartPreflightChecks performs the preflight checks before starting the cluster
func StartPreflightChecks() {
	preflightCheckSucceedsOrFails("",
		checkVirtualizationEnabled,
		"Checking if Virtualization is enabled",
		"",
	)
	preflightCheckSucceedsOrFails("",
		checkKvmEnabled,
		"Checking if KVM is enabled",
		"",
	)
	preflightCheckSucceedsOrFails("",
		checkLibvirtInstalled,
		"Checking if Libvirt is installed",
		"",
	)
	preflightCheckSucceedsOrFails("",
		checkUserPartOfLibvirtGroup,
		"Checking if user is part of libvirt group",
		"",
	)
	preflightCheckSucceedsOrFails("",
		checkLibvirtEnabled,
		"Checking if Libvirt is enabled",
		"",
	)
	preflightCheckSucceedsOrFails("",
		checkLibvirtServiceRunning,
		"Checking if Libvirt daemon is running",
		"",
	)
	preflightCheckSucceedsOrFails("",
		checkDockerMachineDriverKvmInstalled,
		"Checking if docker-machine-driver-kvm is installed",
		"",
	)
	preflightCheckSucceedsOrFails("",
		checkDefaultPoolAvailable,
		"Checking if default pool is available",
		"",
	)
	preflightCheckSucceedsOrFails("",
		checkDefaultPoolHasSufficientSpace,
		"Checking if default pool has sufficient free space",
		"",
	)
	preflightCheckSucceedsOrFails("",
		checkLibvirtCrcNetworkAvailable,
		"Checking if Libvirt crc network is available",
		"",
	)
	preflightCheckSucceedsOrFails("",
		checkLibvirtCrcNetworkActive,
		"Checking if Libvirt crc network is active",
		"",
	)
}

// SetupHost performs the prerequisite checks and setups the host to run the cluster
func SetupHost() {
	preflightCheckAndFix(checkVirtualizationEnabled,
		fixVirtualizationEnabled,
		"Setting up virtualization",
	)
	preflightCheckAndFix(checkKvmEnabled,
		fixKvmEnabled,
		"Setting up KVM",
	)
	preflightCheckAndFix(checkLibvirtInstalled,
		fixLibvirtInstalled,
		"Installing Libvirt",
	)
	preflightCheckAndFix(checkUserPartOfLibvirtGroup,
		fixUserPartOfLibvirtGroup,
		"Adding user to libvirt group",
	)
	preflightCheckAndFix(checkLibvirtEnabled,
		fixLibvirtEnabled,
		"Enabling libvirt",
	)
	preflightCheckAndFix(checkLibvirtServiceRunning,
		fixLibvirtServiceRunning,
		"Starting Libvirt service",
	)
	preflightCheckAndFix(checkDockerMachineDriverKvmInstalled,
		fixDockerMachineDriverInstalled,
		"Installing docker-machine-driver-kvm",
	)
	preflightCheckAndFix(checkDefaultPoolAvailable,
		fixDefaultPoolAvailable,
		"Creating default storage pool",
	)
	preflightCheckAndFix(checkDefaultPoolHasSufficientSpace,
		fixDefaultPoolHasSufficientSpace,
		"Setting up default pool",
	)
	preflightCheckAndFix(checkLibvirtCrcNetworkAvailable,
		fixLibvirtCrcNetworkAvailable,
		"Setting up Libvirt crc network",
	)
	preflightCheckAndFix(checkLibvirtCrcNetworkActive,
		fixLibvirtCrcNetworkActive,
		"Starting Libvirt crc network",
	)
}
