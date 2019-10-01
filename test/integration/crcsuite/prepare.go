// +build integration

package crcsuite

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/code-ready/crc/pkg/download"
)

// Download bundle for testing
func DownloadBundle(bundleLocation string, bundleDestination string) (string, error) {

	if bundleLocation[:4] != "http" {

		// copy the file locall

		if bundleDestination == "." {
			bundleDestination, _ = os.Getwd()
		}
		fmt.Printf("Copying bundle from %s to %s.\n", bundleLocation, bundleDestination)
		bundleDestination = filepath.Join(bundleDestination, bundleName)

		source, err := os.Open(bundleLocation)
		if err != nil {
			return "", err
		}
		defer source.Close()

		destination, err := os.Create(bundleDestination)
		if err != nil {
			return "", err
		}
		defer destination.Close()

		_, err = io.Copy(destination, source)
		if err != nil {
			return "", err
		}

		err = destination.Sync()

		return bundleDestination, err
	}

	filename, err := download.Download(bundleLocation, bundleDestination, 0644)
	fmt.Printf("Downloading bundle from %s to %s.\n", bundleLocation, bundleDestination)
	if err != nil {
		return "", err
	}

	return filename, nil
}

// Parse GODOG flags (in feature files)
func ParseFlags() {

	flag.Parse()
	if flag.NArg() < 2 {
		fmt.Printf("Invalid number of arguments, the path to the pull secret file and the bundle version are required\n")
		os.Exit(1)
	}

	pullSecretFile = flag.Args()[0]

	// embedded bundle
	// this will never occur on Centos CI
	if flag.NArg() == 2 {
		bundleEmbedded = true
		bundleVersion = flag.Args()[1]
		// assume default hypervisor
		var hypervisor string
		switch os := runtime.GOOS; os {
		case "darwin":
			hypervisor = "hyperkit"
		case "linux":
			hypervisor = "libvirt"
		case "windows":
			hypervisor = "hyperv"
		default:
			fmt.Printf("Unsupported OS: %s", os)
		}

		bundleName = fmt.Sprintf("crc_%s_%s.crcbundle", hypervisor, bundleVersion)

	}

	// separate bundle
	// 3: on laptop
	// 4: on Centos CI
	if flag.NArg() > 2 {
		bundleEmbedded = false
		bundleURL = flag.Args()[1]
		_, bundleName = filepath.Split(bundleURL)
	}

}

// Set CRCHome var to ~/.crc
func SetCRCHome() string {
	usr, _ := user.Current()
	crcHome := filepath.Join(usr.HomeDir, ".crc")
	return crcHome
}
