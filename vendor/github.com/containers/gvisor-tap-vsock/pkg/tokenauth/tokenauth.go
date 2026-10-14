package tokenauth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

// GenerateToken generates a cryptographically secure random token
// suitable for use as an API bearer token. The token is 32 bytes of
// random data encoded as base64url (resulting in ~44 characters).
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func validateToken(token string) error {
	tokenLength := len(token)
	if tokenLength < 32 {
		return fmt.Errorf("token is too short (%d characters, minimum 32 required)", tokenLength)
	}
	if tokenLength > 256 {
		return fmt.Errorf("token is too long (%d characters, maximum 256 allowed)", tokenLength)
	}
	return nil
}

// readTokenFromFile reads the token from a file and validates file permissions and token format
func readTokenFromFile(filepath string) (string, error) {
	info, err := os.Stat(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to stat token file: %w", err)
	}

	// Validate file permissions (macOS and Linux only)
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if perm := info.Mode().Perm(); perm != 0o600 {
			return "", fmt.Errorf("token file has insecure permissions %o (must be 0600)", perm)
		}
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to read token file: %w", err)
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("token file is empty")
	}

	if err := validateToken(token); err != nil {
		return "", err
	}

	return token, nil
}

// readTokenFromEnv reads the token from the GV_API_TOKEN environment variable
// Returns empty string and no error if the environment variable is not set
// Returns error if the token is set but invalid (too short)
func readTokenFromEnv() (string, error) {
	token := strings.TrimSpace(os.Getenv("GV_API_TOKEN"))
	if token == "" {
		// Empty token means "no token configured"
		return "", nil
	}
	if err := validateToken(token); err != nil {
		return "", err
	}
	return token, nil
}

// ReadToken reads the API token from the specified file or environment variable
// If tokenFile is empty, only the environment variable is read
// Priority: file > environment variable > empty (no auth)
// Returns empty string (no error) if no token is configured
func ReadToken(tokenFile string) (string, error) {
	if tokenFile != "" {
		token, err := readTokenFromFile(tokenFile)
		if err != nil {
			return "", fmt.Errorf("failed to read API token from file %s: %w", tokenFile, err)
		}
		log.Infof("API token loaded from %s", tokenFile)
		return token, nil
	}

	token, err := readTokenFromEnv()
	if err != nil {
		return "", fmt.Errorf("invalid API token in GV_API_TOKEN environment variable: %w", err)
	}
	if token != "" {
		log.Info("API token loaded from GV_API_TOKEN environment variable")
	}

	return token, nil
}

// BearerAuthMiddleware creates middleware that validates Bearer token authentication
// If expectedToken is empty, the middleware allows all requests (backwards compatibility)
func BearerAuthMiddleware(expectedToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If no token is configured, allow all requests (backwards compatibility)
			if expectedToken == "" {
				log.Debug("No API token configured, allowing unauthenticated request")
				next.ServeHTTP(w, r)
				return
			}

			// Extract Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Warnf("API request from %s denied: missing Authorization header", r.RemoteAddr)
				http.Error(w, "Authorization required", http.StatusUnauthorized)
				return
			}

			// Check Bearer scheme (case-insensitive per RFC 7235)
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				log.Warnf("API request from %s denied: invalid Authorization scheme", r.RemoteAddr)
				http.Error(w, "Invalid authorization header format. Expected: Bearer <token>", http.StatusUnauthorized)
				return
			}

			providedToken := parts[1]

			// Use constant-time comparison to prevent timing attacks
			if subtle.ConstantTimeCompare([]byte(providedToken), []byte(expectedToken)) != 1 {
				log.Warnf("API request from %s denied: invalid token", r.RemoteAddr)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
