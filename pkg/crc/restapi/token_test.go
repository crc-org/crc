package restapi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadOrCreateToken_CreatesAndReuses(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rest-api.token")

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

func TestLoadToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rest-api.token")
	valid := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	require.NoError(t, os.WriteFile(path, []byte(valid+"\n"), 0o600))
	token, err := LoadToken(path)
	require.NoError(t, err)
	assert.Equal(t, valid, token)

	_, err = LoadToken(filepath.Join(dir, "missing"))
	assert.ErrorIs(t, err, os.ErrNotExist)

	require.NoError(t, os.WriteFile(path, []byte(""), 0o600))
	_, err = LoadToken(path)
	assert.ErrorContains(t, err, "is empty")

	require.NoError(t, os.WriteFile(path, []byte("abc"), 0o600))
	_, err = LoadToken(path)
	assert.ErrorContains(t, err, "has unexpected length")

	require.NoError(t, os.WriteFile(path, []byte(strings.Repeat("g", tokenBytes*2)), 0o600))
	_, err = LoadToken(path)
	assert.ErrorContains(t, err, "contains invalid content")
}

func TestBearerTokenAuthorized(t *testing.T) {
	assert.False(t, BearerTokenAuthorized("", "secret"))
	assert.False(t, BearerTokenAuthorized("Bearer ", "secret"))
	assert.False(t, BearerTokenAuthorized("Bearer wrong", "secret"))
	assert.False(t, BearerTokenAuthorized("secret", "secret"))
	assert.False(t, BearerTokenAuthorized("Bearer secret", ""))
	assert.True(t, BearerTokenAuthorized("Bearer secret", "secret"))
}
