package bundle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/code-ready/crc/pkg/compress"
	"github.com/code-ready/crc/pkg/crc/logging"
	crcos "github.com/code-ready/crc/pkg/os"
	godeepcopy "github.com/jinzhu/copier"
)

type Copier struct {
	srcBundle    *CrcBundleInfo
	copiedBundle CrcBundleInfo
}

func NewCopier(srcBundle *CrcBundleInfo, basePath string, customBundleName string) (*Copier, error) {
	var copier Copier

	bundlePath := filepath.Join(basePath, customBundleName)
	if err := os.Mkdir(bundlePath, 0775); err != nil {
		return nil, err
	}

	// Creating a deepCopy of source bundle metadata so that whatever we make
	// modification of any metadata field for copied operation should only
	// effect the copied bundle metadata instead of source bundle metadata.
	if err := godeepcopy.Copy(&copier.copiedBundle, srcBundle); err != nil {
		return nil, err
	}

	copier.copiedBundle.Name = customBundleName
	copier.copiedBundle.Type = "custom"
	copier.srcBundle = srcBundle
	copier.copiedBundle.cachedPath = bundlePath

	return &copier, nil
}

func (copier *Copier) Cleanup() error {
	return os.RemoveAll(copier.copiedBundle.cachedPath)
}

func (copier *Copier) resolvePath(filename string) string {
	return copier.copiedBundle.resolvePath(filename)
}

func (copier *Copier) CachedPath() string {
	return copier.copiedBundle.cachedPath
}

func (copier *Copier) CopyPrivateSSHKey(srcPath string) error {
	sshKeyFileName := filepath.Base(copier.srcBundle.GetSSHKeyPath())
	destPath := copier.resolvePath(sshKeyFileName)
	return crcos.CopyFileContents(srcPath, destPath, 0400)
}

func (copier *Copier) CopyKubeConfig() error {
	kubeConfigFileName := filepath.Base(copier.srcBundle.GetKubeConfigPath())
	srcPath := copier.srcBundle.GetKubeConfigPath()
	destPath := copier.resolvePath(kubeConfigFileName)
	return crcos.CopyFileContents(srcPath, destPath, 0640)
}

func (copier *Copier) CopyFilesFromFileList() error {
	for _, file := range copier.srcBundle.Storage.Files {
		srcPath := copier.srcBundle.resolvePath(file.Name)
		destPath := copier.resolvePath(filepath.Base(srcPath))
		if err := crcos.CopyFileContents(srcPath, destPath, 0755); err != nil {
			return err
		}
	}
	return nil
}

func (copier *Copier) SetDiskImage(path string, format string) error {
	fileInfo, err := getFileInfo(path)
	if err != nil {
		return err
	}
	diskImage := DiskImage{
		File:   *fileInfo,
		Format: format,
	}

	copier.copiedBundle.Storage.DiskImages = []DiskImage{diskImage}

	return nil
}

func (copier *Copier) GenerateBundle(bundleName string) error {
	if err := copier.copiedBundle.verify(); err != nil {
		return err
	}

	// update bundle info
	copier.copiedBundle.BuildInfo.BuildTime = time.Now().String()

	// Create the metadata json for custom bundle
	bundleContent, err := json.MarshalIndent(copier.copiedBundle, "", " ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(copier.resolvePath("crc-bundle-info.json"), bundleContent, 0600)
	if err != nil {
		return fmt.Errorf("error copying bundle metadata  %w", err)
	}

	logging.Infof("Compressing %s...", GetBundleNameWithoutExtension(copier.copiedBundle.Name))
	return compress.Compress(copier.copiedBundle.cachedPath, fmt.Sprintf("%s%s", bundleName, bundleExtension))
}

func sha256sum(path string) (string, error) {
	logging.Infof("Generating sha256sum for %s...", filepath.Base(path))
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func getFileInfo(path string) (*File, error) {
	fileInfo := File{}
	fileInfo.Name = filepath.Base(path)

	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	fileInfo.Size = fmt.Sprintf("%d", stat.Size())

	checksum, err := sha256sum(path)
	if err != nil {
		return nil, err
	}
	fileInfo.Checksum = checksum

	return &fileInfo, nil
}
