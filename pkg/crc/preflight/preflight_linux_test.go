package preflight

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/network"
	crcos "github.com/code-ready/crc/pkg/os/linux"
	"github.com/stretchr/testify/assert"
)

func TestCountConfigurationOptions(t *testing.T) {
	cfg := config.New(config.NewEmptyInMemoryStorage())
	RegisterSettings(cfg)
	options := len(cfg.AllConfigs())

	var preflightChecksCount int
	for _, check := range getAllPreflightChecks() {
		if check.configKeySuffix != "" {
			preflightChecksCount++
		}
	}
	assert.True(t, options == preflightChecksCount, "Unexpected number of preflight configuration flags, got %d, expected %d", options, preflightChecksCount)
}

var (
	fedora = crcos.OsRelease{
		ID:        crcos.Fedora,
		VersionID: "35",
	}

	rhel = crcos.OsRelease{
		ID:        crcos.RHEL,
		VersionID: "8.3",
		IDLike:    string(crcos.Fedora),
	}

	ubuntu = crcos.OsRelease{
		ID:        crcos.Ubuntu,
		VersionID: "20.04",
	}

	unexpected = crcos.OsRelease{
		ID:        "unexpected",
		VersionID: "1234",
	}
)

type checkListForDistro struct {
	distro          *crcos.OsRelease
	networkMode     network.Mode
	systemdResolved bool
	checks          []Check
}

