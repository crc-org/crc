package cache

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	for _, tt := range []struct {
		name     string
		output   string
		expected string
		wantErr  bool
	}{
		// "something version <version>" form
		{"gvproxy", "gvproxy version v0.8.7\n", "v0.8.7", false},
		{"macadam", "macadam version v0.2.0\n", "v0.2.0", false},
		{"no trailing newline", "macadam version v0.2.0", "v0.2.0", false},
		{"admin-helper", "crc-admin-helper version 0.9.0\n", "0.9.0", false},
		// "something: <version>" form
		{"vfkit", "vfkit version: 0.5.3\n", "0.5.3", false},
		{"oc", "Client Version: v4.18.0\nKubernetes Version: v1.31.0\n", "v4.18.0", false},
		// only the first line is considered
		{"multi-line last-token", "tool version 1.2.3\nsome extra output: 4.5.6\n", "1.2.3", false},
		// whitespace robustness
		{"multiple spaces", "tool version   v1.0.0\n", "v1.0.0", false},
		{"tabs", "tool:\tv1.0.0\n", "v1.0.0", false},
		// error cases
		{"empty output", "", "", true},
		{"colon only", ":\n", "", true},
		{"spaces only", "   \n", "", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVersion(tt.output, "fake-executable")
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got version %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("parseVersion(%q) = %q, expected %q", tt.output, got, tt.expected)
			}
		})
	}
}
