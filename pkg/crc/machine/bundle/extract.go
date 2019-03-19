package bundle

import (
	"io"
	"os"
	"path/filepath"

	"archive/tar"
	"compress/gzip"
)

func Extract(sourcepath string, destpath string) error {
	file, err := os.Open(sourcepath)

	if err != nil {
		return err
	}

	defer file.Close()

	var fileReader io.ReadCloser = file

	if fileReader, err = gzip.NewReader(file); err != nil {
		return err
	}
	defer fileReader.Close()

	tarBallReader := tar.NewReader(fileReader)

	// Extracting tarred files
	for {
		header, err := tarBallReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// get the individual filename and extract to the specified directory
		filename := filepath.Join(destpath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// handle directory
			// log directory (filename)
			err = os.MkdirAll(filename, os.FileMode(header.Mode))

			if err != nil {
				return err
			}

		case tar.TypeReg:
			// handle normal file
			// log file (filename)
			writer, err := os.Create(filename)

			if err != nil {
				return err
			}

			io.Copy(writer, tarBallReader)

			err = os.Chmod(filename, os.FileMode(header.Mode))

			if err != nil {
				return err
			}

			writer.Close()

		default:
			// ignore
		}

	}

	return nil
}
