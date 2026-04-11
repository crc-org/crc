package image

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	"github.com/spf13/cobra"
)

func getLoadCmd(config *config.Config) *cobra.Command {
	var namespace string
	var name string

	loadCmd := &cobra.Command{
		Use:   "load SOURCE",
		Short: "Load a container image into the cluster's internal registry",
		Long: `Load a container image into the OpenShift cluster's internal registry.

SOURCE can be a tar file path or a container image reference:

  # From a tar file:
  crc image load myapp.tar
  crc image load myapp.tar --namespace myproject --name myapp:v1

  # From a local container image reference:
  crc image load myapp:latest
  crc image load docker.io/library/nginx:latest --namespace myproject`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runLoad(config, args[0], namespace, name)
		},
	}
	loadCmd.Flags().StringVarP(&namespace, "namespace", "n", "openshift", "Target namespace in the internal registry")
	loadCmd.Flags().StringVar(&name, "name", "", "Override the image name:tag in the registry (derived from source by default)")
	return loadCmd
}

func runLoad(config *config.Config, source, namespace, name string) error {
	isTar := isTarFile(source)

	if name == "" {
		name = deriveImageName(source, isTar)
		if name == "" {
			return fmt.Errorf("cannot derive image name from source %q, use --name to specify it", source)
		}
	}

	client := machine.NewClient(constants.DefaultName, logging.IsDebug(), config)
	return client.ImageLoad(types.ImageLoadConfig{
		Source:    source,
		IsTar:     isTar,
		Namespace: namespace,
		ImageName: name,
	})
}

func isTarFile(source string) bool {
	if _, err := os.Stat(source); err != nil {
		return false
	}
	return strings.HasSuffix(source, ".tar") || strings.HasSuffix(source, ".tar.gz") || strings.HasSuffix(source, ".tgz")
}

func deriveImageName(source string, isTar bool) string {
	if isTar {
		base := filepath.Base(source)
		base = strings.TrimSuffix(base, ".tar.gz")
		base = strings.TrimSuffix(base, ".tgz")
		base = strings.TrimSuffix(base, ".tar")
		if base == "" {
			return ""
		}
		return base + ":latest"
	}

	// For image references, extract the name:tag portion
	ref := source
	if idx := strings.LastIndex(ref, "/"); idx >= 0 {
		ref = ref[idx+1:]
	}
	if !strings.Contains(ref, ":") {
		ref += ":latest"
	}
	return ref
}
