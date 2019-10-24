package constants

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/code-ready/crc/pkg/crc/version"
)

const (
	DefaultName   = "crc"
	DefaultCPUs   = 4
	DefaultMemory = 8192

	DefaultSSHPort = 22
	DefaultSSHUser = "core"

	CrcEnvPrefix = "CRC"

	DefaultWebConsoleURL = "https://console-openshift-console.apps-crc.testing"
	DefaultAPIURL        = "https://api.crc.testing:6443"
	DefaultDiskImage     = "crc.disk"
	DefaultLogLevel      = "info"
	ConfigFile           = "crc.json"
	LogFile              = "crc.log"
	GlobalStateFile      = "globalstate.json"
	CrcLandingPageURL    = "https://cloud.redhat.com/openshift/install/crc/installer-provisioned" // #nosec G101
	PullSecretFile       = "pullsecret.json"

	DefaultOcUrlBase = "https://mirror.openshift.com/pub/openshift-v4/clients/oc/latest"
)

var ocUrlForOs = map[string]string{
	"darwin":  fmt.Sprintf("%s/%s", DefaultOcUrlBase, "macosx/oc.tar.gz"),
	"linux":   fmt.Sprintf("%s/%s", DefaultOcUrlBase, "linux/oc.tar.gz"),
	"windows": fmt.Sprintf("%s/%s", DefaultOcUrlBase, "windows/oc.zip"),
}

func GetOcUrlForOs(os string) string {
	return ocUrlForOs[os]
}

func GetOcUrl() string {
	return GetOcUrlForOs(runtime.GOOS)
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
	ConfigPath         = filepath.Join(CrcBaseDir, ConfigFile)
	LogFilePath        = filepath.Join(CrcBaseDir, LogFile)
	MachineBaseDir     = CrcBaseDir
	MachineCertsDir    = filepath.Join(MachineBaseDir, "certs")
	MachineCacheDir    = filepath.Join(MachineBaseDir, "cache")
	MachineInstanceDir = filepath.Join(MachineBaseDir, "machines")
	GlobalStatePath    = filepath.Join(CrcBaseDir, GlobalStateFile)
	DefaultBundlePath  = filepath.Join(MachineCacheDir, GetDefaultBundle())
	bundleEmbedded     = "false"
	DaemonSocketPath   = filepath.Join(CrcBaseDir, "crc.sock")
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

// EnsureBaseDirExists create the ~/.crc dir if its not there
func EnsureBaseDirExists() error {
	_, err := os.Stat(CrcBaseDir)
	if err != nil {
		return os.Mkdir(CrcBaseDir, 0750)
	}
	return nil
}

// IsBundleEmbedded returns true if the binary was compiled to contain the bundle
func BundleEmbedded() bool {
	return bundleEmbedded == "true"
}

func GetPublicKeyPath() string {
	return filepath.Join(MachineInstanceDir, DefaultName, "id_rsa.pub")
}

func GetPrivateKeyPath() string {
	return filepath.Join(MachineInstanceDir, DefaultName, "id_rsa")
}
