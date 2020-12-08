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

	DefaultWebConsoleURL      = "https://console-openshift-console.apps-crc.testing"
	DefaultAPIURL             = "https://api.crc.testing:6443"
	DefaultLogLevel           = "info"
	ConfigFile                = "crc.json"
	LogFile                   = "crc.log"
	DaemonLogFile             = "crcd.log"
	CrcLandingPageURL         = "https://cloud.redhat.com/openshift/install/crc/installer-provisioned" // #nosec G101
	DefaultPodmanURLBase      = "https://storage.googleapis.com/libpod-master-releases"
	DefaultGoodhostsCliBase   = "https://github.com/code-ready/goodhosts-cli/releases/download/v1.0.0"
	CRCMacTrayDownloadURL     = "https://github.com/code-ready/tray-macos/releases/download/v%s/crc-tray-macos.tar.gz"
	CRCWindowsTrayDownloadURL = "https://github.com/code-ready/tray-windows/releases/download/v%s/crc-tray-windows.zip"
	DefaultContext            = "admin"

	VSockGateway = "192.168.127.1"
	VsockSSHPort = 2222

	OkdPullSecret = `{"auths":{"fake":{"auth": "Zm9vOmJhcgo="}}}` // #nosec G101
)

var podmanURLForOs = map[string]string{
	"darwin":  fmt.Sprintf("%s/%s", DefaultPodmanURLBase, "podman-remote-latest-master-darwin-amd64.zip"),
	"linux":   fmt.Sprintf("%s/%s", DefaultPodmanURLBase, "podman-remote-latest-master-linux---amd64.zip"),
	"windows": fmt.Sprintf("%s/%s", DefaultPodmanURLBase, "podman-remote-latest-master-windows-amd64.zip"),
}

func GetPodmanURLForOs(os string) string {
	return podmanURLForOs[os]
}

func GetPodmanURL() string {
	return podmanURLForOs[runtime.GOOS]
}

var goodhostsURLForOs = map[string]string{
	"darwin":  fmt.Sprintf("%s/%s", DefaultGoodhostsCliBase, "goodhosts-cli-macos-amd64.tar.xz"),
	"linux":   fmt.Sprintf("%s/%s", DefaultGoodhostsCliBase, "goodhosts-cli-linux-amd64.tar.xz"),
	"windows": fmt.Sprintf("%s/%s", DefaultGoodhostsCliBase, "goodhosts-cli-windows-amd64.tar.xz"),
}

func GetGoodhostsURLForOs(os string) string {
	return goodhostsURLForOs[os]
}

func GetGoodhostsURL() string {
	return goodhostsURLForOs[runtime.GOOS]
}

var defaultBundleForOs = map[string]string{
	"darwin":  fmt.Sprintf("crc_hyperkit_%s.crcbundle", version.GetBundleVersion()),
	"linux":   fmt.Sprintf("crc_libvirt_%s.crcbundle", version.GetBundleVersion()),
	"windows": fmt.Sprintf("crc_hyperv_%s.crcbundle", version.GetBundleVersion()),
}

func GetDefaultBundleForOs(os string) string {
	return defaultBundleForOs[os]
}

func GetDefaultBundle() string {
	return GetDefaultBundleForOs(runtime.GOOS)
}

var (
	CrcBaseDir         = filepath.Join(GetHomeDir(), ".crc")
	CrcBinDir          = filepath.Join(CrcBaseDir, "bin")
	CrcOcBinDir        = filepath.Join(CrcBinDir, "oc")
	ConfigPath         = filepath.Join(CrcBaseDir, ConfigFile)
	LogFilePath        = filepath.Join(CrcBaseDir, LogFile)
	DaemonLogFilePath  = filepath.Join(CrcBaseDir, DaemonLogFile)
	MachineBaseDir     = CrcBaseDir
	MachineCertsDir    = filepath.Join(MachineBaseDir, "certs")
	MachineCacheDir    = filepath.Join(MachineBaseDir, "cache")
	MachineInstanceDir = filepath.Join(MachineBaseDir, "machines")
	DefaultBundlePath  = filepath.Join(MachineCacheDir, GetDefaultBundle())
	DaemonSocketPath   = filepath.Join(CrcBaseDir, "crc.sock")
	NetworkSocketPath  = filepath.Join(CrcBaseDir, "network.sock")
)

// GetHomeDir returns the home directory for the current user
func GetHomeDir() string {
	if runtime.GOOS == "windows" {
		if homeDrive, homePath := os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"); len(homeDrive) > 0 && len(homePath) > 0 {
			homeDir := filepath.Join(homeDrive, homePath)
			if _, err := os.Stat(homeDir); err == nil {
				return homeDir
			}
		}
		if userProfile := os.Getenv("USERPROFILE"); len(userProfile) > 0 {
			if _, err := os.Stat(userProfile); err == nil {
				return userProfile
			}
		}
	}
	return os.Getenv("HOME")
}

// EnsureBaseDirectoryExists create the ~/.crc directory if it is not present
func EnsureBaseDirectoriesExist() error {
	return os.MkdirAll(CrcBaseDir, 0750)
}

// IsBundleEmbedded returns true if the executable was compiled to contain the bundle
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

// TODO: follow the same pattern as oc and podman above
func GetCRCMacTrayDownloadURL() string {
	return fmt.Sprintf(CRCMacTrayDownloadURL, version.GetCRCMacTrayVersion())
}

func GetCRCWindowsTrayDownloadURL() string {
	return fmt.Sprintf(CRCWindowsTrayDownloadURL, version.GetCRCWindowsTrayVersion())
}
