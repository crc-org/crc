package hostsapi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadOrCreateToken_CreatesAndReuses(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts-api.token")

	token1, err := LoadOrCreateToken(path)
	require.NoError(t, err)
	assert.Len(t, token1, tokenBytes*2)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	token2, err := LoadOrCreateToken(path)
	require.NoError(t, err)
	assert.Equal(t, token1, token2)
}

func TestBearerTokenAuthorized(t *testing.T) {
	assert.False(t, BearerTokenAuthorized("", "secret"))
	assert.False(t, BearerTokenAuthorized("Bearer ", "secret"))
	assert.False(t, BearerTokenAuthorized("Bearer wrong", "secret"))
	assert.False(t, BearerTokenAuthorized("secret", "secret"))
	assert.False(t, BearerTokenAuthorized("Bearer secret", ""))
	assert.True(t, BearerTokenAuthorized("Bearer secret", "secret"))
}
