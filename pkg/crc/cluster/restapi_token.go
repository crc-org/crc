package cluster

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/crc-org/crc/v2/pkg/crc/oc"
	"github.com/crc-org/crc/v2/pkg/crc/restapi"
	crcssh "github.com/crc-org/crc/v2/pkg/crc/ssh"
)

// EnsureRestAPITokenSecret creates or updates the Secret that routes-controller
// mounts as CRC_REST_API_TOKEN.
func EnsureRestAPITokenSecret(ctx context.Context, sshRunner *crcssh.Runner, ocConfig oc.Config, token string) error {
	if token == "" {
		return fmt.Errorf("REST API token must not be empty")

	}

	tokenBase64 := base64.StdEncoding.EncodeToString([]byte(token))

	secret := fmt.Sprintf(restapi.TokenSecret, tokenBase64)
	if err := sshRunner.CopyDataPrivileged([]byte(secret), "/opt/crc/rest-api-token.yaml", 0o644); err != nil {
		return err
	}

	_, _, err := ocConfig.RunOcCommand("apply", "-f", "/opt/crc/rest-api-token.yaml")
	if err != nil {
		return err
	}
	return nil
}
