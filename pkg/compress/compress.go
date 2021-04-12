package compress

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/klauspost/compress/zstd"
)

func Compress(src, dest string) (err error) {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	enc, err := zstd.NewWriter(out)
	if err != nil {
		return err
	}
	defer enc.Close()

	tarWriter := tar.NewWriter(enc)
	defer tarWriter.Close()

	basePath, _ := filepath.Split(src)

	// Just use top level directory as part of tarball
	// $ zstdcat crc_libvirt_4.7.1_custom.zstd  | tar t
	// crc_libvirt_4.7.1_custom
	// crc_libvirt_4.7.1_custom/test
	return filepath.Walk(src, func(file string, fi os.FileInfo, err1 error) error {
		if err1 != nil {
			return err1
		}
		logging.Debugf("Adding %s", file)
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}
		// must provide real name
		// (see https://golang.org/src/archive/tar/common.go?#L626)
		header.Name, err = filepath.Rel(basePath, filepath.ToSlash(file))
		if err != nil {
			return err
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			defer data.Close()
			if _, err := io.Copy(tarWriter, data); err != nil {
				return err
			}
		}
		return nil
	})
}
