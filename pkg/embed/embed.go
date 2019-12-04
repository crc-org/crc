package embed

import (
	"fmt"
	"io"
	"os"

	"github.com/code-ready/crc/pkg/crc/logging"

	"github.com/YourFin/binappend"
)

func openEmbeddedFile(binaryPath, embedName string) (*binappend.Reader, error) {
	extractor, err := binappend.MakeExtractor(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("Could not data embedded in %s: %v", binaryPath, err)
	}
	reader, err := extractor.GetReader(embedName)
	if err != nil {
		return nil, fmt.Errorf("Could not open embedded '%s' in %s: %v", embedName, binaryPath, err)
	}
	return reader, nil
}

func Extract(embedName, destFile string) error {
	binaryPath, err := os.Executable()
	if err != nil {
		return err
	}
	return ExtractFromBinary(binaryPath, embedName, destFile)
}

func ExtractFromBinary(binaryPath, embedName, destFile string) error {
	logging.Debugf("Extracting embedded '%s' from %s to %s", embedName, binaryPath, destFile)
	reader, err := openEmbeddedFile(binaryPath, embedName)
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
		return fmt.Errorf("Failed to copy embedded '%s' from %s to %s: %v", embedName, binaryPath, destFile, err)
	}
	return nil
}
