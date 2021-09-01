package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommand(t *testing.T) {
	assert.Equal(t, "crc setup", Command{
		command: "setup",
	}.ToString())
	assert.Equal(t, "CRC_DEBUG_ENABLE_STOP_NTP=true crc start -b bundle", Command{
		command:     "start -b bundle",
		disableNTP:  true,
		updateCheck: true,
	}.ToString())
	assert.Equal(t, "CRC_DISABLE_UPDATE_CHECK=true CRC_DEBUG_ENABLE_STOP_NTP=true crc start -b bundle", Command{
		command:    "start -b bundle",
		disableNTP: true,
	}.ToString())
}
