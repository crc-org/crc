package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/download"

	"github.com/code-ready/crc/pkg/crc/machine/hyperkit"
	"github.com/code-ready/crc/pkg/crc/machine/libvirt"

	"github.com/YourFin/binappend"
	"github.com/spf13/cobra"
)

var (
	bundleDir string
	goos      string
)

func init() {
	embedCmd.Flags().StringVar(&bundleDir, "bundle-dir", constants.MachineCacheDir, "Directory where the OpenShift bundle can be found")
	embedCmd.Flags().StringVar(&goos, "goos", runtime.GOOS, "Target platform (darwin, linux or windows)")
	rootCmd.AddCommand(embedCmd)
}

var embedCmd = &cobra.Command{
	Args:  cobra.ExactArgs(1),
	Use:   "embed [path to the (non-embedded) crc executable]",
	Short: "Embed data files in crc executable",
	Long:  `Embed the OpenShift bundle and the binaries needed at runtime in the crc executable`,
	Run: func(cmd *cobra.Command, args []string) {
		runEmbed(args)
	},
}

func runEmbed(args []string) {
	executablePath := args[0]
	destDir, err := ioutil.TempDir("", "crc-embedder")
	if err != nil {
		logging.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(destDir)
	downloadedFiles, err := downloadDataFiles(goos, destDir)
	if err != nil {
		logging.Fatalf("Failed to download data files: %v", err)
	}

	bundlePath := path.Join(bundleDir, constants.GetDefaultBundleForOs(goos))
	downloadedFiles = append(downloadedFiles, bundlePath)
	err = embedFiles(executablePath, downloadedFiles)
	if err != nil {
		logging.Fatalf("Failed to embed data files: %v", err)
	}
}

func embedFiles(executablePath string, filenames []string) error {
	appender, err := binappend.MakeAppender(executablePath)
	if err != nil {
		return err
	}
	defer appender.Close()
	for _, filename := range filenames {
		logging.Debugf("Embedding %s in %s", filename, executablePath)
		f, err := os.Open(filename) // #nosec G304
		if err != nil {
			return fmt.Errorf("Failed to open %s: %v", filename, err)
		}
		defer f.Close()

		err = appender.AppendStreamReader(path.Base(filename), f, false)
		if err != nil {
			return fmt.Errorf("Failed to append %s to %s: %v", filename, executablePath, err)
		}
	}

	return nil
}

type remoteFileInfo struct {
	url         string
	permissions os.FileMode
}

var (
	dataFileUrls = map[string][]remoteFileInfo{
		"darwin": {
			{hyperkit.MachineDriverDownloadURL, 0755},
			{hyperkit.HyperKitDownloadURL, 0755},
			{hyperkit.QcowToolDownloadURL, 0755},
			{constants.GetCRCMacTrayDownloadURL(), 0644},
			{constants.GetAdminHelperURLForOs("darwin"), 0755},
		},
		"linux": {
			{libvirt.MachineDriverDownloadURL, 0755},
			{constants.GetAdminHelperURLForOs("linux"), 0755},
		},
		"windows": {
			{constants.GetAdminHelperURLForOs("windows"), 0755},
			{constants.GetCRCWindowsTrayDownloadURL(), 0644},
		},
	}
)

func downloadDataFiles(goos string, destDir string) ([]string, error) {
	downloadedFiles := []string{}
	downloads := dataFileUrls[goos]
	for _, dl := range downloads {
		filename, err := download.Download(dl.url, destDir, dl.permissions, nil)
		if err != nil {
			return nil, err
		}
		downloadedFiles = append(downloadedFiles, filename)
	}

	return downloadedFiles, nil
}
