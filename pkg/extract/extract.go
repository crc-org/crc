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

	"github.com/cheggaaa/pb/v3"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/h2non/filetype"
	"github.com/pkg/errors"
	"github.com/xi2/xz"
	"golang.org/x/crypto/ssh/terminal"
)

const minSizeForProgressBar = 100_000_000

func UncompressWithFilter(tarball, targetDir string, showProgress bool, fileFilter func(string) bool) ([]string, error) {
	logging.Debugf("Uncompressing %s to %s", tarball, targetDir)

	file, err := os.Open(filepath.Clean(tarball))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "cannot read file information")
	}
	return uncompress(file, stat.Size(), targetDir, fileFilter, showProgress && terminal.IsTerminal(int(os.Stdout.Fd())))
}

func Uncompress(tarball, targetDir string, showProgress bool) ([]string, error) {
	return UncompressWithFilter(tarball, targetDir, showProgress, nil)
}

func uncompress(tarball io.ReaderAt, size int64, targetDir string, fileFilter func(string) bool, showProgress bool) ([]string, error) {
	reader := io.NewSectionReader(tarball, 0, size)

	header := make([]byte, min(262, size))
	if _, err := io.ReadFull(reader, header); err != nil {
		return nil, errors.Wrap(err, "cannot determine type by reading file header")
	}
	if _, err := reader.Seek(0, 0); err != nil {
		return nil, errors.Wrap(err, "cannot seek file")
	}

	switch {
	case filetype.Is(header, "xz"):
		reader, err := xz.NewReader(reader, 0)
		if err != nil {
			return nil, err
		}
		return untar(reader, targetDir, fileFilter, showProgress)
	case filetype.Is(header, "gz"):
		reader, err := gzip.NewReader(reader)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return untar(io.Reader(reader), targetDir, fileFilter, showProgress)
	case filetype.Is(header, "zip"):
		return unzip(reader, size, targetDir, fileFilter, showProgress)
	case filetype.Is(header, "tar"):
		return untar(reader, targetDir, fileFilter, showProgress)
	default:
		return nil, fmt.Errorf("Unknown file format when trying to uncompress to %s", targetDir)
	}
}

func min(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func untar(reader io.Reader, targetDir string, fileFilter func(string) bool, showProgress bool) ([]string, error) {
	var extractedFiles []string
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		switch {
		// if no more files are found return
		case err == io.EOF:
			return extractedFiles, nil

		// return any other error
		case err != nil:
			return extractedFiles, err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		path := filepath.Join(targetDir, header.Name)
		if fileFilter != nil && !fileFilter(path) {
			logging.Debugf("untar: Skipping %s", path)
			continue
		}

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(path); err != nil {
				if err := os.MkdirAll(path, header.FileInfo().Mode()); err != nil {
					return nil, err
				}
			}

		// if it's a file create it
		case tar.TypeReg, tar.TypeGNUSparse:
			// tar.Next() will externally only iterate files, so we might have to create intermediate directories here
			if err := untarFile(tarReader, header, path, showProgress); err != nil {
				return nil, err
			}
			extractedFiles = append(extractedFiles, path)
		}
	}
}

func untarFile(tarReader *tar.Reader, header *tar.Header, path string, showProgress bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, header.FileInfo().Mode())
	if err != nil {
		return err
	}
	defer file.Close()

	reader, cleanup := progressBarReader(tarReader, header.FileInfo(), showProgress)
	defer cleanup()

	// copy over contents
	// #nosec G110
	_, err = io.Copy(file, reader)
	return err
}

func unzip(archive io.ReaderAt, size int64, target string, fileFilter func(string) bool, showProgress bool) ([]string, error) {
	var extractedFiles []string
	reader, err := zip.NewReader(archive, size)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(target, 0750); err != nil {
		return nil, err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name) // #nosec G305

		// Check for ZipSlip. More Info: https://snyk.io/research/zip-slip-vulnerability
		if !strings.HasPrefix(path, filepath.Clean(target)+string(os.PathSeparator)) {
			return nil, fmt.Errorf("%s: illegal file path", path)
		}

		if fileFilter != nil && !fileFilter(path) {
			logging.Debugf("untar: Skipping %s", path)
			continue
		}
		if file.FileInfo().IsDir() {
			err = os.MkdirAll(path, file.Mode())
			if err != nil {
				return nil, err
			}
			continue
		}

		if err := unzipFile(file, path, showProgress); err != nil {
			return nil, err
		}
		extractedFiles = append(extractedFiles, path)
	}

	return extractedFiles, nil
}

func unzipFile(file *zip.File, path string, showProgress bool) error {
	// with a file filter, we may have skipped the intermediate directories, make sure they exist
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}

	fileReader, err := file.Open()
	if err != nil {
		return err
	}
	defer fileReader.Close()

	reader, cleanup := progressBarReader(fileReader, file.FileInfo(), showProgress)
	defer cleanup()

	targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer targetFile.Close()

	// #nosec G110
	_, err = io.Copy(targetFile, reader)
	return err
}

func progressBarReader(reader io.Reader, info os.FileInfo, showProgress bool) (io.Reader, func()) {
	if showProgress && info.Size() >= minSizeForProgressBar {
		bar := progressBar(info.Size(), info.Name())
		return bar.NewProxyReader(reader), func() { bar.Finish() }
	}
	return reader, func() {}
}

func progressBar(size int64, name string) *pb.ProgressBar {
	bar := pb.Simple.Start64(size)
	bar.Set("prefix", fmt.Sprintf("%s: ", filepath.Base(name)))
	return bar
}
