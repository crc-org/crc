package main

import (
	_ "embed"
	"encoding/json"
	"testing"
)

//go:embed release-info.json
var releaseInfo []byte

func TestReleaseInfo(t *testing.T) {
	if !json.Valid(releaseInfo) {
		t.Fatal("release-info.json is not a valid json file")
	}
}
