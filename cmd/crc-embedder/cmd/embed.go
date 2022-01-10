package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/download"

	"github.com/code-ready/crc/pkg/crc/machine/libvirt"
	"github.com/code-ready/crc/pkg/crc/machine/vfkit"

	"github.com/YourFin/binappend"
	"github.com/spf13/cobra"
)

var (
	goos       string
	cacheDir   string
	noDownload bool
)

func init() {
	embedCmd.Flags().StringVar(&goos, "goos", runtime.GOOS, "Target platform (darwin, linux or windows)")
	embedCmd.Flags().StringVar(&cacheDir, "cache-dir", "", "Destination directory for the downloaded files")
	embedCmd.Flags().BoolVar(&noDownload, "no-download", false, "Only embed files, don't download")
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
	var err error
	executablePath := args[0]
	if cacheDir == "" {
		cacheDir, err = ioutil.TempDir("", "crc-embedder")
		if err != nil {
			logging.Fatalf("Failed to create temporary directory: %v", err)
		}
		defer os.RemoveAll(cacheDir)
	}
	var embedFileList []string
	if noDownload {
		embedFileList = getEmbedFileList(goos, cacheDir)
	} else {
		embedFileList, err = downloadDataFiles(goos, cacheDir)
		if err != nil {
			logging.Fatalf("Failed to download data files: %v", err)
		}
	}
	err = embedFiles(executablePath, embedFileList)
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
			{vfkit.VfkitDownloadURL, 0755},
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

func getEmbedFileList(goos string, destDir string) []string {
	fileList := []string{}
	urls := dataFileUrls[goos]
	for _, dlDetails := range urls {
		filename := filepath.Base(dlDetails.url)
		fileList = append(fileList, filepath.Join(destDir, filename))
	}

	return fileList
}

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
