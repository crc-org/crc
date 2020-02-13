// +build integration

package crcsuite

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/code-ready/crc/pkg/download"
)

var wg sync.WaitGroup

type proxy struct {
	HttpProxy  string
	HttpsProxy string
}

type cluster struct {
	KubeConfig    string
	KubeAdminPass string
	ClusterAPI    string
	WebConsoleURL string
	ProxyConfig   proxy
}

type message struct {
	Name             string
	CrcStatus        string
	OpenshiftStatus  string
	DiskUse          int
	DiskSize         int
	Error            string
	Success          bool
	CrcVersion       string
	CommitSha        string
	OpenshiftVersion string
	ClusterConfig    cluster
	KubeletStarted   bool
	State            int
}

func (m *message) recordResponse() bool {

	// record response to JSON file ~/.crc/answer.json
	msgJson, _ := json.Marshal(m)
	dest := filepath.Join(CRCHome, "answer.json")
	err := ioutil.WriteFile(dest, msgJson, 0644)

	if err != nil {
		fmt.Printf("Error recording response: %v\n", err)
		return false
	}

	return true
}

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

func CopyFilesToTestDir() {

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error retrieving current dir: %s", err)
		os.Exit(1)
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
		os.Exit(1)
	}

	destLoc, _ := os.Getwd()
	for _, file := range files {

		sFileName := filepath.Join(dataDir, file.Name())
		fmt.Printf("Copying %s to %s\n", sFileName, destLoc)

		sFile, err := os.Open(sFileName)
		if err != nil {
			fmt.Printf("Error occurred opening file: %s", err)
			os.Exit(1)
		}
		defer sFile.Close()

		dFileName := file.Name()
		dFile, err := os.Create(dFileName)
		if err != nil {
			fmt.Printf("Error occurred creating file: %s", err)
			os.Exit(1)
		}
		defer dFile.Close()

		_, err = io.Copy(dFile, sFile) // ignore num of bytes
		if err != nil {
			fmt.Printf("Error occurred copying file: %s", err)
			os.Exit(1)
		}

		err = dFile.Sync()
		if err != nil {
			fmt.Printf("Error occurred syncing file: %s", err)
			os.Exit(1)
		}
	}
}

func SockReader(r io.Reader, command string, reader_err chan error) {

	defer wg.Done() // schedule call to WaitGroup's Done to announce goroutine finished

	d := json.NewDecoder(r)

	var msg message

	err := d.Decode(&msg)

	if err == io.EOF {
		fmt.Printf("End of File exception: %s\n", err)
		reader_err <- err
	} else if err != nil {
		fmt.Printf("Unexpected error: %s\n", err)
		reader_err <- err
	}

	// record response to JSON file ~/.crc/answer.json
	msgJson, _ := json.Marshal(msg)
	dest := filepath.Join(CRCHome, "answer.json")
	err = ioutil.WriteFile(dest, msgJson, 0644)

	reader_err <- err
}

func ParseFlags() {

	flag.StringVar(&bundleURL, "bundle-location", "embedded", "Path to the bundle to be used in tests")
	flag.StringVar(&pullSecretFile, "pull-secret-file", "", "Path to the file containing pull secret")
	flag.StringVar(&CRCBinary, "crc-binary", "", "Path to the CRC binary to be tested")
	flag.StringVar(&bundleVersion, "bundle-version", "", "Version of the bundle used in tests")
}
