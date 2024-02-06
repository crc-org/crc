package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	crcConfig "github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/machine"
	"github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/crc/ssh"
	"github.com/crc-org/crc/v2/pkg/download"
)

var (
	CRCHome string
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
	files, err := os.ReadDir(resourcesPath)
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

	filename, err := download.Download(bundleLocation, bundleDestination, 0644, nil)
	fmt.Printf("Downloading bundle from %s to %s.\n", bundleLocation, bundleDestination)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func RemoveCRCHome() error {
	keepFile := filepath.Join(CRCHome, ".keep")
	_, err := os.Stat(keepFile)
	if err != nil { // cannot get keepFile's status
		err = os.RemoveAll(CRCHome)

		if err != nil {
			fmt.Printf("Problem deleting CRC home folder %s.\n", CRCHome)
			return err
		}

		fmt.Printf("Deleted CRC home folder %s.\n", CRCHome)
		return nil

	}
	// keepFile exists
	return fmt.Errorf("folder %s not removed as per request: %s present", CRCHome, keepFile)
}

func RemoveCRCConfig() error {
	configFile := filepath.Join(CRCHome, "crc.json")

	err := os.RemoveAll(configFile)

	if err != nil {
		fmt.Printf("Problem deleting CRC config file %s.\n", configFile)
		return err
	}

	return nil
}

// MatchWithRetry will execute match function with expression as arg
// for #iterations with a timeout
func MatchWithRetry(expression string, match func(string) error, iterations, timeoutInSeconds int) error {
	return MatchRepetitionsWithRetry(expression, match, 1, iterations, timeoutInSeconds)
}

// MatchRepetitionsWithRetry will execute match function with expression as arg
// for #iterations with a timeout, expression should be matched # matchRepetitions in a row
func MatchRepetitionsWithRetry(expression string, match func(string) error, matchRepetitions int, iterations, timeoutInSeconds int) error {
	timeout := time.After(time.Duration(timeoutInSeconds) * time.Second)
	tick := time.NewTicker(time.Duration(timeoutInSeconds/iterations) * time.Second)
	matchRepetition := 0
	for {
		select {
		case <-timeout:
			tick.Stop()
			return fmt.Errorf("not found: %s. Timeout", expression)
		case <-tick.C:
			if err := match(expression); err == nil {
				matchRepetition++
				if matchRepetition == matchRepetitions {
					tick.Stop()
					return nil
				}
			} else {
				// repetions should be matched in a row, otherwise reset the counter
				matchRepetition = 0
			}
		}
	}
}

// GetBundlePath returns a path to the cached bundle, depending on the preset
func GetBundlePath(preset preset.Preset) string {
	bundle := constants.GetDefaultBundle(preset)
	return filepath.Join(CRCHome, "cache", bundle)

}

// WriteTempFile returns full path of the temp file it created, and an error
func WriteTempFile(content string, name string) (string, error) {
	tmpFile, err := os.CreateTemp("", name)
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()
	_, err = tmpFile.WriteString(content)
	return tmpFile.Name(), err
}

// Send command to CRC VM via SSH
func SendCommandToVM(cmd string) (string, error) {
	client := machine.NewClient(constants.DefaultName, false,
		crcConfig.New(crcConfig.NewEmptyInMemoryStorage(), crcConfig.NewEmptyInMemorySecretStorage()),
	)
	connectionDetails, err := client.ConnectionDetails()
	if err != nil {
		return "", err
	}
	ssh, err := ssh.NewClient(connectionDetails.SSHUsername, connectionDetails.IP, connectionDetails.SSHPort, connectionDetails.SSHKeys...)
	if err != nil {
		return "", err
	}
	out, _, err := ssh.Run(cmd)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func AddOCToPath() error {
	path := os.ExpandEnv("${HOME}/.crc/bin/oc:$PATH")
	if runtime.GOOS == "windows" {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		unexpandedPath := filepath.Join(userHomeDir, ".crc/bin/oc;${PATH}")
		path = os.ExpandEnv(unexpandedPath)
	}
	err := os.Setenv("PATH", path)
	if err != nil {
		return err
	}

	return nil
}

// LoginToOcCluster logs into the cluster as admin with oc command
// 'options' should have a form of a string slice like: [--option1 --option2 --option3] (string slice)
func LoginToOcCluster(options []string) error {

	credentialsCommand := "crc console --credentials" //#nosec G101
	err := ExecuteCommand(credentialsCommand)
	if err != nil {
		return err
	}
	out := GetLastCommandOutput("stdout")
	ocLoginAsAdminCommand := strings.Split(out, "'")[3]

	for _, option := range options {
		ocLoginAsAdminCommand = ocLoginAsAdminCommand + " " + option
	}

	return ExecuteCommand(ocLoginAsAdminCommand)
}

// LoginToOcClusterSucceedsOrFails is a wrapper for LoginToOcCluster
func LoginToOcClusterSucceedsOrFails(expected string) error {

	if expected == "fails" {
		err := LoginToOcCluster([]string{})
		if err != nil {
			return nil
		}
		_ = LogMessage("error:", "Login succeeded but was not supposed to")
		return fmt.Errorf("Login succeeded but was not supposed to")
	}

	return LoginToOcCluster([]string{})
}
