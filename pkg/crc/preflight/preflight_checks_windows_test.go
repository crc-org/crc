package preflight

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserInMemberList(t *testing.T) {
	members := `DESKTOP-R05QDNL\someUser1
DESKTOP-R05QDNL\someUser2
NT AUTHORITY\INTERACTIVE
DESKTOP-G7H96M0\crc`

	tests := []struct {
		username string
		expected bool
	}{
		{`some`, false},
		{`DESKTOP-G7H96M0\someUser1`, false},
		{`DESK`, false},
		{`someUser2`, true},
		{`crc`, true},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, usernameInMembersList(tt.username, members))
	}
}
