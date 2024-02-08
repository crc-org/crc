package constants

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	crcpreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/crc/version"
)

const (
	DefaultName     = "crc"
	DefaultDiskSize = 31

	DefaultPersistentVolumeSize = 15

	DefaultSSHUser = "core"
	DefaultSSHPort = 22

	CrcEnvPrefix = "CRC"

	ConfigFile                = "crc.json"
	LogFile                   = "crc.log"
	DaemonLogFile             = "crcd.log"
	CrcLandingPageURL         = "https://console.redhat.com/openshift/create/local" // #nosec G101
	DefaultAdminHelperURLBase = "https://github.com/crc-org/admin-helper/releases/download/v%s/%s"
	BackgroundLauncherURL     = "https://github.com/crc-org/win32-background-launcher/releases/download/v%s/win32-background-launcher.exe"
	DefaultBundleURLBase      = "https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/%s/%s/%s"
	DefaultContext            = "admin"
	DaemonHTTPEndpoint        = "http://unix/api"
	DaemonVsockPort           = 1024
	DefaultPodmanNamedPipe    = `\\.\pipe\crc-podman`
	RootlessPodmanSocket      = "/run/user/1000/podman/podman.sock"
	RootfulPodmanSocket       = "/run/podman/podman.sock"

	VSockGateway = "192.168.127.1"
	VsockSSHPort = 2222
	LocalIP      = "127.0.0.1"

	OkdPullSecret = `{"auths":{"fake":{"auth": "Zm9vOmJhcgo="}}}` // #nosec G101

	RegistryURI         = "quay.io/crcont"
	ClusterDomain       = ".crc.testing"
	AppsDomain          = ".apps-crc.testing"
	MicroShiftAppDomain = ".apps.crc.testing"

	OpenShiftIngressHTTPPort  = 80
	OpenShiftIngressHTTPSPort = 443

	BackgroundLauncherExecutable = "crc-background-launcher.exe"
)

var adminHelperExecutableForOs = map[string]string{
	"darwin":  "crc-admin-helper-darwin",
	"linux":   "crc-admin-helper-linux",
	"windows": "crc-admin-helper-windows.exe",
}

func GetAdminHelperExecutableForOs(os string) string {
	return adminHelperExecutableForOs[os]
}

func GetAdminHelperURLForOs(os string) string {
	return fmt.Sprintf(DefaultAdminHelperURLBase, version.GetAdminHelperVersion(), GetAdminHelperExecutableForOs(os))
}

func GetAdminHelperURL() string {
	return GetAdminHelperURLForOs(runtime.GOOS)
}

func BundleForPreset(preset crcpreset.Preset, version string) string {
	var bundleName strings.Builder

	bundleName.WriteString("crc")

	switch preset {
	case crcpreset.Podman:
		bundleName.WriteString("_podman")
	case crcpreset.OKD:
		bundleName.WriteString("_okd")
	case crcpreset.Microshift:
		bundleName.WriteString("_microshift")
	}

	switch runtime.GOOS {
	case "darwin":
		bundleName.WriteString("_vfkit")
	case "linux":
		bundleName.WriteString("_libvirt")
	case "windows":
		bundleName.WriteString("_hyperv")
	}

	fmt.Fprintf(&bundleName, "_%s_%s.crcbundle", version, runtime.GOARCH)
	return bundleName.String()
}

func GetDefaultBundle(preset crcpreset.Preset) string {
	return BundleForPreset(preset, version.GetBundleVersion(preset))
}

var (
	CrcBaseDir         = filepath.Join(GetHomeDir(), ".crc")
	CrcBinDir          = filepath.Join(CrcBaseDir, "bin")
	CrcOcBinDir        = filepath.Join(CrcBinDir, "oc")
	CrcPodmanBinDir    = filepath.Join(CrcBinDir, "podman")
	CrcSymlinkPath     = filepath.Join(CrcBinDir, "crc")
	ConfigPath         = filepath.Join(CrcBaseDir, ConfigFile)
	LogFilePath        = filepath.Join(CrcBaseDir, LogFile)
	DaemonLogFilePath  = filepath.Join(CrcBaseDir, DaemonLogFile)
	MachineBaseDir     = CrcBaseDir
	MachineCacheDir    = filepath.Join(MachineBaseDir, "cache")
	MachineInstanceDir = filepath.Join(MachineBaseDir, "machines")
	DaemonSocketPath   = filepath.Join(CrcBaseDir, "crc.sock")
	KubeconfigFilePath = filepath.Join(MachineInstanceDir, DefaultName, "kubeconfig")
	PasswdFilePath     = filepath.Join(MachineInstanceDir, DefaultName, "passwd")
)

