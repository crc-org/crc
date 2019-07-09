// +build integration

package crcsuite

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"

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

	filename, err := download.Download(bundleLocation, bundleDestination)
	fmt.Printf("Downloading bundle from %s to %s.\n", bundleLocation, bundleDestination)
	if err != nil {
		return "", err
	}

	return filename, nil
}

// Parse GODOG flags (in feature files)
func ParseFlags() {

	flag.Parse()
	bundleURL = flag.Args()[0]
	_, bundleName = filepath.Split(bundleURL)
	pullSecretFile = flag.Args()[1]
}

// Set CRCHome var to ~/.crc
func SetCRCHome() string {
	usr, _ := user.Current()
	crcHome := filepath.Join(usr.HomeDir, ".crc")
	return crcHome
}
