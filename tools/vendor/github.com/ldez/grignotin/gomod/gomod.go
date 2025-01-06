// Package gomod A function to get information about module (go list).
package gomod

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// ModInfo Module information.
//
//nolint:tagliatelle // temporary: the next version of golangci-lint will allow configuration by package.
type ModInfo struct {
	Path      string `json:"Path"`
	Dir       string `json:"Dir"`
	GoMod     string `json:"GoMod"`
	GoVersion string `json:"GoVersion"`
	Main      bool   `json:"Main"`
}

// GetModuleInfo gets modules information from `go list`.
func GetModuleInfo() ([]ModInfo, error) {
	// https://github.com/golang/go/issues/44753#issuecomment-790089020
	cmd := exec.Command("go", "list", "-m", "-json")

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("command %q: %w: %s", strings.Join(cmd.Args, " "), err, string(out))
	}

	var infos []ModInfo

	for dec := json.NewDecoder(bytes.NewBuffer(out)); dec.More(); {
		var v ModInfo
		if err := dec.Decode(&v); err != nil {
			return nil, fmt.Errorf("unmarshaling error: %w: %s", err, string(out))
		}

		if v.GoMod == "" {
			return nil, errors.New("working directory is not part of a module")
		}

		if !v.Main || v.Dir == "" {
			continue
		}

		infos = append(infos, v)
	}

	if len(infos) == 0 {
		return nil, errors.New("go.mod file not found")
	}

	return infos, nil
}

type goEnv struct {
	GOMOD string `json:"GOMOD"` //nolint:tagliatelle // Based on en var name.
}

// GetGoModPath extracts go.mod path from "go env".
func GetGoModPath() (string, error) {
	cmd := exec.Command("go", "env", "-json", "GOMOD")

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("command %q: %w: %s", strings.Join(cmd.Args, " "), err, string(out))
	}

	v := &goEnv{}
	err = json.NewDecoder(bytes.NewBuffer(out)).Decode(v)
	if err != nil {
		return "", err
	}

	return v.GOMOD, nil
}

// GetModulePath extracts module path from go.mod.
func GetModulePath() (string, error) {
	p, err := GetGoModPath()
	if err != nil {
		return "", err
	}

	b, err := os.ReadFile(filepath.Clean(p))
	if err != nil {
		return "", fmt.Errorf("reading go.mod: %w", err)
	}

	return modfile.ModulePath(b), nil
}
