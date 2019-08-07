package bundle

import (
	"archive/tar"
	"github.com/xi2/xz"
	"io"
	"os"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
)

func Extract(sourcepath string) (*CrcBundleInfo, error) {
	file, err := os.Open(sourcepath)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	var fileReader io.Reader

	if fileReader, err = xz.NewReader(file, 0); err != nil {
		return nil, err
	}

	tarBallReader := tar.NewReader(fileReader)

	// Extracting tarred files
	for {
		header, err := tarBallReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		// get the individual filename and extract to the specified directory
		filename := filepath.Join(constants.MachineCacheDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// handle directory
			// log directory (filename)
			err = os.MkdirAll(filename, os.FileMode(header.Mode))

			if err != nil {
				return nil, err
			}

		case tar.TypeReg, tar.TypeGNUSparse:
			// handle normal file
			// log file (filename)
			writer, err := os.Create(filename)

			if err != nil {
				return nil, err
			}
			defer writer.Close()

			io.Copy(writer, tarBallReader)

			err = os.Chmod(filename, os.FileMode(header.Mode))

			if err != nil {
				return nil, err
			}

		default:
			// ignore
		}

	}

	return GetCachedBundleInfo(filepath.Base(sourcepath))
}
