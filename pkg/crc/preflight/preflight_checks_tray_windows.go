package preflight

import (
	"fmt"
	"io/ioutil"
	goos "os"
	"path/filepath"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/input"
	"github.com/code-ready/crc/pkg/crc/logging"
	dl "github.com/code-ready/crc/pkg/download"
	"github.com/code-ready/crc/pkg/embed"
	"github.com/code-ready/crc/pkg/extract"
	"github.com/code-ready/crc/pkg/os"
	"github.com/code-ready/crc/pkg/os/windows/powershell"
)

func checkIfTrayInstalled() error {
	/* We want to force installation whenever setup is ran
	 * as we want the service config to point to the executable
	 * with which the setup command was issued.
	 */

	return fmt.Errorf("Ignoring check and forcing installation of System Tray")
}

func fixTrayInstalled() error {
	/* To avoid asking for elevated privileges again and again
	 * this takes care of all the steps needed to have a running
	 * tray by invoking a ps script which does the following:
	 * a) Add logon as service permission for the user
	 * b) Create the daemon service and start it
	 * c) Add tray to startup folder and start it
	 */

	tempDir, err := ioutil.TempDir("", "crc")
	if err != nil {
		logging.Error("Failed creating temporary directory for tray installation")
		return err
	}

	defer func() {
		_ = goos.RemoveAll(tempDir)
	}()

	// prepare the ps script
	binPath, err := goos.Executable()
	if err != nil {
		return fmt.Errorf("Unable to find the current executables location: %v", err)
	}
	binPathWithArgs := fmt.Sprintf("%s daemon", strings.TrimSpace(binPath))

	// get the password from user
	password, err := input.PromptUserForSecret("Enter account login password for service installation", "This is the login password of your current account, needed to install the daemon service")
	if err != nil {
		return fmt.Errorf("Unable to get login password: %v", err)
	}

	// sanitize password
	password = escapeWindowsPassword(password)

	psScriptContent := genTrayInstallScript(
		password,
		tempDir,
		binPathWithArgs,
		constants.TrayExecutablePath,
		constants.TrayShortcutName,
		constants.DaemonServiceName,
	)
	psFilePath := filepath.Join(tempDir, "trayInstallation.ps1")

	// write temporary ps script
	if err = writePsScriptContentToFile(psScriptContent, psFilePath); err != nil {
		return err
	}

	// invoke the ps script
	_, _, err = powershell.ExecuteAsAdmin("Installing System Tray for CodeReady Containers", psFilePath)
	// wait for the script to finish executing
	time.Sleep(2 * time.Second)
	if err != nil {
		logging.Debug("Failed to execute tray installation script")
		return err
	}

	// check for 'success' file
	if _, err = goos.Stat(filepath.Join(tempDir, "success")); goos.IsNotExist(err) {
		return fmt.Errorf("Installation script didn't execute successfully: %v", err)
	}
	return nil
}

func escapeWindowsPassword(password string) string {
	// escape specials characters (|`|$|"|') with '`' if present in password
	if strings.Contains(password, "`") {
		password = strings.ReplaceAll(password, "`", "``")
	}
	if strings.Contains(password, "$") {
		password = strings.ReplaceAll(password, "$", "`$")
	}
	if strings.Contains(password, "\"") {
		password = strings.ReplaceAll(password, "\"", "`\"")
	}
	if strings.Contains(password, "'") {
		password = strings.ReplaceAll(password, "'", "`'")
	}
	return password
}

func removeTray() error {
	trayProcessName := constants.TrayExecutableName[:len(constants.TrayExecutableName)-4]

	tempDir, err := ioutil.TempDir("", "crc")
	if err != nil {
		logging.Debug("Failed to create temporary directory for System Tray removal")
		return nil
	}
	defer func() {
		_ = goos.RemoveAll(tempDir)
	}()

	psScriptContent := genTrayRemovalScript(
		trayProcessName,
		constants.TrayShortcutName,
		constants.DaemonServiceName,
		tempDir,
	)
	psFilePath := filepath.Join(tempDir, "trayRemoval.ps1")

	// write script content to temporary file
	if err = writePsScriptContentToFile(psScriptContent, psFilePath); err != nil {
		logging.Debug(err)
		return nil
	}

	_, _, err = powershell.ExecuteAsAdmin("Uninstalling CodeReady Containers System Tray", psFilePath)
	// wait for the script to finish executing
	time.Sleep(2 * time.Second)
	if err != nil {
		logging.Debugf("Unable to execute System Tray uninstall script: %v", err)
	}
	return nil
}

func checkTrayExecutableExists() error {
	if os.FileExists(constants.TrayExecutablePath) {
		return nil
	}
	return fmt.Errorf("Tray executable does not exists")
}

func fixTrayExecutableExists() error {
	tmpArchivePath, err := ioutil.TempDir("", "crc")
	if err != nil {
		logging.Error("Failed creating temporary directory for extracting tray")
		return err
	}
	defer func() {
		_ = goos.RemoveAll(tmpArchivePath)
	}()

	logging.Debug("Trying to extract tray from crc executable")
	err = embed.Extract(filepath.Base(constants.GetCRCWindowsTrayDownloadURL()), tmpArchivePath)
	if err != nil {
		logging.Debug("Could not extract tray from crc executable", err)
		logging.Debug("Downloading crc tray")
		_, err = dl.Download(constants.GetCRCWindowsTrayDownloadURL(), tmpArchivePath, 0600)
		if err != nil {
			return err
		}
	}
	archivePath := filepath.Join(tmpArchivePath, filepath.Base(constants.GetCRCWindowsTrayDownloadURL()))
	_, err = extract.Uncompress(archivePath, constants.TrayExecutableDir, false)
	if err != nil {
		return fmt.Errorf("Cannot uncompress '%s': %v", archivePath, err)
	}

	return nil
}

func writePsScriptContentToFile(psScriptContent, psFilePath string) error {
	psFile, err := goos.Create(psFilePath)
	if err != nil {
		logging.Debugf("Unable to create file to write scipt content: %v", err)
		return err
	}
	defer psFile.Close()
	// write the ps script
	/* Add UTF-8 BOM at the beginning of the script so that Windows
	 * correctly detects the file encoding
	 */
	_, err = psFile.Write([]byte{0xef, 0xbb, 0xbf})
	if err != nil {
		logging.Debugf("Unable to write script content to file: %v", err)
		return err
	}
	_, err = psFile.WriteString(psScriptContent)
	if err != nil {
		logging.Debugf("Unable to write script content to file: %v", err)
		return err
	}

	return nil
}
