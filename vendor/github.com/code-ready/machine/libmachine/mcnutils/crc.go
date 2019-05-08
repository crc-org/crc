package mcnutils

import (
	"fmt"
	"github.com/code-ready/machine/libmachine/log"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

const (
	// This is the disk file name which gets copied to machine directory.
	// crc.disk naming make sure it doesn't conflict with the machine name
	// which right now is `crc`
	defaultDiskFilename = "crc.disk"
)

func defaultTimeout(network, addr string) (net.Conn, error) {
	return net.Dial(network, addr)
}

func getClient() *http.Client {
	transport := http.Transport{
		DisableKeepAlives: true,
		Proxy:             http.ProxyFromEnvironment,
		Dial:              defaultTimeout,
	}

	return &http.Client{
		Transport: &transport,
	}
}

// releaseGetter is a client that gets release information of a product and downloads it.
type releaseGetter interface {
	// filename returns filename of the product.
	filename() string

	// download downloads a file from the given dlURL and saves it under dir.
	download(dir, file, dlURL string) error
}

// disk is an disk path.
type disk interface {
	// path returns the path of the Disk.
	path() string
	// exists reports whether the Disk exists.
	exists() bool
}

// crcDisk represents a CRC disk Image. It implements the Disk interface.
type crcDisk struct {
	// path of Disk ISO
	commonDiskPath string
}

func (b *crcDisk) path() string {
	if b == nil {
		return ""
	}
	return b.commonDiskPath
}

func (b *crcDisk) exists() bool {
	if b == nil {
		return false
	}

	_, err := os.Stat(b.commonDiskPath)
	return !os.IsNotExist(err)
}

func removeFileIfExists(name string) error {
	if _, err := os.Stat(name); err == nil {
		if err := os.Remove(name); err != nil {
			return fmt.Errorf("Error removing temporary download file: %s", err)
		}
	}
	return nil
}

// crcReleaseGetter implements the releaseGetter interface for getting the Disk image.
type crcReleaseGetter struct {
	diskFilename string
}

func (b *crcReleaseGetter) filename() string {
	if b == nil {
		return ""
	}
	return b.diskFilename
}

func (*crcReleaseGetter) download(dir, file, isoURL string) error {
	u, err := url.Parse(isoURL)

	var src io.ReadCloser
	if u.Scheme == "file" || u.Scheme == "" {
		s, err := os.Open(u.Path)
		if err != nil {
			return err
		}

		src = s
	} else {
		client := getClient()
		s, err := client.Get(isoURL)
		if err != nil {
			return err
		}

		src = &ReaderWithProgress{
			ReadCloser:     s.Body,
			out:            os.Stdout,
			expectedLength: s.ContentLength,
		}
	}

	defer src.Close()

	// Download to a temp file first then rename it to avoid partial download.
	f, err := ioutil.TempFile(dir, file+".tmp")
	if err != nil {
		return err
	}

	defer func() {
		if err := removeFileIfExists(f.Name()); err != nil {
			log.Warnf("Error removing file: %s", err)
		}
	}()

	if _, err := io.Copy(f, src); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	// Dest is the final path of the disk image
	dest := filepath.Join(dir, file)

	// Windows can't rename in place, so remove the old file before
	// renaming the temporary downloaded file.
	if err := removeFileIfExists(dest); err != nil {
		return err
	}

	return os.Rename(f.Name(), dest)
}

type B2dUtils struct {
	releaseGetter
	disk
	storePath    string
	imgCachePath string
}

func NewB2dUtils(storePath string) *B2dUtils {
	imgCachePath := filepath.Join(storePath, "cache")

	return &B2dUtils{
		releaseGetter: &crcReleaseGetter{diskFilename: defaultDiskFilename},
		disk: &crcDisk{
			commonDiskPath: filepath.Join(imgCachePath, defaultDiskFilename),
		},
		storePath:    storePath,
		imgCachePath: imgCachePath,
	}
}

// DownloadISO downloads boot2docker ISO image for the given tag and save it at dest.
func (b *B2dUtils) DownloadDisk(dir, file, isoURL string) error {
	log.Infof("Downloading %s from %s...", b.path(), isoURL)
	return b.download(dir, file, isoURL)
}

type ReaderWithProgress struct {
	io.ReadCloser
	out                io.Writer
	bytesTransferred   int64
	expectedLength     int64
	nextPercentToPrint int64
}

func (r *ReaderWithProgress) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)

	if n > 0 {
		r.bytesTransferred += int64(n)
		percentage := r.bytesTransferred * 100 / r.expectedLength

		for percentage >= r.nextPercentToPrint {
			if r.nextPercentToPrint%10 == 0 {
				fmt.Fprintf(r.out, "%d%%", r.nextPercentToPrint)
			} else if r.nextPercentToPrint%2 == 0 {
				fmt.Fprint(r.out, ".")
			}
			r.nextPercentToPrint += 2
		}
	}

	return n, err
}

func (r *ReaderWithProgress) Close() error {
	fmt.Fprintln(r.out)
	return r.ReadCloser.Close()
}

func (b *B2dUtils) CopyDiskToMachineDir(diskURL, machineName string) error {
	// TODO: This is a bit off-color.
	machineDir := filepath.Join(b.storePath, "machines", machineName)
	machineIsoPath := filepath.Join(machineDir, b.filename())

	// By default just copy the existing "cached" iso to the machine's directory...
	if diskURL == "" {
		log.Infof("Copying %s to %s...", b.path(), machineIsoPath)
		return CopyFile(b.path(), machineIsoPath)
	}

	return b.DownloadDisk(machineDir, b.filename(), diskURL)
}
