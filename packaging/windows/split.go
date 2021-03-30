package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const chunkSize = 1024 * 1024 * 1024 // 1GiB chunk size

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Split takes only one argument (the file to split)")
	}

	if err := split(os.Args[1]); err != nil {
		log.Fatal(err.Error())
	}
}

func split(filePath string) error {
	bundle, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer bundle.Close()
	bundleName := filepath.Base(filePath)
	for i := 0; ; i++ {
		partFileName := fmt.Sprintf("%s.%d", bundleName, i)
		partFile, err := os.Create(filepath.Join(filepath.Dir(filePath), partFileName))
		if err != nil {
			return err
		}
		defer partFile.Close()
		n, err := io.CopyN(partFile, bundle, chunkSize)
		fmt.Printf("Copied %d bytes from %s to %s\n", n, bundleName, partFileName)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}
