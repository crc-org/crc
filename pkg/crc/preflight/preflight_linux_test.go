package preflight

import (
	"reflect"
	"runtime"
	"testing"

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
			preflightChecksCount += 2
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
		VersionID: "8.2",
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
	distro      *crcos.OsRelease
	networkMode network.Mode
	checks      []Check
}

var checkListForDistros = []checkListForDistro{
	{
		distro:      &fedora,
		networkMode: network.DefaultMode,
		checks: []Check{
			{check: checkPodmanExecutableCached},
			{check: checkGoodhostsExecutableCached},
			{check: checkBundleExtracted},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{check: checkIfRunningAsNormalUser},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
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
		},
	},
	{
		distro:      &fedora,
		networkMode: network.VSockMode,
		checks: []Check{
			{check: checkPodmanExecutableCached},
			{check: checkGoodhostsExecutableCached},
			{check: checkBundleExtracted},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{check: checkIfRunningAsNormalUser},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
			{check: checkLibvirtServiceRunning},
			{check: checkLibvirtVersion},
			{check: checkMachineDriverLibvirtInstalled},
			{cleanup: removeCrcVM},
			{check: checkVsock},
		},
	},
	{
		distro:      &rhel,
		networkMode: network.DefaultMode,
		checks: []Check{
			{check: checkPodmanExecutableCached},
			{check: checkGoodhostsExecutableCached},
			{check: checkBundleExtracted},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{check: checkIfRunningAsNormalUser},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
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
		},
	},
	{
		distro:      &rhel,
		networkMode: network.VSockMode,
		checks: []Check{
			{check: checkPodmanExecutableCached},
			{check: checkGoodhostsExecutableCached},
			{check: checkBundleExtracted},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{check: checkIfRunningAsNormalUser},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
			{check: checkLibvirtServiceRunning},
			{check: checkLibvirtVersion},
			{check: checkMachineDriverLibvirtInstalled},
			{cleanup: removeCrcVM},
			{check: checkVsock},
		},
	},
	{
		distro:      &unexpected,
		networkMode: network.DefaultMode,
		checks: []Check{
			{check: checkPodmanExecutableCached},
			{check: checkGoodhostsExecutableCached},
			{check: checkBundleExtracted},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{check: checkIfRunningAsNormalUser},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
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
		},
	},
	{
		distro:      &unexpected,
		networkMode: network.VSockMode,
		checks: []Check{
			{check: checkPodmanExecutableCached},
			{check: checkGoodhostsExecutableCached},
			{check: checkBundleExtracted},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{check: checkIfRunningAsNormalUser},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
			{check: checkLibvirtServiceRunning},
			{check: checkLibvirtVersion},
			{check: checkMachineDriverLibvirtInstalled},
			{cleanup: removeCrcVM},
			{check: checkVsock},
		},
	},
	{
		distro:      &ubuntu,
		networkMode: network.DefaultMode,
		checks: []Check{
			{check: checkPodmanExecutableCached},
			{check: checkGoodhostsExecutableCached},
			{check: checkBundleExtracted},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{check: checkIfRunningAsNormalUser},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
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
		},
	},
	{
		distro:      &ubuntu,
		networkMode: network.VSockMode,
		checks: []Check{
			{check: checkPodmanExecutableCached},
			{check: checkGoodhostsExecutableCached},
			{check: checkBundleExtracted},
			{configKeySuffix: "check-ram"},
			{cleanup: removeCRCMachinesDir},
			{check: checkIfRunningAsNormalUser},
			{check: checkVirtualizationEnabled},
			{check: checkKvmEnabled},
			{check: checkLibvirtInstalled},
			{check: checkUserPartOfLibvirtGroup},
			{check: checkLibvirtServiceRunning},
			{check: checkLibvirtVersion},
			{check: checkMachineDriverLibvirtInstalled},
			{cleanup: removeCrcVM},
			{configKeySuffix: "check-apparmor-profile-setup"},
			{check: checkVsock},
		},
	},
}

func funcToString(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func assertFuncEqual(t *testing.T, func1 interface{}, func2 interface{}) {
	assert.Equal(t, reflect.ValueOf(func1).Pointer(), reflect.ValueOf(func2).Pointer(), "%s != %s", funcToString(func1), funcToString(func2))
}

func assertExpectedPreflights(t *testing.T, distro *crcos.OsRelease, networkMode network.Mode) {
	preflights := getPreflightChecksForDistro(distro, networkMode)
	var expected checkListForDistro
	for _, expected = range checkListForDistros {
		if expected.distro == distro && expected.networkMode == networkMode {
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
	assertExpectedPreflights(t, &fedora, network.DefaultMode)
	assertExpectedPreflights(t, &fedora, network.VSockMode)

	assertExpectedPreflights(t, &rhel, network.DefaultMode)
	assertExpectedPreflights(t, &rhel, network.VSockMode)

	assertExpectedPreflights(t, &unexpected, network.DefaultMode)
	assertExpectedPreflights(t, &unexpected, network.VSockMode)

	assertExpectedPreflights(t, &ubuntu, network.DefaultMode)
	assertExpectedPreflights(t, &ubuntu, network.VSockMode)
}
