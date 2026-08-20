package hostsapi

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// EnvVar is the environment variable routes-controller should send as a Bearer token.
	EnvVar = "CRC_HOSTS_API_TOKEN"
	// SecretName is the Kubernetes Secret holding the hosts API token.
	SecretName = "crc-hosts-api-token" // nolint:gosec
	// SecretKey is the key inside SecretName.
	SecretKey = "CRC_HOSTS_API_TOKEN" // nolint:gosec
	// SecretNamespace is where routes-controller runs.
	SecretNamespace = "openshift-ingress"

	bearerPrefix = "Bearer "
	tokenBytes   = 32
)

// LoadOrCreateToken returns the hosts API token from path, creating it if missing.
func LoadOrCreateToken(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		token := strings.TrimSpace(string(data))
		if token == "" {
			return "", fmt.Errorf("hosts API token file %q is empty", path)
		}
		return token, nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("reading hosts API token: %w", err)
	}

	token, err := generateToken()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", fmt.Errorf("creating hosts API token directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(token+"\n"), 0o600); err != nil {
		return "", fmt.Errorf("writing hosts API token: %w", err)
	}
	return token, nil
}

// BearerTokenAuthorized reports whether Authorization header carries the expected Bearer token.
func BearerTokenAuthorized(authorizationHeader, expectedToken string) bool {
	if expectedToken == "" {
		return false
	}
	if !strings.HasPrefix(authorizationHeader, bearerPrefix) {
		return false
	}
	got := strings.TrimPrefix(authorizationHeader, bearerPrefix)
	if len(got) != len(expectedToken) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(expectedToken)) == 1
}

func generateToken() (string, error) {
	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generating hosts API token: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
