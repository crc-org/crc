package state

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/code-ready/crc/pkg/crc/errors"
)

var GlobalState *GlobalStateType

type GlobalStateType struct {
	FilePath string `json:"-"`

	DnsPID int
}

// Create new object with data if file exists or
// Create json file and return object if doesn't exists
func NewGlobalState(path string) (*GlobalStateType, error) {
	state := &GlobalStateType{}
	state.FilePath = path

	// Check json file existence
	_, err := os.Stat(state.FilePath)
	if os.IsNotExist(err) {
		if errWrite := state.Write(); errWrite != nil {
			return nil, errWrite
		}
	} else {
		if errRead := state.read(); errRead != nil {
			return nil, errRead
		}
	}

	return state, nil
}

func (state *GlobalStateType) Write() error {
	jsonData, err := json.MarshalIndent(state, "", "\t")
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(state.FilePath, jsonData, 0644); err != nil {
		return err
	}

	return nil
}

func (state *GlobalStateType) Delete() error {
	if err := os.Remove(state.FilePath); err != nil {
		return err
	}

	return nil
}

func (cfg *GlobalStateType) read() error {
	raw, err := ioutil.ReadFile(cfg.FilePath)
	if err != nil {
		return err
	}

	if !json.Valid(raw) {
		return errors.New("Invalid JSON")
	}

	err = json.Unmarshal(raw, &cfg)
	if err != nil {
		return err
	}
	return nil
}
