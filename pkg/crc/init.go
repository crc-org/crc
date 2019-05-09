package crc

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	log "github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/state"
)

func init() {
	var err error
	state.GlobalState, err = state.NewGlobalState(constants.GlobalStatePath)
	if err != nil {
		log.InfoF("Error loading global state: %s", err.Error())
	}
}
