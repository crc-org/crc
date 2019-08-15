package extract

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/xi2/xz"
)

func Uncompress(tarball, targetDir string) error {
	logging.Debugf("Uncompressing %s to %s", tarball, targetDir)

	if strings.HasSuffix(tarball, ".zip") {
		return Unzip(tarball, targetDir)
	}

	var filereader io.Reader
	file, err := os.Open(filepath.Clean(tarball))
	if err != nil {
		return err
	}
	defer file.Close()

	if strings.HasSuffix(tarball, ".tar.xz") || strings.HasSuffix(tarball, ".crcbundle") {
		filereader, err = xz.NewReader(file, 0)
		if err != nil {
			return err
		}
	} else if strings.HasSuffix(tarball, ".tar.gz") {
		reader, err := gzip.NewReader(file)
		if err != nil {
			return err
		}
		defer reader.Close()
		filereader = io.Reader(reader)
	} else if strings.HasSuffix(tarball, ".tar") {
		filereader = file
	} else {
		logging.Warnf("Unknown file format when trying to uncompress %s", tarball)
	}

	return untar(filereader, targetDir)
}

func untar(reader io.Reader, targetDir string) error {
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		switch {
		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		path := filepath.Join(targetDir, header.Name)

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(path); err != nil {
				if err := os.MkdirAll(path, header.FileInfo().Mode()); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg, tar.TypeGNUSparse:
			// tar.Next() will externally only iterate files, so we might have to create intermediate directories here
			if err = os.MkdirAll(filepath.Dir(path), 0750); err != nil {
				return err
			}
			file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, header.FileInfo().Mode())
			if err != nil {
				return err
			}
			defer file.Close()

			// copy over contents
			if _, err := io.Copy(file, tarReader); err != nil {
				return err
			}
		}
	}
}

func Unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(target, 0750); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name) // #nosec G305

		// Check for ZipSlip. More Info: https://snyk.io/research/zip-slip-vulnerability
		if !strings.HasPrefix(path, filepath.Clean(target)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", path)
		}

		if file.FileInfo().IsDir() {
			err = os.MkdirAll(path, file.Mode())
			if err != nil {
				return err
			}
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}
