package crcsuite

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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

func CopyFilesToTestDir() error {

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error retrieving current dir: %s", err)
		return err
	}

	l := strings.Split(cwd, string(filepath.Separator))
	dataDirPieces := append(l[:len(l)-3], "testdata")
	var volume string
	if runtime.GOOS == "windows" {
		volume = filepath.VolumeName(cwd)
		dataDirPieces = dataDirPieces[1:] // drop volume from list of dirs
	}
	dataDir := filepath.Join(dataDirPieces...)
	dataDir = fmt.Sprintf("%s%c%s", volume, filepath.Separator, dataDir) // prepend volume back

	files, err := ioutil.ReadDir(dataDir)
	if err != nil {
		fmt.Printf("Error occurred loading data files: %s", err)
		return err
	}

	destLoc, _ := os.Getwd()
	for _, file := range files {

		sFileName := filepath.Join(dataDir, file.Name())
		fmt.Printf("Copying %s to %s\n", sFileName, destLoc)

		sFile, err := os.Open(sFileName)
		if err != nil {
			fmt.Printf("Error occurred opening file: %s", err)
			return err
		}
		defer sFile.Close()

		dFileName := file.Name()
		dFile, err := os.Create(dFileName)
		if err != nil {
			fmt.Printf("Error occurred creating file: %s", err)
			return err
		}
		defer dFile.Close()

		_, err = io.Copy(dFile, sFile) // ignore num of bytes
		if err != nil {
			fmt.Printf("Error occurred copying file: %s", err)
			return err
		}

		err = dFile.Sync()
		if err != nil {
			fmt.Printf("Error occurred syncing file: %s", err)
			return err
		}
	}

	return nil
}

func ParseFlags() {

	flag.StringVar(&bundleLocation, "bundle-location", "", "Path to the bundle to be used in tests")
	flag.StringVar(&pullSecretFile, "pull-secret-file", "", "Path to the file containing pull secret")
	flag.StringVar(&CRCExecutable, "crc-binary", "", "Path to the CRC executable to be tested")
	flag.StringVar(&bundleVersion, "bundle-version", "", "Version of the bundle used in tests")
}
