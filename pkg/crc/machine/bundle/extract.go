package bundle

import (
	"archive/tar"
	"github.com/xi2/xz"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Extract(sourcepath string, destpath string) (string, error) {
	file, err := os.Open(sourcepath)

	if err != nil {
		return "", err
	}

	extractedPath := strings.Split(filepath.Base(file.Name()), ".tar")[0]
	extractedPath = filepath.Join(destpath, extractedPath)

	_, err = os.Stat(extractedPath)
	if err == nil {
		return extractedPath, nil
	}

	defer file.Close()

	var fileReader io.Reader = file

	if fileReader, err = xz.NewReader(file, 0); err != nil {
		return "", err
	}

	tarBallReader := tar.NewReader(fileReader)

	// Extracting tarred files
	for {
		header, err := tarBallReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		// get the individual filename and extract to the specified directory
		filename := filepath.Join(destpath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// handle directory
			// log directory (filename)
			err = os.MkdirAll(filename, os.FileMode(header.Mode))

			if err != nil {
				return "", err
			}

		case tar.TypeReg, tar.TypeGNUSparse:
			// handle normal file
			// log file (filename)
			writer, err := os.Create(filename)

			if err != nil {
				return "", err
			}

			io.Copy(writer, tarBallReader)

			err = os.Chmod(filename, os.FileMode(header.Mode))

			if err != nil {
				return "", err
			}

			writer.Close()

		default:
			// ignore
		}

	}

	return extractedPath, nil
}
