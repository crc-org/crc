package constants

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	crcpreset "github.com/crc-org/crc/pkg/crc/preset"
	"github.com/crc-org/crc/pkg/crc/version"
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
	CrcLandingPageURL         = "https://console.redhat.com/openshift/create/local" // #nosec G101
	DefaultAdminHelperURLBase = "https://github.com/crc-org/admin-helper/releases/download/v%s/%s"
	CRCMacTrayDownloadURL     = "https://github.com/crc-org/tray-electron/releases/download/%s/crc-tray-macos.tar.gz"
	CRCWindowsTrayDownloadURL = "https://github.com/crc-org/tray-electron/releases/download/%s/crc-tray-windows.zip"
	DefaultContext            = "admin"
	DaemonHTTPEndpoint        = "http://unix/api"
	DaemonVsockPort           = 1024
	DefaultPodmanNamedPipe    = `\\.\pipe\crc-podman`
	RootlessPodmanSocket      = "/run/user/1000/podman/podman.sock"
	RootfulPodmanSocket       = "/run/podman/podman.sock"

	VSockGateway = "192.168.127.1"
	VsockSSHPort = 2222

	OkdPullSecret = `{"auths":{"fake":{"auth": "Zm9vOmJhcgo="}}}` // #nosec G101

	RegistryURI   = "quay.io/crcont"
	ClusterDomain = ".crc.testing"
	AppsDomain    = ".apps-crc.testing"

	OpenShiftIngressHTTPPort  = 80
	OpenShiftIngressHTTPSPort = 443

	// This public key is owned by the CRC team (crc@crc.dev), and is used
	// to sign bundles uploaded to an image registry.
	// It can be fetched with: `gpg --recv-key DC7EAC400A1BFDFB`
	GPGPublicKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mDMEYrvgDRYJKwYBBAHaRw8BAQdAoW+hjSRYpTAdLEE1u6ZuYNER1g97e8ygT4ic
mvo1AKi0MmNyYyAoS2V5IHRvIHNpZ24gYnVuZGxlIHVzZWQgYnkgY3JjKSA8Y3Jj
QGNyYy5kZXY+iJkEExYKAEEWIQS4RlW/rByOBn/ZyofcfqxAChv9+wUCYrvgDQIb
AwUJEswDAAULCQgHAgIiAgYVCgkICwIEFgIDAQIeBwIXgAAKCRDcfqxAChv9+/ep
APwISi03R7npwimqdL7NYKDGMO8ikOwmmPkqh9CKwt4CdwD8Cc6HNcZumHDpJ4gH
x7FXxIS9KLwDihpm1Gxr4t1t5Qy4OARiu+ANEgorBgEEAZdVAQUBAQdA/w7pM7hf
bxZ2qwSuoBuhcA1sAlPSb3NrIZf3CceoqzQDAQgHiH4EGBYKACYWIQS4RlW/rByO
Bn/ZyofcfqxAChv9+wUCYrvgDQIbDAUJEswDAAAKCRDcfqxAChv9+2UkAQCNCdaf
vnhbvfPHDltmwDZ3aD4l3jjSKpeySeKQocgjQAD6A7kawst/50k4wb+vUDUnEoYo
9Ix7lKfKWCXil/z0vg4=
=lmb/
-----END PGP PUBLIC KEY BLOCK-----`
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
	CrcSymlinkPath     = filepath.Join(CrcBinDir, "crc")
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

func ResolveHelperPath(executableName string) string {
	if version.IsInstaller() {
		return filepath.Join(version.InstallPath(), executableName)
	}
	return filepath.Join(CrcBinDir, executableName)
}

func AdminHelperPath() string {
	return ResolveHelperPath(GetAdminHelperExecutableForOs(runtime.GOOS))
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

// TODO: follow the same pattern as oc and podman above
func GetCRCMacTrayDownloadURL() string {
	return fmt.Sprintf(CRCMacTrayDownloadURL, version.GetTrayVersion())
}

func GetCRCWindowsTrayDownloadURL() string {
	return fmt.Sprintf(CRCWindowsTrayDownloadURL, version.GetTrayVersion())
}

func GetDefaultCPUs(preset crcpreset.Preset) int {
	switch preset {
	case crcpreset.OpenShift, crcpreset.OKD:
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
	case crcpreset.OpenShift, crcpreset.OKD:
		return 9216
	case crcpreset.Podman:
		return 2048
	default:
		// should not be reached
		return 9216
	}
}
