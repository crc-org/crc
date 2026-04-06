package machine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/cluster"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	"github.com/crc-org/crc/v2/pkg/crc/oc"
	"github.com/crc-org/crc/v2/pkg/crc/preset"
	crcos "github.com/crc-org/crc/v2/pkg/os"
)

const registryHost = "default-route-openshift-image-registry" + constants.AppsDomain

func (client *client) ImageLoad(cfg types.ImageLoadConfig) error {
	if client.GetPreset() != preset.OpenShift && client.GetPreset() != preset.OKD {
		return fmt.Errorf("image load is only supported for OpenShift clusters")
	}

	ocConfig := oc.UseOCWithConfig(client.name)

	if err := checkRegistryRunning(ocConfig); err != nil {
		return err
	}

	if err := ensureDefaultRoute(ocConfig); err != nil {
		return err
	}

	runner := crcos.NewLocalCommandRunner()

	if err := registryLogin(ocConfig, runner); err != nil {
		return err
	}

	dest := fmt.Sprintf("%s/%s/%s", registryHost, cfg.Namespace, cfg.ImageName)

	if cfg.IsTar {
		return loadFromTar(ocConfig, runner, cfg.Source, dest)
	}
	return loadFromRef(ocConfig, runner, cfg.Source, dest)
}

func checkRegistryRunning(ocConfig oc.Config) error {
	stdout, stderr, err := ocConfig.RunOcCommand(
		"get", "pods",
		"-n", "openshift-image-registry",
		"-l", "docker-registry=default",
		"-o", "jsonpath={.items[0].status.phase}",
	)
	if err != nil {
		return fmt.Errorf("failed to check image registry status: %s: %w", stderr, err)
	}
	if strings.TrimSpace(stdout) != "Running" {
		return fmt.Errorf("openShift image registry is not running (status: %s)", strings.TrimSpace(stdout))
	}
	return nil
}

func ensureDefaultRoute(ocConfig oc.Config) error {
	stdout, _, err := ocConfig.RunOcCommand(
		"get", "configs.imageregistry.operator.openshift.io/cluster",
		"-o", "jsonpath={.spec.defaultRoute}",
	)
	if err != nil {
		return fmt.Errorf("failed to check registry default route: %w", err)
	}

	if strings.TrimSpace(stdout) == "true" {
		return nil
	}

	logging.Info("Enabling default route for the image registry...")
	_, stderr, err := ocConfig.RunOcCommand(
		"patch", "configs.imageregistry.operator.openshift.io/cluster",
		"--type=merge",
		"-p", `{"spec":{"defaultRoute":true}}`,
	)
	if err != nil {
		return fmt.Errorf("failed to enable registry default route: %s: %w", stderr, err)
	}

	logging.Info("Waiting for registry route to become available...")
	_, stderr, err = ocConfig.RunOcCommand(
		"wait", "--for=condition=Available",
		"configs.imageregistry.operator.openshift.io/cluster",
		"--timeout=120s",
	)
	if err != nil {
		return fmt.Errorf("timed out waiting for registry route: %s: %w", stderr, err)
	}

	return nil
}

// registryLogin obtains an OAuth token by logging in as kubeadmin, then uses
// that token to authenticate with the internal registry via podman login.
func registryLogin(ocConfig oc.Config, runner crcos.CommandRunner) error {
	password, err := cluster.GetUserPassword(constants.GetKubeAdminPasswordPath())
	if err != nil {
		return fmt.Errorf("failed to read kubeadmin password: %w", err)
	}

	// Login as kubeadmin using the CRC kubeconfig to get an OAuth token
	_, stderr, err := runner.Run(
		ocConfig.OcExecutablePath,
		"login", "-u", "kubeadmin", "-p", password,
		fmt.Sprintf("https://api%s:6443", constants.ClusterDomain),
		"--insecure-skip-tls-verify=true",
		"--kubeconfig="+ocConfig.KubeconfigPath,
	)
	if err != nil {
		return fmt.Errorf("failed to login as kubeadmin: %s: %w", stderr, err)
	}

	// Get the OAuth token
	token, stderr, err := runner.RunPrivate(
		ocConfig.OcExecutablePath, "whoami", "-t",
		"--kubeconfig="+ocConfig.KubeconfigPath,
	)
	if err != nil {
		return fmt.Errorf("failed to get OAuth token: %s: %w", stderr, err)
	}
	token = strings.TrimSpace(token)

	// Login to registry with the OAuth token
	_, stderr, err = runner.RunPrivate(
		"podman", "login",
		registryHost,
		"-u", "kubeadmin",
		"-p", token,
		"--tls-verify=false",
	)
	if err != nil {
		return fmt.Errorf("failed to login to registry: %s: %w", stderr, err)
	}

	return nil
}

func loadFromTar(ocConfig oc.Config, runner crcos.CommandRunner, tarPath, dest string) error {
	absPath, err := filepath.Abs(tarPath)
	if err != nil {
		return fmt.Errorf("failed to resolve tar path: %w", err)
	}

	logging.Infof("Loading image from %s ...", absPath)

	stdout, stderr, err := runner.Run("podman", "load", "-i", absPath)
	if err != nil {
		return fmt.Errorf("failed to load tar into local podman: %s: %w", stderr, err)
	}

	loadedRef := parseLoadedImageRef(stdout)
	if loadedRef == "" {
		return fmt.Errorf("could not determine image reference from podman load output: %s", stdout)
	}
	logging.Infof("Loaded image %s, mirroring to %s ...", loadedRef, dest)

	return mirrorAndReport(ocConfig, runner, loadedRef, dest)
}

func loadFromRef(ocConfig oc.Config, runner crcos.CommandRunner, source, dest string) error {
	logging.Infof("Mirroring image %s to %s ...", source, dest)
	return mirrorAndReport(ocConfig, runner, source, dest)
}

func mirrorAndReport(ocConfig oc.Config, runner crcos.CommandRunner, source, dest string) error {
	_, stderr, err := runner.Run(
		ocConfig.OcExecutablePath,
		"image", "mirror",
		"--insecure=true",
		"--filter-by-os=.*",
		fmt.Sprintf("%s=%s", source, dest),
	)
	if err != nil {
		return fmt.Errorf("failed to mirror image: %s: %w", stderr, err)
	}

	fmt.Fprintf(os.Stdout, "Image loaded successfully to image-registry.openshift-image-registry.svc:5000/%s\n", imageDest(dest))
	return nil
}

func parseLoadedImageRef(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Loaded image") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

func imageDest(dest string) string {
	return strings.TrimPrefix(dest, registryHost+"/")
}
