package embed

import (
	"fmt"
	"io"
	"os"

	"github.com/YourFin/binappend"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
)

func openEmbeddedFile(executablePath, embedName string) (io.ReadCloser, error) {
	extractor, err := binappend.MakeExtractor(executablePath)
	if err != nil {
		return nil, fmt.Errorf("could not find data embedded in %s: %w", executablePath, err)
	}
	reader, err := extractor.GetReader(embedName)
	if err != nil {
		return nil, fmt.Errorf("could not open embedded '%s' in %s: %w", embedName, executablePath, err)
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
		return fmt.Errorf("could not create '%s': %w", destFile, err)
	}
	defer writer.Close()

	_, err = io.Copy(writer, reader)
	if err != nil {
		return fmt.Errorf("failed to copy embedded '%s' from %s to %s: %w", embedName, executablePath, destFile, err)
	}
	return writer.Close()
}
