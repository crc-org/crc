package extract

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
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

func UncompressWithFilter(ctx context.Context, tarball, targetDir string, fileFilter func(string) bool) ([]string, error) {
	return uncompress(ctx, tarball, targetDir, fileFilter, false) // never show detailed output
}

func Uncompress(ctx context.Context, tarball, targetDir string) ([]string, error) {
	return uncompress(ctx, tarball, targetDir, nil, terminal.IsShowTerminalOutput())
}

func uncompress(ctx context.Context, tarball, targetDir string, fileFilter func(string) bool, showProgress bool) ([]string, error) {
	logging.Debugf("Uncompressing %s to %s", tarball, targetDir)

	if strings.HasSuffix(tarball, ".zip") {
		return unzip(ctx, tarball, targetDir, fileFilter, terminal.IsShowTerminalOutput())
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
		return untar(ctx, reader, targetDir, fileFilter, showProgress)
	case filetype.Is(header, "zst"):
		reader, err := zstd.NewReader(file)
		if err != nil {
			return nil, err
		}
		return untar(ctx, reader, targetDir, fileFilter, showProgress)
	case filetype.Is(header, "gz"):
		reader, err := gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return untar(ctx, io.Reader(reader), targetDir, fileFilter, showProgress)
	case filetype.Is(header, "zip"):
		return unzip(ctx, tarball, targetDir, fileFilter, showProgress)
	case filetype.Is(header, "tar"):
		return untar(ctx, file, targetDir, fileFilter, showProgress)
	default:
		return nil, fmt.Errorf("Unknown file format when trying to uncompress %s", tarball)
	}
}

func untar(ctx context.Context, reader io.Reader, targetDir string, fileFilter func(string) bool, showProgress bool) ([]string, error) {
	var extractedFiles []string
	tarReader := tar.NewReader(reader)

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return nil, err
	}
	targetDirRoot, err := os.OpenRoot(targetDir)
	if err != nil {
		return nil, err
	}
	defer targetDirRoot.Close()

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
		// path must not be absolute, so it is not joined to targetDir
		// and the filename is just cleaned
		path := filepath.Clean(header.Name)

		if fileFilter != nil && !fileFilter(path) {
			logging.Debugf("untar: Skipping %s", path)
			continue
		}

		// check the file type
		switch header.Typeflag {
		// if it's a dir, and it doesn't exist, create it
		case tar.TypeDir:
			if err := targetDirRoot.MkdirAll(path, header.FileInfo().Mode().Perm()); err != nil {
				return nil, err
			}

		// if it's a file, create it
		case tar.TypeReg, tar.TypeGNUSparse:
			// tar.Next() will externally only iterate files, so we might have to create intermediate directories here
			if err := uncompressFile(ctx, tarReader, header.FileInfo(), targetDirRoot, path, showProgress); err != nil {
				return nil, err
			}
			// return full paths
			extractedFiles = append(extractedFiles, filepath.Join(targetDirRoot.Name(), path))
		}
	}
}

func uncompressFile(ctx context.Context, tarReader io.Reader, fileInfo os.FileInfo, rootDir *os.Root, path string, showProgress bool) error {
	// with a file filter, we may have skipped the intermediate directories, make sure they exist
	if err := rootDir.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}

	file, err := rootDir.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fileInfo.Mode().Perm())
	if err != nil {
		return err
	}
	defer file.Close()

	reader, cleanup := progressBarReader(tarReader, fileInfo, showProgress)
	defer cleanup()

	_, err = crcos.CopySparse(ctx, file, reader)
	if err != nil {
		return err
	}
	return file.Close()
}

// BuildPathChecked joins filename to baseDir and checks for ZipSlip and path traversal vulnerabilities.
// More info: https://snyk.io/research/zip-slip-vulnerability, https://learn.snyk.io/lesson/directory-traversal/?ecosystem=golang
func BuildPathChecked(baseDir, filename string) (string, error) {
	path := filepath.Join(baseDir, filename) // #nosec G305
	baseDir = filepath.Clean(baseDir)
	// clean prefix to ensure it isn't "//" when baseDir is "/"
	prefix := filepath.Clean(baseDir + string(os.PathSeparator))
	if path != baseDir && !strings.HasPrefix(path, prefix) {
		return "", fmt.Errorf("%s: illegal file path (expected prefix: %s)", path, prefix)
	}
	return path, nil
}

func unzip(ctx context.Context, archive, target string, fileFilter func(string) bool, showProgress bool) ([]string, error) {
	var extractedFiles []string
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	if err := os.MkdirAll(target, 0o750); err != nil {
		return nil, err
	}
	targetDirRoot, err := os.OpenRoot(target)
	if err != nil {
		return nil, err
	}
	defer targetDirRoot.Close()

	for _, file := range reader.File {
		// path must not be absolute, so it is not joined to targetDir
		// and the filename is just cleaned
		path := filepath.Clean(file.Name)

		if fileFilter != nil && !fileFilter(path) {
			logging.Debugf("untar: Skipping %s", path)
			continue
		}
		if file.FileInfo().IsDir() {
			if err = targetDirRoot.MkdirAll(path, file.Mode().Perm()); err != nil {
				return nil, err
			}
			continue
		}

		if err := unzipFile(ctx, file, targetDirRoot, path, showProgress); err != nil {
			return nil, err
		}
		// return full paths
		extractedFiles = append(extractedFiles, filepath.Join(targetDirRoot.Name(), path))
	}

	return extractedFiles, nil
}

func unzipFile(ctx context.Context, file *zip.File, rootDir *os.Root, path string, showProgress bool) error {
	fileReader, err := file.Open()
	if err != nil {
		return err
	}
	defer fileReader.Close()

	return uncompressFile(ctx, fileReader, file.FileInfo(), rootDir, path, showProgress)
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
