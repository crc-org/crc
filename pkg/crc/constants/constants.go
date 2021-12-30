package constants

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	crcpreset "github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/crc/version"
)

const (
	DefaultName     = "crc"
	DefaultDiskSize = 31

	DefaultSSHUser = "core"
	DefaultSSHPort = 22

	CrcEnvPrefix = "CRC"

	ConfigFile                = "crc.json"
	LogFile                   = "crc.log"
	DaemonLogFile             = "crcd.log"
	CrcLandingPageURL         = "https://cloud.redhat.com/openshift/create/local" // #nosec G101
	DefaultPodmanURLBase      = "https://storage.googleapis.com/libpod-master-releases"
	DefaultAdminHelperCliBase = "https://github.com/code-ready/admin-helper/releases/download/v0.0.8"
	CRCMacTrayDownloadURL     = "https://github.com/code-ready/tray-electron/releases/download/%s/crc-tray-macos.tar.gz"
	CRCWindowsTrayDownloadURL = "https://github.com/code-ready/tray-electron/releases/download/%s/crc-tray-windows.zip"
	DefaultContext            = "admin"
	DaemonHTTPEndpoint        = "http://unix/api"

	VSockGateway = "192.168.127.1"
	VsockSSHPort = 2222

	OkdPullSecret = `{"auths":{"fake":{"auth": "Zm9vOmJhcgo="}}}` // #nosec G101

	ClusterDomain = ".crc.testing"
	AppsDomain    = ".apps-crc.testing"
)

var adminHelperExecutableForOs = map[string]string{
	"darwin":  "crc-admin-helper-darwin",
	"linux":   "crc-admin-helper-linux",
	"windows": "crc-admin-helper-windows.exe",
}

func GetAdminHelperExecutableForOs(os string) string {
	return adminHelperExecutableForOs[os]
}

func GetAdminHelperExecutable() string {
	return GetAdminHelperExecutableForOs(runtime.GOOS)
}

func GetAdminHelperURLForOs(os string) string {
	return fmt.Sprintf("%s/%s", DefaultAdminHelperCliBase, GetAdminHelperExecutableForOs(os))
}

func GetAdminHelperURL() string {
	return GetAdminHelperURLForOs(runtime.GOOS)
}

func defaultBundleForOs(preset crcpreset.Preset) map[string]string {
	if preset == crcpreset.Podman {
		return map[string]string{
			"darwin":  fmt.Sprintf("crc_podman_hyperkit_%s.crcbundle", version.GetPodmanVersion()),
			"linux":   fmt.Sprintf("crc_podman_libvirt_%s.crcbundle", version.GetPodmanVersion()),
			"windows": fmt.Sprintf("crc_podman_hyperv_%s.crcbundle", version.GetPodmanVersion()),
		}
	}
	return map[string]string{
		"darwin":  fmt.Sprintf("crc_hyperkit_%s.crcbundle", version.GetBundleVersion()),
		"linux":   fmt.Sprintf("crc_libvirt_%s.crcbundle", version.GetBundleVersion()),
		"windows": fmt.Sprintf("crc_hyperv_%s.crcbundle", version.GetBundleVersion()),
	}
}

func GetDefaultBundle(preset crcpreset.Preset) string {
	bundles := defaultBundleForOs(preset)
	return bundles[runtime.GOOS]
}

var (
	CrcBaseDir         = filepath.Join(GetHomeDir(), ".crc")
	crcBinDir          = filepath.Join(CrcBaseDir, "bin")
	CrcOcBinDir        = filepath.Join(crcBinDir, "oc")
	CrcSymlinkPath     = filepath.Join(crcBinDir, "crc")
	ConfigPath         = filepath.Join(CrcBaseDir, ConfigFile)
	LogFilePath        = filepath.Join(CrcBaseDir, LogFile)
	DaemonLogFilePath  = filepath.Join(CrcBaseDir, DaemonLogFile)
	MachineBaseDir     = CrcBaseDir
	MachineCacheDir    = filepath.Join(MachineBaseDir, "cache")
	MachineInstanceDir = filepath.Join(MachineBaseDir, "machines")
	DaemonSocketPath   = filepath.Join(CrcBaseDir, "crc.sock")
	KubeconfigFilePath = filepath.Join(MachineInstanceDir, DefaultName, "kubeconfig")
)

func GetDefaultBundlePath(preset crcpreset.Preset) string {
	return filepath.Join(MachineCacheDir, GetDefaultBundle(preset))
}

func BinDir() string {
	if version.IsInstaller() {
		return version.InstallPath()
	}
	return crcBinDir
}

// GetHomeDir returns the home directory for the current user
func GetHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("Failed to get homeDir: " + err.Error())
	}
	return homeDir
}

// EnsureBaseDirectoryExists create the ~/.crc directory if it is not present
func EnsureBaseDirectoriesExist() error {
	return os.MkdirAll(CrcBaseDir, 0750)
}

func IsRelease() bool {
	return version.IsInstaller() || version.IsLinuxRelease()
}

func GetPublicKeyPath() string {
	return filepath.Join(MachineInstanceDir, DefaultName, "id_ecdsa.pub")
}

func GetPrivateKeyPath() string {
	return filepath.Join(MachineInstanceDir, DefaultName, "id_ecdsa")
}

// For backward compatibility to v 1.20.0
func GetRsaPrivateKeyPath() string {
	return filepath.Join(MachineInstanceDir, DefaultName, "id_rsa")
}

func GetKubeAdminPasswordPath() string {
	return filepath.Join(MachineInstanceDir, DefaultName, "kubeadmin-password")
}

// TODO: follow the same pattern as oc and podman above
func GetCRCMacTrayDownloadURL() string {
	return fmt.Sprintf(CRCMacTrayDownloadURL, version.GetTrayVersion())
}

func GetCRCWindowsTrayDownloadURL() string {
	return fmt.Sprintf(CRCWindowsTrayDownloadURL, version.GetTrayVersion())
}

func GetDefaultCPUs(preset crcpreset.Preset) int {
	switch preset {
	case crcpreset.OpenShift:
		return 4
	case crcpreset.Podman:
		return 2
	default:
		// should not be reached
		return 4
	}
}

func GetDefaultMemory(preset crcpreset.Preset) int {
	switch preset {
	case crcpreset.OpenShift:
		return 9216
	case crcpreset.Podman:
		return 2048
	default:
		// should not be reached
		return 9216
	}
}
