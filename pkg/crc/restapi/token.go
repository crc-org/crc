package restapi

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// SecretName is the Kubernetes Secret holding the REST API token.
	SecretName = "crc-rest-api-token" // nolint:gosec
	// SecretKey is the key inside SecretName.
	SecretKey = "CRC_REST_API_TOKEN" // nolint:gosec
	// SecretNamespace is where routes-controller runs.
	SecretNamespace = "openshift-ingress"

	bearerPrefix = "Bearer "
	tokenBytes   = 32
)

// LoadOrCreateToken returns the REST API token from path, creating it if missing.
func LoadOrCreateToken(path string) (string, error) {
	token, err := LoadToken(path)
	if err == nil {
		return token, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	token, err = generateToken()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", fmt.Errorf("creating REST API token directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(token+"\n"), 0o600); err != nil {
		return "", fmt.Errorf("writing REST API token: %w", err)
	}
	return token, nil
}

// LoadToken reads and validates the REST API token at path.
func LoadToken(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading REST API token: %w", err)
	}
	token := strings.TrimSpace(string(data))
	if err := validateToken(token); err != nil {
		return "", fmt.Errorf("REST API token file %q %w", path, err)
	}
	return token, nil
}

func validateToken(token string) error {
	if token == "" {
		return fmt.Errorf("is empty")
	}
	if len(token) != tokenBytes*2 {
		return fmt.Errorf("has unexpected length %d (want %d)", len(token), tokenBytes*2)
	}
	if _, err := hex.DecodeString(token); err != nil {
		return fmt.Errorf("contains invalid content: %w", err)
	}
	return nil
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
	return subtle.ConstantTimeCompare([]byte(got), []byte(expectedToken)) == 1
}

func generateToken() (string, error) {
	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generating REST API token: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
