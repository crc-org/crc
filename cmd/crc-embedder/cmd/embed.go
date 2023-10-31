package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/download"

	"github.com/crc-org/crc/v2/pkg/crc/machine/libvirt"
	"github.com/crc-org/crc/v2/pkg/crc/machine/vfkit"

	"github.com/YourFin/binappend"
	"github.com/spf13/cobra"
)

var (
	goos       string
	cacheDir   string
	components []string
	noDownload bool
)

const (
	vfkitDriver        = "vfkit-driver"
	vfkitEntitlement   = "vfkit-entitlement"
	libvirtDriver      = "libvirt-driver"
	adminHelper        = "admin-helper"
	backgroundLauncher = "background-launcher"
)

func init() {
	embedCmd.Flags().StringVar(&goos, "goos", runtime.GOOS, "Target platform (darwin, linux or windows)")
	embedCmd.Flags().StringVar(&cacheDir, "cache-dir", "", "Destination directory for the downloaded files")
	embedCmd.Flags().BoolVar(&noDownload, "no-download", false, "Only embed files, don't download")
	embedCmd.Flags().StringSliceVar(&components, "components", []string{}, fmt.Sprintf("List of component(s) to download (%s)", strings.Join(getAllComponentNames(goos), ", ")))
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
		cacheDir, err = os.MkdirTemp("", "crc-embedder")
		if err != nil {
			logging.Fatalf("Failed to create temporary directory: %v", err)
		}
		defer os.RemoveAll(cacheDir)
	}
	var embedFileList []string
	if noDownload {
		embedFileList = getEmbedFileList(goos, cacheDir)
	} else {
		embedFileList, err = downloadDataFiles(goos, components, cacheDir)
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
	dataFileUrls = map[string]map[string]remoteFileInfo{
		"darwin": {
			vfkitDriver:      {vfkit.VfkitDownloadURL, 0755},
			vfkitEntitlement: {vfkit.VfkitEntitlementsURL, 0644},
			adminHelper:      {constants.GetAdminHelperURLForOs("darwin"), 0755},
		},
		"linux": {
			libvirtDriver: {libvirt.MachineDriverDownloadURL, 0755},
			adminHelper:   {constants.GetAdminHelperURLForOs("linux"), 0755},
		},
		"windows": {
			adminHelper:        {constants.GetAdminHelperURLForOs("windows"), 0755},
			backgroundLauncher: {constants.GetWin32BackgroundLauncherDownloadURL(), 0755},
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

func getAllComponentNames(goos string) []string {
	var components []string
	for component := range dataFileUrls[goos] {
		components = append(components, component)
	}
	return components
}

func shouldDownload(components []string, component string) bool {
	if len(components) == 0 {
		return true
	}
	for _, v := range components {
		if v == component {
			return true
		}
	}
	return false
}

func downloadDataFiles(goos string, components []string, destDir string) ([]string, error) {
	downloadedFiles := []string{}
	downloads := dataFileUrls[goos]

	for componentName, dl := range downloads {
		if !shouldDownload(components, componentName) {
			continue
		}
		filename, err := download.Download(dl.url, destDir, dl.permissions, nil)
		if err != nil {
			return nil, err
		}
		downloadedFiles = append(downloadedFiles, filename)
	}

	if len(components) != 0 && len(components) != len(downloadedFiles) {
		logging.Warnf("invalid components requested, supported components are: %s", strings.Join(getAllComponentNames(goos), ", "))
	}

	return downloadedFiles, nil
}
