package cluster

import (
	"os"
	"path/filepath"
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

func TestGenerateUserPassword_WhenValidFileProvided_ThenWritePasswordToFile(t *testing.T) {
	// Given
	dir := t.TempDir()
	userPasswordPath := filepath.Join(dir, "test-user-password")
	// When
	err := GenerateUserPassword(userPasswordPath, "test-user")
	// Then
	assert.NoError(t, err)
	actualPasswordFileContents, err := os.ReadFile(userPasswordPath)
	assert.NoError(t, err)
	assert.Equal(t, 23, len(actualPasswordFileContents))
}

var testResolveUserPasswordArguments = map[string]struct {
	kubeAdminPasswordViaConfig string
	developerPasswordViaConfig string
	expectedKubeAdminPassword  string
	expectedDeveloperPassword  string
}{
	"When no password configured in config, then read kubeadmin and developer passwords from password files": {
		"", "", "kubeadmin-password-via-file", "developer-password-via-file",
	},
	"When developer password configured in config, then use developer password from config": {
		"", "developer-password-via-config", "kubeadmin-password-via-file", "developer-password-via-config",
	},
	"When kube admin password configured in config, then use kube admin password from config": {
		"kubeadmin-password-via-config", "", "kubeadmin-password-via-config", "developer-password-via-file",
	},
	"When kube admin and developer password configured in config, then use kube admin and developer passwords from config": {
		"kubeadmin-password-via-config", "developer-password-via-config", "kubeadmin-password-via-config", "developer-password-via-config",
	},
}

func TestResolveUserPassword_WhenNothingProvided_ThenUsePasswordFromFiles(t *testing.T) {
	for name, test := range testResolveUserPasswordArguments {
		t.Run(name, func(t *testing.T) {
			// Given
			dir := t.TempDir()
			kubeAdminPasswordPath := filepath.Join(dir, "kubeadmin-password")
			err := os.WriteFile(kubeAdminPasswordPath, []byte("kubeadmin-password-via-file"), 0600)
			assert.NoError(t, err)
			developerPasswordPath := filepath.Join(dir, "developer-password")
			err = os.WriteFile(developerPasswordPath, []byte("developer-password-via-file"), 0600)
			assert.NoError(t, err)

			// When
			credentials, err := resolveUserPasswords(test.kubeAdminPasswordViaConfig, test.developerPasswordViaConfig, kubeAdminPasswordPath, developerPasswordPath)

			// Then
			assert.NoError(t, err)
			assert.Equal(t, map[string]string{"developer": test.expectedDeveloperPassword, "kubeadmin": test.expectedKubeAdminPassword}, credentials)
		})
	}
}
