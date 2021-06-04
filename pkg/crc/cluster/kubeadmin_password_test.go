package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareHtpasswdWithOneUsername(t *testing.T) {
	htpasswd, err := getHtpasswd(map[string]string{"username": "password1"}, []string{})
	assert.NoError(t, err)

	ok, _, err := compareHtpasswd(htpasswd, map[string]string{"username": "password1"})
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, _, err = compareHtpasswd(htpasswd, map[string]string{"username": "password2"})
	assert.NoError(t, err)
	assert.False(t, ok)

	ok, _, err = compareHtpasswd(htpasswd, map[string]string{"other": "password1"})
	assert.NoError(t, err)
	assert.False(t, ok)

	ok, _, err = compareHtpasswd(htpasswd, map[string]string{"username1": "password1", "username2": "password2"})
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestCompareHtpasswdWithTwoUsernames(t *testing.T) {
	htpasswd, err := getHtpasswd(map[string]string{"username1": "password1", "username2": "password2"}, []string{})
	assert.NoError(t, err)

	ok, _, err := compareHtpasswd(htpasswd, map[string]string{"username1": "password1", "username2": "password2"})
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, _, err = compareHtpasswd(htpasswd, map[string]string{"username1": "password1", "username2": "password3"})
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestCompareFalseWithCustomEntries(t *testing.T) {
	wanted := map[string]string{"username": "password"}

	htpasswd, err := getHtpasswd(map[string]string{"username": "wrong", "external1": "external1", "external2": "external2"}, []string{})
	assert.NoError(t, err)

	ok, externals, err := compareHtpasswd(htpasswd, wanted)
	assert.NoError(t, err)
	assert.False(t, ok)

	htpasswd, err = getHtpasswd(wanted, externals)
	assert.NoError(t, err)
	ok, _, err = compareHtpasswd(htpasswd, map[string]string{"username": "password", "external1": "external1", "external2": "external2"})
	assert.NoError(t, err)
	assert.True(t, ok)
}