var checkListForDistros = []checkListForDistro{
	{
		distro:          &fedora,
		networkMode:     network.SystemNetworkingMode,
		systemdResolved: true,
		checks: []Check{
			{check: checkIfRunningAsNormalUser},
			{check: checkRunningInsideWSL2},
			{check: checkAdminHelperExecutableCached},
			{check: checkOldAdminHelperExecutableCached},
			{check: checkSupportedCPUArch},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{cleanup: removeOldLogs},
			{cleanup: cluster.ForgetPullSecret},
			{cleanup: removeHostsFileEntry},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
			{configKeySuffix: "check-libvirt-group-active"},
			{check: checkLibvirtServiceRunning},
			{check: checkLibvirtVersion},
			{check: checkMachineDriverLibvirtInstalled},
			{cleanup: removeCrcVM},
			{check: checkSystemdNetworkdIsNotRunning},
			{check: checkNetworkManagerInstalled},
			{check: checkNetworkManagerIsRunning},
			{check: checkCrcDnsmasqAndNetworkManagerConfigFile},
			{check: checkSystemdResolvedIsRunning},
			{check: checkCrcNetworkManagerDispatcherFile},
			{check: checkLibvirtCrcNetworkAvailable},
			{check: checkLibvirtCrcNetworkActive},
			{check: checkBundleExtracted},
		},
	},
	{
		distro:          &fedora,
		networkMode:     network.SystemNetworkingMode,
		systemdResolved: false,
		checks: []Check{
			{check: checkIfRunningAsNormalUser},
			{check: checkRunningInsideWSL2},
			{check: checkAdminHelperExecutableCached},
			{check: checkOldAdminHelperExecutableCached},
			{check: checkSupportedCPUArch},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{cleanup: removeOldLogs},
			{cleanup: cluster.ForgetPullSecret},
			{cleanup: removeHostsFileEntry},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
			{configKeySuffix: "check-libvirt-group-active"},
			{check: checkLibvirtServiceRunning},
			{check: checkLibvirtVersion},
			{check: checkMachineDriverLibvirtInstalled},
			{cleanup: removeCrcVM},
			{check: checkSystemdNetworkdIsNotRunning},
			{check: checkNetworkManagerInstalled},
			{check: checkNetworkManagerIsRunning},
			{check: checkCrcNetworkManagerConfig},
			{check: checkCrcDnsmasqConfigFile},
			{check: checkLibvirtCrcNetworkAvailable},
			{check: checkLibvirtCrcNetworkActive},
			{check: checkBundleExtracted},
		},
	},
	{
		distro:          &fedora,
		networkMode:     network.UserNetworkingMode,
		systemdResolved: false,
		checks: []Check{
			{check: checkIfRunningAsNormalUser},
			{check: checkRunningInsideWSL2},
			{check: checkAdminHelperExecutableCached},
			{check: checkOldAdminHelperExecutableCached},
			{check: checkSupportedCPUArch},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{cleanup: removeOldLogs},
			{cleanup: cluster.ForgetPullSecret},
			{cleanup: removeHostsFileEntry},
			{check: checkCrcSymlink},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
			{configKeySuffix: "check-libvirt-group-active"},
			{check: checkLibvirtServiceRunning},
			{check: checkLibvirtVersion},
			{check: checkMachineDriverLibvirtInstalled},
			{cleanup: removeCrcVM},
			{check: checkDaemonSystemdService},
			{check: checkDaemonSystemdSockets},
			{check: checkVsock},
			{check: checkBundleExtracted},
		},
	},
	{
		distro:          &rhel,
		networkMode:     network.SystemNetworkingMode,
		systemdResolved: true,
		checks: []Check{
			{check: checkIfRunningAsNormalUser},
			{check: checkRunningInsideWSL2},
			{check: checkAdminHelperExecutableCached},
			{check: checkOldAdminHelperExecutableCached},
			{check: checkSupportedCPUArch},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{cleanup: removeOldLogs},
			{cleanup: cluster.ForgetPullSecret},
			{cleanup: removeHostsFileEntry},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
			{configKeySuffix: "check-libvirt-group-active"},
			{check: checkLibvirtServiceRunning},
			{check: checkLibvirtVersion},
			{check: checkMachineDriverLibvirtInstalled},
			{cleanup: removeCrcVM},
			{check: checkSystemdNetworkdIsNotRunning},
			{check: checkNetworkManagerInstalled},
			{check: checkNetworkManagerIsRunning},
			{check: checkCrcDnsmasqAndNetworkManagerConfigFile},
			{check: checkSystemdResolvedIsRunning},
			{check: checkCrcNetworkManagerDispatcherFile},
			{check: checkLibvirtCrcNetworkAvailable},
			{check: checkLibvirtCrcNetworkActive},
			{check: checkBundleExtracted},
		},
	},
	{
		distro:          &rhel,
		networkMode:     network.SystemNetworkingMode,
		systemdResolved: false,
		checks: []Check{
			{check: checkIfRunningAsNormalUser},
			{check: checkRunningInsideWSL2},
			{check: checkAdminHelperExecutableCached},
			{check: checkOldAdminHelperExecutableCached},
			{check: checkSupportedCPUArch},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{cleanup: removeOldLogs},
			{cleanup: cluster.ForgetPullSecret},
			{cleanup: removeHostsFileEntry},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
			{configKeySuffix: "check-libvirt-group-active"},
			{check: checkLibvirtServiceRunning},
			{check: checkLibvirtVersion},
			{check: checkMachineDriverLibvirtInstalled},
			{cleanup: removeCrcVM},
			{check: checkSystemdNetworkdIsNotRunning},
			{check: checkNetworkManagerInstalled},
			{check: checkNetworkManagerIsRunning},
			{check: checkCrcNetworkManagerConfig},
			{check: checkCrcDnsmasqConfigFile},
			{check: checkLibvirtCrcNetworkAvailable},
			{check: checkLibvirtCrcNetworkActive},
			{check: checkBundleExtracted},
		},
	},
	{
		distro:          &rhel,
		networkMode:     network.UserNetworkingMode,
		systemdResolved: false,
		checks: []Check{
			{check: checkIfRunningAsNormalUser},
			{check: checkRunningInsideWSL2},
			{check: checkAdminHelperExecutableCached},
			{check: checkOldAdminHelperExecutableCached},
			{check: checkSupportedCPUArch},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{cleanup: removeOldLogs},
			{cleanup: cluster.ForgetPullSecret},
			{cleanup: removeHostsFileEntry},
			{check: checkCrcSymlink},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
			{configKeySuffix: "check-libvirt-group-active"},
			{check: checkLibvirtServiceRunning},
			{check: checkLibvirtVersion},
			{check: checkMachineDriverLibvirtInstalled},
			{cleanup: removeCrcVM},
			{check: checkDaemonSystemdService},
			{check: checkDaemonSystemdSockets},
			{check: checkVsock},
			{check: checkBundleExtracted},
		},
	},
	{
		distro:          &unexpected,
		networkMode:     network.SystemNetworkingMode,
		systemdResolved: true,
		checks: []Check{
			{check: checkIfRunningAsNormalUser},
			{check: checkRunningInsideWSL2},
			{check: checkAdminHelperExecutableCached},
			{check: checkOldAdminHelperExecutableCached},
			{check: checkSupportedCPUArch},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{cleanup: removeOldLogs},
			{cleanup: cluster.ForgetPullSecret},
			{cleanup: removeHostsFileEntry},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
			{configKeySuffix: "check-libvirt-group-active"},
			{check: checkLibvirtServiceRunning},
			{check: checkLibvirtVersion},
			{check: checkMachineDriverLibvirtInstalled},
			{cleanup: removeCrcVM},
			{check: checkSystemdNetworkdIsNotRunning},
			{check: checkNetworkManagerInstalled},
			{check: checkNetworkManagerIsRunning},
			{check: checkCrcDnsmasqAndNetworkManagerConfigFile},
			{check: checkSystemdResolvedIsRunning},
			{check: checkCrcNetworkManagerDispatcherFile},
			{check: checkLibvirtCrcNetworkAvailable},
			{check: checkLibvirtCrcNetworkActive},
			{check: checkBundleExtracted},
		},
	},
	{
		distro:          &unexpected,
		networkMode:     network.SystemNetworkingMode,
		systemdResolved: false,
		checks: []Check{
			{check: checkIfRunningAsNormalUser},
			{check: checkRunningInsideWSL2},
			{check: checkAdminHelperExecutableCached},
			{check: checkOldAdminHelperExecutableCached},
			{check: checkSupportedCPUArch},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{cleanup: removeOldLogs},
			{cleanup: cluster.ForgetPullSecret},
			{cleanup: removeHostsFileEntry},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
			{configKeySuffix: "check-libvirt-group-active"},
			{check: checkLibvirtServiceRunning},
			{check: checkLibvirtVersion},
			{check: checkMachineDriverLibvirtInstalled},
			{cleanup: removeCrcVM},
			{check: checkSystemdNetworkdIsNotRunning},
			{check: checkNetworkManagerInstalled},
			{check: checkNetworkManagerIsRunning},
			{check: checkCrcNetworkManagerConfig},
			{check: checkCrcDnsmasqConfigFile},
			{check: checkLibvirtCrcNetworkAvailable},
			{check: checkLibvirtCrcNetworkActive},
			{check: checkBundleExtracted},
		},
	},
	{
		distro:          &unexpected,
		networkMode:     network.UserNetworkingMode,
		systemdResolved: false,
		checks: []Check{
			{check: checkIfRunningAsNormalUser},
			{check: checkRunningInsideWSL2},
			{check: checkAdminHelperExecutableCached},
			{check: checkOldAdminHelperExecutableCached},
			{check: checkSupportedCPUArch},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{cleanup: removeOldLogs},
			{cleanup: cluster.ForgetPullSecret},
			{cleanup: removeHostsFileEntry},
			{check: checkCrcSymlink},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
			{configKeySuffix: "check-libvirt-group-active"},
			{check: checkLibvirtServiceRunning},
			{check: checkLibvirtVersion},
			{check: checkMachineDriverLibvirtInstalled},
			{cleanup: removeCrcVM},
			{check: checkDaemonSystemdService},
			{check: checkDaemonSystemdSockets},
			{check: checkVsock},
			{check: checkBundleExtracted},
		},
	},
	{
		distro:          &ubuntu,
		networkMode:     network.SystemNetworkingMode,
		systemdResolved: true,
		checks: []Check{
			{check: checkIfRunningAsNormalUser},
			{check: checkRunningInsideWSL2},
			{check: checkAdminHelperExecutableCached},
			{check: checkOldAdminHelperExecutableCached},
			{check: checkSupportedCPUArch},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{cleanup: removeOldLogs},
			{cleanup: cluster.ForgetPullSecret},
			{cleanup: removeHostsFileEntry},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
			{configKeySuffix: "check-libvirt-group-active"},
			{check: checkLibvirtServiceRunning},
			{check: checkLibvirtVersion},
			{check: checkMachineDriverLibvirtInstalled},
			{cleanup: removeCrcVM},
			{configKeySuffix: "check-apparmor-profile-setup"},
			{check: checkSystemdNetworkdIsNotRunning},
			{check: checkNetworkManagerInstalled},
			{check: checkNetworkManagerIsRunning},
			{check: checkCrcDnsmasqAndNetworkManagerConfigFile},
			{check: checkSystemdResolvedIsRunning},
			{check: checkCrcNetworkManagerDispatcherFile},
			{check: checkLibvirtCrcNetworkAvailable},
			{check: checkLibvirtCrcNetworkActive},
			{check: checkBundleExtracted},
		},
	},
	{
		distro:          &ubuntu,
		networkMode:     network.SystemNetworkingMode,
		systemdResolved: false,
		checks: []Check{
			{check: checkIfRunningAsNormalUser},
			{check: checkRunningInsideWSL2},
			{check: checkAdminHelperExecutableCached},
			{check: checkOldAdminHelperExecutableCached},
			{check: checkSupportedCPUArch},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{cleanup: removeOldLogs},
			{cleanup: cluster.ForgetPullSecret},
			{cleanup: removeHostsFileEntry},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
			{configKeySuffix: "check-libvirt-group-active"},
			{check: checkLibvirtServiceRunning},
			{check: checkLibvirtVersion},
			{check: checkMachineDriverLibvirtInstalled},
			{cleanup: removeCrcVM},
			{configKeySuffix: "check-apparmor-profile-setup"},
			{check: checkSystemdNetworkdIsNotRunning},
			{check: checkNetworkManagerInstalled},
			{check: checkNetworkManagerIsRunning},
			{check: checkCrcNetworkManagerConfig},
			{check: checkCrcDnsmasqConfigFile},
			{check: checkLibvirtCrcNetworkAvailable},
			{check: checkLibvirtCrcNetworkActive},
			{check: checkBundleExtracted},
		},
	},
	{
		distro:          &ubuntu,
		networkMode:     network.UserNetworkingMode,
		systemdResolved: false,
		checks: []Check{
			{check: checkIfRunningAsNormalUser},
			{check: checkRunningInsideWSL2},
			{check: checkAdminHelperExecutableCached},
			{check: checkOldAdminHelperExecutableCached},
			{check: checkSupportedCPUArch},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{cleanup: removeOldLogs},
			{cleanup: cluster.ForgetPullSecret},
			{cleanup: removeHostsFileEntry},
			{check: checkCrcSymlink},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
			{configKeySuffix: "check-libvirt-group-active"},
			{check: checkLibvirtServiceRunning},
			{check: checkLibvirtVersion},
			{check: checkMachineDriverLibvirtInstalled},
			{cleanup: removeCrcVM},
			{check: checkDaemonSystemdService},
			{check: checkDaemonSystemdSockets},
			{configKeySuffix: "check-apparmor-profile-setup"},
			{check: checkVsock},
			{check: checkBundleExtracted},
		},
	},
}

