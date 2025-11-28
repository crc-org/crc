package cloudinit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateUserData(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	oldMachineInstanceDir := constants.MachineInstanceDir
	constants.MachineInstanceDir = tempDir
	defer func() {
		constants.MachineInstanceDir = oldMachineInstanceDir
	}()

	opts := UserDataOptions{
		PublicKey:         "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMockPublicKey test@example.com",
		PullSecret:        `{"auths":{"registry.redhat.io":{"auth":"mockauth"}}}`,
		KubeAdminPassword: "test-kubeadmin-pass",
		DeveloperPassword: "test-developer-pass",
	}

	userDataPath, err := GenerateUserData("test-machine", opts)
	require.NoError(t, err)

	// Verify the file was created
	assert.FileExists(t, userDataPath)

	// Read and verify the content
	content, err := os.ReadFile(userDataPath)
	require.NoError(t, err)

	contentStr := string(content)

	// Verify cloud-config header
	assert.Contains(t, contentStr, "#cloud-config")

	// Verify runcmd section
	assert.Contains(t, contentStr, "runcmd:")
	assert.Contains(t, contentStr, "systemctl enable --now kubelet")

	// Verify SSH key is present
	assert.Contains(t, contentStr, opts.PublicKey)
	assert.Contains(t, contentStr, "/home/core/.ssh/authorized_keys")

	// Verify pull secret
	assert.Contains(t, contentStr, opts.PullSecret)
	assert.Contains(t, contentStr, "/opt/crc/pull-secret")

	// Verify passwords
	assert.Contains(t, contentStr, opts.KubeAdminPassword)
	assert.Contains(t, contentStr, opts.DeveloperPassword)

	// Verify CRC environment variables
	assert.Contains(t, contentStr, "CRC_SELF_SUFFICIENT=1")
	assert.Contains(t, contentStr, "CRC_NETWORK_MODE_USER=1")

	// Verify file paths
	assert.Contains(t, contentStr, "/etc/sysconfig/crc-env")
	assert.Contains(t, contentStr, "/opt/crc/pass_kubeadmin")
	assert.Contains(t, contentStr, "/opt/crc/pass_developer")
}

func TestGetUserDataPath(t *testing.T) {
	tempDir := t.TempDir()
	oldMachineInstanceDir := constants.MachineInstanceDir
	constants.MachineInstanceDir = tempDir
	defer func() {
		constants.MachineInstanceDir = oldMachineInstanceDir
	}()

	path := getUserDataPath("test-machine")
	expectedPath := filepath.Join(tempDir, "test-machine", "user-data")
	assert.Equal(t, expectedPath, path)
}

func TestRemoveUserData(t *testing.T) {
	tempDir := t.TempDir()
	oldMachineInstanceDir := constants.MachineInstanceDir
	constants.MachineInstanceDir = tempDir
	defer func() {
		constants.MachineInstanceDir = oldMachineInstanceDir
	}()

	// Create a user-data file
	opts := UserDataOptions{
		PublicKey:         "test-key",
		PullSecret:        `{"auths":{"test.io":{"auth":"testauth"}}}`,
		KubeAdminPassword: "test-kubeadmin",
		DeveloperPassword: "test-developer",
	}

	userDataPath, err := GenerateUserData("test-machine", opts)
	require.NoError(t, err)
	assert.FileExists(t, userDataPath)

	// Remove the file
	err = RemoveUserData("test-machine")
	require.NoError(t, err)
	assert.NoFileExists(t, userDataPath)

	// Removing non-existent file should not error
	err = RemoveUserData("non-existent-machine")
	assert.NoError(t, err)
}

func TestGenerateUserDataMultilineContent(t *testing.T) {
	tempDir := t.TempDir()
	oldMachineInstanceDir := constants.MachineInstanceDir
	constants.MachineInstanceDir = tempDir
	defer func() {
		constants.MachineInstanceDir = oldMachineInstanceDir
	}()

	opts := UserDataOptions{
		PublicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMockPublicKey test@example.com",
		PullSecret: `{
  "auths": {
    "registry.redhat.io": {
      "auth": "mockauth"
    }
  }
}`,
		KubeAdminPassword: "test-kubeadmin-pass",
		DeveloperPassword: "test-developer-pass",
	}

	userDataPath, err := GenerateUserData("test-machine", opts)
	require.NoError(t, err)

	content, err := os.ReadFile(userDataPath)
	require.NoError(t, err)

	// Verify pull secret is compacted (no newlines or extra spaces)
	contentStr := string(content)
	assert.Contains(t, contentStr, "registry.redhat.io")
	// Should be compacted, not the original multiline format
	assert.Contains(t, contentStr, `{"auths":{"registry.redhat.io":{"auth":"mockauth"}}}`)
	// Should NOT contain the multiline version with spaces
	assert.NotContains(t, contentStr, "  \"auths\"")
}

func TestCompactJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name: "multiline JSON",
			input: `{
  "auths": {
    "registry.redhat.io": {
      "auth": "mockauth"
    }
  }
}`,
			expected: `{"auths":{"registry.redhat.io":{"auth":"mockauth"}}}`,
			wantErr:  false,
		},
		{
			name:     "already compact JSON",
			input:    `{"auths":{"registry.redhat.io":{"auth":"mockauth"}}}`,
			expected: `{"auths":{"registry.redhat.io":{"auth":"mockauth"}}}`,
			wantErr:  false,
		},
		{
			name:     "invalid JSON",
			input:    `{invalid json}`,
			expected: "",
			wantErr:  true,
		},
		{
			name:     "empty object",
			input:    `{}`,
			expected: `{}`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := compactJSON(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
