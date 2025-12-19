//go:build darwin

package constants

import (
	"log"
	"os"
	"path/filepath"
)

const (
	OcExecutableName           = "oc"
	PodmanRemoteExecutableName = "podman"
	DaemonAgentLabel           = "com.redhat.crc.daemon"
	QemuGuestAgentPort         = 1234

	VfkitCommand   = "vfkit"
	KrunkitCommand = "krunkit"
)

func Provider() string {
	provider := os.Getenv("CRC_PROVIDER")
	if provider == "" {
		provider = VfkitCommand
	}
	switch provider {
	case VfkitCommand:
		return VfkitCommand
	case KrunkitCommand:
		return KrunkitCommand
	default:
		log.Fatalf("Invalid provider: %s. Choose between %s or %s", provider, VfkitCommand, KrunkitCommand)
		return ""
	}
}

var (
	TapSocketPath        = filepath.Join(CrcBaseDir, "tap.sock")
	DaemonHTTPSocketPath = filepath.Join(CrcBaseDir, "crc-http.sock")
	UnixgramSocketPath   = filepath.Join(CrcBaseDir, "crc-unixgram.sock")
)