func GetDefaultBundlePath(preset crcpreset.Preset) string {
	return filepath.Join(MachineCacheDir, GetDefaultBundle(preset))
}

func GetDefaultBundleDownloadURL(preset crcpreset.Preset) string {
	return fmt.Sprintf(DefaultBundleURLBase,
		preset.String(),
		version.GetBundleVersion(preset),
		GetDefaultBundle(preset),
	)
}

func GetDefaultBundleSignedHashURL(preset crcpreset.Preset) string {
	return fmt.Sprintf(DefaultBundleURLBase,
		preset.String(),
		version.GetBundleVersion(preset),
		"sha256sum.txt.sig",
	)
}

func ResolveHelperPath(executableName string) string {
	if version.IsInstaller() {
		return filepath.Join(version.InstallPath(), executableName)
	}
	return filepath.Join(CrcBinDir, executableName)
}

func AdminHelperPath() string {
	return ResolveHelperPath(GetAdminHelperExecutableForOs(runtime.GOOS))
}

func Win32BackgroundLauncherPath() string {
	return ResolveHelperPath(BackgroundLauncherExecutable)
}

// GetHomeDir returns the home directory for the current user
func GetHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("Failed to get homeDir: " + err.Error())
	}
	return homeDir
}

// EnsureBaseDirectoriesExist creates ~/.crc, ~/.crc/bin and ~/.crc/cache directories if it is not present
func EnsureBaseDirectoriesExist() error {
	baseDirectories := []string{CrcBaseDir, MachineCacheDir, CrcBinDir}
	for _, baseDir := range baseDirectories {
		err := os.MkdirAll(baseDir, 0750)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetPublicKeyPath() string {
	return filepath.Join(MachineInstanceDir, DefaultName, "id_ecdsa.pub")
}

func GetPrivateKeyPath() string {
	return filepath.Join(MachineInstanceDir, DefaultName, "id_ecdsa")
}

func GetHostDockerSocketPath() string {
	return filepath.Join(MachineInstanceDir, DefaultName, "docker.sock")
}

// For backward compatibility to v 1.20.0
func GetRsaPrivateKeyPath() string {
	return filepath.Join(MachineInstanceDir, DefaultName, "id_rsa")
}

func GetKubeAdminPasswordPath() string {
	return filepath.Join(MachineInstanceDir, DefaultName, "kubeadmin-password")
}

func GetWin32BackgroundLauncherDownloadURL() string {
	return fmt.Sprintf(BackgroundLauncherURL,
		version.GetWin32BackgroundLauncherVersion())
}

func GetDefaultCPUs(preset crcpreset.Preset) int {
	switch preset {
	case crcpreset.OpenShift, crcpreset.OKD:
		return 4
	case crcpreset.Podman, crcpreset.Microshift:
		return 2
	default:
		// should not be reached
		return 4
	}
}

func GetDefaultMemory(preset crcpreset.Preset) int {
	switch preset {
	case crcpreset.OpenShift, crcpreset.OKD:
		return 9216
	case crcpreset.Podman:
		return 2048
	case crcpreset.Microshift:
		return 4096
	default:
		// should not be reached
		return 9216
	}
}

func GetDefaultBundleImageRegistry(preset crcpreset.Preset) string {
	return fmt.Sprintf("//%s/%s:%s", RegistryURI, getImageName(preset), version.GetBundleVersion(preset))
}

func getImageName(preset crcpreset.Preset) string {
	switch preset {
	case crcpreset.Podman:
		return "podman-bundle"
	case crcpreset.OKD:
		return "okd-bundle"
	case crcpreset.Microshift:
		return "microshift-bundle"
	default:
		return "openshift-bundle"
	}
}
