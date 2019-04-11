package constants

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	DomName      = "crc"
	NodeMac      = "52:fd:fc:07:21:82"
	NodeIP       = "192.168.126.11"
	PoolName     = "crc"
	PoolDir      = "/var/lib/libvirt/images"
	CrcEnvPrefix = "CRC"
)

var (
	ConfigFile  = ".crc.json"
	ConfigPath  = filepath.Join(GetHomeDir(), ConfigFile)
	LogFile     = ".crc.log"
	LogFilePath = filepath.Join(GetHomeDir(), LogFile)
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
