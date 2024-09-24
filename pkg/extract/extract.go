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
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	crcos "github.com/crc-org/crc/v2/pkg/os"
	"github.com/crc-org/crc/v2/pkg/os/terminal"

	"github.com/h2non/filetype"
	"github.com/klauspost/compress/zstd"
	"github.com/pkg/errors"
	"github.com/xi2/xz"
)

const minSizeForProgressBar = 100_000_000

func UncompressWithFilter(tarball, targetDir string, fileFilter func(string) bool) ([]string, error) {
	return uncompress(tarball, targetDir, fileFilter, false) // never show detailed output
}

func Uncompress(tarball, targetDir string) ([]string, error) {
	return uncompress(tarball, targetDir, nil, terminal.IsShowTerminalOutput())
}

func uncompress(tarball, targetDir string, fileFilter func(string) bool, showProgress bool) ([]string, error) {
	logging.Debugf("Uncompressing %s to %s", tarball, targetDir)

	if strings.HasSuffix(tarball, ".zip") {
		return unzip(tarball, targetDir, fileFilter, terminal.IsShowTerminalOutput())
	}

	file, err := os.Open(filepath.Clean(tarball))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "cannot read file information")
	}
	header := make([]byte, min(262, stat.Size()))
	if _, err := io.ReadFull(file, header); err != nil {
		return nil, errors.Wrap(err, "cannot determine type by reading file header")
	}
	if _, err := file.Seek(0, 0); err != nil {
		return nil, errors.Wrap(err, "cannot seek file")
	}

	switch {
	case filetype.Is(header, "xz"):
		reader, err := xz.NewReader(file, 0)
		if err != nil {
			return nil, err
		}
		return untar(reader, targetDir, fileFilter, showProgress)
	case filetype.Is(header, "zst"):
		reader, err := zstd.NewReader(file)
		if err != nil {
			return nil, err
		}
		return untar(reader, targetDir, fileFilter, showProgress)
	case filetype.Is(header, "gz"):
		reader, err := gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return untar(io.Reader(reader), targetDir, fileFilter, showProgress)
	case filetype.Is(header, "zip"):
		return unzip(tarball, targetDir, fileFilter, showProgress)
	case filetype.Is(header, "tar"):
		return untar(file, targetDir, fileFilter, showProgress)
	default:
		return nil, fmt.Errorf("Unknown file format when trying to uncompress %s", tarball)
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
		path, err := buildPath(targetDir, header.Name)
		if err != nil {
			return nil, err
		}

		if fileFilter != nil && !fileFilter(path) {
			logging.Debugf("untar: Skipping %s", path)
			continue
		}

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if err := os.MkdirAll(path, header.FileInfo().Mode()); err != nil {
				return nil, err
			}

		// if it's a file create it
		case tar.TypeReg, tar.TypeGNUSparse:
			// tar.Next() will externally only iterate files, so we might have to create intermediate directories here
			if err := uncompressFile(tarReader, header.FileInfo(), path, showProgress); err != nil {
				return nil, err
			}
			extractedFiles = append(extractedFiles, path)
		}
	}
}

func uncompressFile(tarReader io.Reader, fileInfo os.FileInfo, path string, showProgress bool) error {
	// with a file filter, we may have skipped the intermediate directories, make sure they exist
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fileInfo.Mode())
	if err != nil {
		return err
	}
	defer file.Close()

	reader, cleanup := progressBarReader(tarReader, fileInfo, showProgress)
	defer cleanup()

	_, err = crcos.CopySparse(file, reader)
	if err != nil {
		return err
	}
	return file.Close()
}

func buildPath(baseDir, filename string) (string, error) {
	path := filepath.Join(baseDir, filename) // #nosec G305

	// Check for ZipSlip. More Info: https://snyk.io/research/zip-slip-vulnerability
	baseDir = filepath.Clean(baseDir)
	if path != baseDir && !strings.HasPrefix(path, baseDir+string(os.PathSeparator)) {
		return "", fmt.Errorf("%s: illegal file path (expected prefix: %s)", path, baseDir+string(os.PathSeparator))
	}

	return path, nil
}
func unzip(archive, target string, fileFilter func(string) bool, showProgress bool) ([]string, error) {
	var extractedFiles []string
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(target, 0750); err != nil {
		return nil, err
	}

	for _, file := range reader.File {
		path, err := buildPath(target, file.Name)
		if err != nil {
			return nil, err
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

		if err := unzipFile(file, filepath.Clean(path), showProgress); err != nil {
			return nil, err
		}
		extractedFiles = append(extractedFiles, path)
	}

	return extractedFiles, nil
}

func unzipFile(file *zip.File, path string, showProgress bool) error {
	fileReader, err := file.Open()
	if err != nil {
		return err
	}
	defer fileReader.Close()

	return uncompressFile(fileReader, file.FileInfo(), path, showProgress)
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
