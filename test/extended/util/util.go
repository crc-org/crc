package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/download"
)

func CopyFilesToTestDir() error {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error retrieving current dir: %s", err)
		return err
	}

	l := strings.Split(cwd, string(filepath.Separator))
	dataDirPieces := l[:len(l)-3]
	dataDirPieces = append(dataDirPieces, "testdata")
	var volume string
	if runtime.GOOS == "windows" {
		volume = filepath.VolumeName(cwd)
		dataDirPieces = dataDirPieces[1:] // drop volume from list of dirs
	}
	dataDir := filepath.Join(dataDirPieces...)
	dataDir = fmt.Sprintf("%s%c%s", volume, filepath.Separator, dataDir) // prepend volume back

	return CopyResourcesFromPath(dataDir)
}

func CopyResourcesFromPath(resourcesPath string) error {
	files, err := ioutil.ReadDir(resourcesPath)
	if err != nil {
		fmt.Printf("Error occurred loading data files: %s", err)
		return err
	}
	destLoc, _ := os.Getwd()
	for _, file := range files {

		sFileName := filepath.Join(resourcesPath, file.Name())
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
		err = dFile.Close()
		if err != nil {
			fmt.Printf("Error closing file: %s", err)
			return err
		}
	}
	return nil
}

// Download bundle for testing
func DownloadBundle(bundleLocation string, bundleDestination string, bundleName string) (string, error) {

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

func RemoveCRCHome(crcHome string) error {
	keepFile := filepath.Join(crcHome, ".keep")
	_, err := os.Stat(keepFile)
	if err != nil { // cannot get keepFile's status
		err = os.RemoveAll(crcHome)

		if err != nil {
			fmt.Printf("Problem deleting CRC home folder %s.\n", crcHome)
			return err
		}

		fmt.Printf("Deleted CRC home folder %s.\n", crcHome)
		return nil

	}
	// keepFile exists
	return fmt.Errorf("folder %s not removed as per request: %s present", crcHome, keepFile)
}

// Based on the number of iterations for a given timeout in seconds the function returns the duration of echa loop
// and the extra time in case required to complete the timeout
func GetRetryParametersFromTimeoutInSeconds(iterations int, timeout string) (time.Duration, time.Duration, error) {
	totalTime, err := strconv.Atoi(timeout)
	if err != nil {
		return 0, 0, err
	}
	iterationDuration, err :=
		time.ParseDuration(strconv.Itoa(totalTime/iterations) + "s")
	if err != nil {
		return 0, 0, err
	}
	extraTime := totalTime % iterations
	if extraTime != 0 {
		extraTimeDuration, err :=
			time.ParseDuration(strconv.Itoa(extraTime) + "s")
		if err != nil {
			return 0, 0, err
		}
		return iterationDuration, extraTimeDuration, nil
	}
	return iterationDuration, 0, nil
}
