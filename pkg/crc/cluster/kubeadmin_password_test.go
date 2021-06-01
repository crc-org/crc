package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareHtpasswdWithOneUsername(t *testing.T) {
	htpasswd, err := getHtpasswd(map[string]string{"username": "password1"})
	assert.NoError(t, err)

	ok, err := compareHtpasswd(htpasswd, map[string]string{"username": "password1"})
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = compareHtpasswd(htpasswd, map[string]string{"username": "password2"})
	assert.NoError(t, err)
	assert.False(t, ok)

	ok, err = compareHtpasswd(htpasswd, map[string]string{"other": "password1"})
	assert.NoError(t, err)
	assert.False(t, ok)

	ok, err = compareHtpasswd(htpasswd, map[string]string{"username1": "password1", "username2": "password2"})
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestCompareHtpasswdWithTwoUsernames(t *testing.T) {
	htpasswd, err := getHtpasswd(map[string]string{"username1": "password1", "username2": "password2"})
	assert.NoError(t, err)

	ok, err := compareHtpasswd(htpasswd, map[string]string{"username1": "password1", "username2": "password2"})
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = compareHtpasswd(htpasswd, map[string]string{"username1": "password1", "username2": "password3"})
	assert.NoError(t, err)
	assert.False(t, ok)
}
