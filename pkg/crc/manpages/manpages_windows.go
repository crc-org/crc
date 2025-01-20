package manpages

import "github.com/spf13/cobra/doc"

var CrcManPageHeader = &doc.GenManHeader{
	Title:   "CRC",
	Section: "1",
}

// GenerateManPages generates man pages for cli
// This method is a no-operation placeholder, it's used to make project compile.
func GenerateManPages(_ func(targetDir string) error, _ string) error {
	return nil
}

// RemoveCrcManPages cleans up man pages directory
// This method is a no-operation placeholder, it's used to make project compile
func RemoveCrcManPages(_ string) error {
	return nil
}
