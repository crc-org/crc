package embed

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/YourFin/binappend"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/version"
)

func openEmbeddedFile(executablePath, embedName string) (io.ReadCloser, error) {
	if version.IsInstaller() {
		path := filepath.Join(version.InstallPath(), embedName)
		if _, err := os.Stat(path); err == nil {
			return os.Open(path)
		}
	}

	extractor, err := binappend.MakeExtractor(executablePath)
	if err != nil {
		return nil, fmt.Errorf("Could not data embedded in %s: %v", executablePath, err)
	}
	reader, err := extractor.GetReader(embedName)
	if err != nil {
		return nil, fmt.Errorf("Could not open embedded '%s' in %s: %v", embedName, executablePath, err)
	}
	return reader, nil
}

func Extract(embedName, destFile string) error {
	executablePath, err := os.Executable()
	if err != nil {
		return err
	}
	return ExtractFromExecutable(executablePath, embedName, destFile)
}

func ExtractFromExecutable(executablePath, embedName, destFile string) error {
	logging.Debugf("Extracting embedded '%s' from %s to %s", embedName, executablePath, destFile)
	reader, err := openEmbeddedFile(executablePath, embedName)
	if err != nil {
		return err
	}

	defer reader.Close()
	writer, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("Could not create '%s': %v", destFile, err)
	}
	defer writer.Close()

	_, err = io.Copy(writer, reader)
	if err != nil {
		return fmt.Errorf("Failed to copy embedded '%s' from %s to %s: %v", embedName, executablePath, destFile, err)
	}
	return writer.Close()
}
