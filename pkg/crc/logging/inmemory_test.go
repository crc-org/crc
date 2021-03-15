package logging

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestAddLogs(t *testing.T) {
	memory := newInMemoryHook(5)

	assert.Len(t, memory.Messages(), 0)

	for i := 1; i < 10; i++ {
		assert.NoError(t, memory.Fire(&logrus.Entry{
			Message: fmt.Sprintf("message %d", i),
		}))
	}

	assert.Equal(t, []string{"message 5", "message 6", "message 7", "message 8", "message 9"}, memory.Messages())
}
