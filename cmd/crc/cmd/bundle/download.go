package bundle

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	crcConfig "github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/gpg"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine/bundle"
	crcPreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/download"
	"github.com/spf13/cobra"
)

func getDownloadCmd(config *crcConfig.Config) *cobra.Command {
	downloadCmd := &cobra.Command{
		Use:   "download [version] [architecture]",
		Short: "Download a specific CRC bundle",
		Long:  "Download a specific CRC bundle from the mirrors. If no version or architecture is specified, the bundle for the current CRC version will be downloaded.",
		RunE: func(cmd *cobra.Command, args []string) error {
			force, _ := cmd.Flags().GetBool("force")
			presetStr, _ := cmd.Flags().GetString("preset")

			var preset crcPreset.Preset
			if presetStr != "" {
				var err error
				preset, err = crcPreset.ParsePresetE(presetStr)
				if err != nil {
					return err
				}
			} else {
				preset = crcConfig.GetPreset(config)
			}

			return runDownload(args, preset, force)
		},
	}
	downloadCmd.Flags().BoolP("force", "f", false, "Overwrite existing bundle if present")
	downloadCmd.Flags().StringP("preset", "p", "", "Target preset (openshift, okd, microshift)")

	return downloadCmd
}

func runDownload(args []string, preset crcPreset.Preset, force bool) error {
	// Disk space check (simple check for ~10GB free)
	// This is a basic check, more robust checking would require syscall/windows specific implementations
	// We skip this for now to avoid adding heavy OS-specific deps, assuming user manages disk space or download fails naturally.

	// If no args, use default bundle path
	if len(args) == 0 {
		defaultBundlePath := constants.GetDefaultBundlePath(preset)
		if !force {
			if _, err := os.Stat(defaultBundlePath); err == nil {
				logging.Infof("Bundle %s already exists. Use --force to overwrite.", defaultBundlePath)
				return nil
			}
		}

		logging.Debugf("Source: %s", constants.GetDefaultBundleDownloadURL(preset))
		logging.Debugf("Destination: %s", defaultBundlePath)
		// For default bundle, we use the existing logic which handles verification internally
		_, err := bundle.Download(context.Background(), preset, defaultBundlePath, false)
		return err
	}

	// If args provided, we are constructing a URL
	version := args[0]

	// Check if version is partial (Major.Minor) and resolve it if necessary
	resolvedVersion, err := resolveOpenShiftVersion(preset, version)
	if err != nil {
		logging.Warnf("Could not resolve version %s: %v. Trying with original version string.", version, err)
	} else if resolvedVersion != version {
		logging.Debugf("Resolved version %s to %s", version, resolvedVersion)
		version = resolvedVersion
	}

	architecture := runtime.GOARCH
	if len(args) > 1 {
		architecture = args[1]
	}

	bundleName := constants.BundleName(preset, version, architecture)
	bundlePath := filepath.Join(constants.MachineCacheDir, bundleName)

	if !force {
		if _, err := os.Stat(bundlePath); err == nil {
			logging.Infof("Bundle %s already exists. Use --force to overwrite.", bundleName)
			return nil
		}
	}

	// Base URL for the directory containing the bundle and signature
	baseVersionURL := fmt.Sprintf("%s/%s/%s/", constants.DefaultMirrorURL, preset.String(), version)
	bundleURL := fmt.Sprintf("%s%s", baseVersionURL, bundleName)
	sigURL := fmt.Sprintf("%s%s", baseVersionURL, "sha256sum.txt.sig")

	logging.Infof("Downloading bundle: %s", bundleName)
	logging.Debugf("Source: %s", bundleURL)
	logging.Debugf("Destination: %s", constants.MachineCacheDir)

	// Implement verification logic
	logging.Infof("Verifying signature for %s...", version)
	sha256sum, err := getVerifiedHashForCustomVersion(sigURL, bundleName)
	if err != nil {
		// Fallback: try without .sig if .sig not found, maybe just sha256sum.txt?
		// For now, fail if signature verification fails as requested for "Safeguards"
		return fmt.Errorf("signature verification failed: %w", err)
	}

	sha256bytes, err := hex.DecodeString(sha256sum)
	if err != nil {
		return fmt.Errorf("failed to decode sha256sum: %w", err)
	}

	_, err = download.Download(context.Background(), bundleURL, bundlePath, 0664, sha256bytes)
	return err
}

func getVerifiedHashForCustomVersion(sigURL string, bundleName string) (string, error) {
	// Reuse existing verification logic from bundle package via a helper here
	// We essentially replicate getVerifiedHash but with our custom URL

	res, err := download.InMemory(sigURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch signature file: %w", err)
	}
	defer res.Close()

	signedHashes, err := io.ReadAll(res)
	if err != nil {
		return "", fmt.Errorf("failed to read signature file: %w", err)
	}

	verifiedHashes, err := gpg.GetVerifiedClearsignedMsgV3(constants.RedHatReleaseKey, string(signedHashes))
	if err != nil {
		return "", fmt.Errorf("invalid signature: %w", err)
	}

	lines := strings.Split(verifiedHashes, "\n")
	for _, line := range lines {
		if strings.HasSuffix(line, bundleName) {
			sha256sum := strings.TrimSuffix(line, "  "+bundleName)
			return sha256sum, nil
		}
	}
	return "", fmt.Errorf("hash for %s not found in signature file", bundleName)
}