func funcToString(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func assertFuncEqual(t *testing.T, func1 interface{}, func2 interface{}) {
	assert.Equal(t, reflect.ValueOf(func1).Pointer(), reflect.ValueOf(func2).Pointer(), "%s != %s", funcToString(func1), funcToString(func2))
}

func assertExpectedPreflights(t *testing.T, distro *crcos.OsRelease, networkMode network.Mode, systemdResolved bool) {
	preflights := getPreflightChecksForDistro(distro, networkMode, systemdResolved)
	var expected checkListForDistro
	for _, expected = range checkListForDistros {
		if expected.distro == distro && expected.networkMode == networkMode && expected.systemdResolved == systemdResolved {
			break
		}
	}

	assert.Equal(t, len(preflights), len(expected.checks), "%s %s - %s - expected: %d - got: %d", distro.ID, distro.VersionID, networkMode, len(expected.checks), len(preflights))

	for i := range preflights {
		expectedCheck := expected.checks[i]
		if expectedCheck.configKeySuffix != "" {
			assert.Equal(t, preflights[i].configKeySuffix, expectedCheck.configKeySuffix)
		}
		if expectedCheck.check != nil {
			assertFuncEqual(t, preflights[i].check, expectedCheck.check)
		}
		if expectedCheck.fix != nil {
			assertFuncEqual(t, preflights[i].fix, expectedCheck.fix)
		}
		if expectedCheck.cleanup != nil {
			assertFuncEqual(t, preflights[i].cleanup, expectedCheck.cleanup)
		}
	}
}

func TestCountPreflights(t *testing.T) {
	assertExpectedPreflights(t, &fedora, network.SystemNetworkingMode, true)
	assertExpectedPreflights(t, &fedora, network.SystemNetworkingMode, false)
	assertExpectedPreflights(t, &fedora, network.UserNetworkingMode, false)

	assertExpectedPreflights(t, &rhel, network.SystemNetworkingMode, true)
	assertExpectedPreflights(t, &rhel, network.SystemNetworkingMode, false)
	assertExpectedPreflights(t, &rhel, network.UserNetworkingMode, false)

	assertExpectedPreflights(t, &unexpected, network.SystemNetworkingMode, true)
	assertExpectedPreflights(t, &unexpected, network.SystemNetworkingMode, false)
	assertExpectedPreflights(t, &unexpected, network.UserNetworkingMode, false)

	assertExpectedPreflights(t, &ubuntu, network.SystemNetworkingMode, true)
	assertExpectedPreflights(t, &ubuntu, network.SystemNetworkingMode, false)
	assertExpectedPreflights(t, &ubuntu, network.UserNetworkingMode, false)
}
