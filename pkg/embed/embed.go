package embed

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/code-ready/crc/pkg/crc/logging"
)

//go:generate go run -tags=build generator.go

type embedFiles struct {
	storage map[string][]byte
}

// Create new box for embed files
func newEmbedBox() *embedFiles {
	return &embedFiles{storage: make(map[string][]byte)}
}

// Add a file to box
func (e *embedFiles) Add(file string, content []byte) {
	e.storage[file] = content
}

// Get file's content
func (e *embedFiles) Get(file string) []byte {
	if f, ok := e.storage[file]; ok {
		return f
	}
	return nil
}

// Embed box expose
var box = newEmbedBox()

// Add a file content to box
func Add(file string, content []byte) {
	box.Add(file, content)
}

// Get a file from box
func Get(file string) []byte {
	return box.Get(file)
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
	content := Get(embedName)
	if content == nil {
		return fmt.Errorf("Not embed in %s", embedName)
	}

	if err := ioutil.WriteFile(destFile, content, 0600); err != nil {
		return fmt.Errorf("Failed to copy embedded '%s' from %s to %s: %v", embedName, executablePath, destFile, err)
	}
	return nil
}
