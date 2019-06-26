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

// Get the bundle for testing
func GetBundle(bundleLocation string, bundleDestination string) (string, error) {

	// Copy the bundle locally
	if bundleLocation[:4] != "http" {

		if bundleDestination == "." {
			bundleDestination, _ = os.Getwd()
		}
		fmt.Printf("Copying bundle from %s to %s.\n", bundleLocation, bundleDestination)
		bundleDestination = filepath.Join(bundleDestination, BundleName)

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

	// Download the bundle over http
	filename, err := download.Download(bundleLocation, bundleDestination)
	fmt.Printf("Downloading bundle from %s to %s.\n", bundleLocation, bundleDestination)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func ParseFlags() {

	flag.StringVar(&BundleLocation, "bundle", "", "Bundle URL or filepath")

}

// Set CRCHome var to ~/.crc
func SetCRCHome() string {
	usr, _ := user.Current()
	crcHome := filepath.Join(usr.HomeDir, ".crc")
	return crcHome
}

func SetBundleName() string {

	_, bname := filepath.Split(BundleLocation)
	return bname
}
