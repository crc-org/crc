package constants

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/YourFin/binappend"
	"github.com/code-ready/crc/pkg/crc/version"
)

const (
	DefaultName     = "crc"
	DefaultCPUs     = 4
	DefaultMemory   = 9216
	DefaultDiskSize = 31

	DefaultSSHUser = "core"
	DefaultSSHPort = 22

	CrcEnvPrefix = "CRC"

	ConfigFile                = "crc.json"
	LogFile                   = "crc.log"
	DaemonLogFile             = "crcd.log"
	CrcLandingPageURL         = "https://cloud.redhat.com/openshift/create/local" // #nosec G101
	DefaultPodmanURLBase      = "https://storage.googleapis.com/libpod-master-releases"
	DefaultAdminHelperCliBase = "https://github.com/code-ready/admin-helper/releases/download/v0.0.9"
	CRCMacTrayDownloadURL     = "https://github.com/code-ready/tray-macos/releases/download/v%s/crc-tray-macos.tar.gz"
	CRCWindowsTrayDownloadURL = "https://github.com/code-ready/tray-windows/releases/download/v%s/crc-tray-windows.zip"
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

func defaultBundleForOs(bundleVersion string) map[string]string {
	return map[string]string{
		"darwin":  fmt.Sprintf("crc_hyperkit_%s.crcbundle", bundleVersion),
		"linux":   fmt.Sprintf("crc_libvirt_%s.crcbundle", bundleVersion),
		"windows": fmt.Sprintf("crc_hyperv_%s.crcbundle", bundleVersion),
	}
}

func GetDefaultBundleForOs(os string) string {
	return GetBundleFosOs(os, version.GetBundleVersion())
}

func GetBundleFosOs(os, bundleVersion string) string {
	return defaultBundleForOs(bundleVersion)[os]
}

func GetDefaultBundle() string {
	return GetDefaultBundleForOs(runtime.GOOS)
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
	DefaultBundlePath  = defaultBundlePath()
	DaemonSocketPath   = filepath.Join(CrcBaseDir, "crc.sock")
	KubeconfigFilePath = filepath.Join(MachineInstanceDir, DefaultName, "kubeconfig")
)

func defaultBundlePath() string {
	if version.IsInstaller() {
		return filepath.Join(version.InstallPath(), GetDefaultBundle())
	}
	return filepath.Join(MachineCacheDir, GetDefaultBundle())
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

// BundleEmbedded returns true if the executable was compiled to contain the bundle
func BundleEmbedded() bool {
	executablePath, err := os.Executable()
	if err != nil {
		return false
	}
	extractor, err := binappend.MakeExtractor(executablePath)
	if err != nil {
		return false
	}
	return contains(extractor.AvalibleData(), GetDefaultBundle())
}

func IsRelease() bool {
	return BundleEmbedded() || version.IsInstaller()
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
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
	return fmt.Sprintf(CRCMacTrayDownloadURL, version.GetCRCMacTrayVersion())
}

func GetCRCWindowsTrayDownloadURL() string {
	return fmt.Sprintf(CRCWindowsTrayDownloadURL, version.GetCRCWindowsTrayVersion())
}
