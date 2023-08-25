package image

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/directory"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/signature"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/gpg"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	crcpreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/extract"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type imageHandler struct {
	imageURI string
}

func ValidateURI(uri *url.URL) error {
	/* For now we are very restrictive on the docker:// URLs we accept, it
	 * has to contain the bundle version as the tag, and it has to use the
	 * same image names as the official registry (openshift-bundle,
	 * okd-bundle, podman-bundle). In future releases, we'll make this more
	 * flexible
	 */
	imageAndTag := strings.Split(path.Base(uri.Path), ":")
	if len(imageAndTag) != 2 {
		return fmt.Errorf("invalid %s registry URL, tag is required (such as docker://quay.io/crcont/openshift-bundle:4.11.0)", uri)
	}
	_, err := getPresetNameE(imageAndTag[0])
	return err
}

func (img *imageHandler) policyContext() (*signature.PolicyContext, error) {
	policy := &signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}}
	policyContext, err := signature.NewPolicyContext(policy)
	if err != nil {
		return nil, fmt.Errorf("error creating security context: %w", err)
	}

	return policyContext, nil
}

// copyImage pulls the image from the registry and puts it to destination path
func (img *imageHandler) copyImage(destPath string, reportWriter io.Writer) (*v1.Manifest, error) {
	// Source Image from docker transport
	srcImg := img.imageURI
	srcRef, err := docker.ParseReference(srcImg)
	if err != nil {
		return nil, fmt.Errorf("invalid source image name %s: %w", srcImg, err)
	}

	destRef, err := directory.Transport.ParseReference(destPath)
	if err != nil {
		return nil, fmt.Errorf("invalid destination name %s: %w", destPath, err)
	}

	policyContext, err := img.policyContext()
	if err != nil {
		return nil, err
	}

	manifestData, err := copy.Image(context.Background(), policyContext,
		destRef, srcRef, &copy.Options{
			ReportWriter: reportWriter,
		})
	if err != nil {
		return nil, err
	}
	manifest := &v1.Manifest{}
	if err := json.Unmarshal(manifestData, manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

func getLayerPath(m *v1.Manifest, index int, mediaType string) (string, error) {
	if len(m.Layers) < (index + 1) {
		return "", fmt.Errorf("image layers in manifest is less than %d", index+1)
	}
	if m.Layers[index].MediaType != mediaType {
		return "", fmt.Errorf("expected media type for layer %s, got %s", mediaType, m.Layers[index].MediaType)
	}

	return strings.TrimPrefix(m.Layers[index].Digest.String(), "sha256:"), nil
}

func getPresetNameE(imageName string) (crcpreset.Preset, error) {
	switch imageName {
	case "openshift-bundle":
		return crcpreset.OpenShift, nil
	case "okd-bundle":
		return crcpreset.OKD, nil
	case "podman-bundle":
		return crcpreset.Podman, nil
	case "microshift-bundle":
		return crcpreset.Microshift, nil
	default:
		return crcpreset.OpenShift, fmt.Errorf("invalid image name '%s' (Should be openshift-bundle, okd-bundle, podman-bundle or microshift-bundle)", imageName)
	}
}
func GetPresetName(imageName string) crcpreset.Preset {
	preset, _ := getPresetNameE(imageName)
	return preset
}

func PullBundle(imageURI string) (string, error) {
	imgHandler := imageHandler{
		imageURI: strings.TrimPrefix(imageURI, "docker:"),
	}
	destDir, err := os.MkdirTemp(constants.MachineCacheDir, "tmpBundleImage")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(destDir)
	imgManifest, err := imgHandler.copyImage(destDir, os.Stdout)
	if err != nil {
		return "", err
	}

	logging.Info("Extracting the image bundle layer...")
	imgLayer, err := getLayerPath(imgManifest, 0, "application/vnd.oci.image.layer.v1.tar+gzip")
	if err != nil {
		return "", err
	}
	fileList, err := extract.Uncompress(filepath.Join(destDir, imgLayer), constants.MachineCacheDir)
	if err != nil {
		return "", err
	}
	logging.Debugf("Bundle and sign path: %v", fileList)

	logging.Info("Verifying the bundle signature...")
	if len(fileList) != 2 {
		return "", fmt.Errorf("image layer contains more files than expected: %v", fileList)
	}
	bundleFilePath, sigFilePath := fileList[0], fileList[1]
	if !strings.HasSuffix(sigFilePath, ".crcbundle.sig") {
		sigFilePath, bundleFilePath = fileList[0], fileList[1]
	}

	if err := gpg.Verify(bundleFilePath, sigFilePath); err != nil {
		return "", err
	}

	return bundleFilePath, nil
}
