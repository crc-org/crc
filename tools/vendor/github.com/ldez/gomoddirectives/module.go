package gomoddirectives

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ldez/grignotin/gomod"
	"golang.org/x/mod/modfile"
)

// GetModuleFile gets module file.
// It's better to use [GetGoModFile] instead of this function.
func GetModuleFile() (*modfile.File, error) {
	info, err := gomod.GetModuleInfo()
	if err != nil {
		return nil, err
	}

	if info[0].GoMod == "" {
		return nil, errors.New("working directory is not part of a module")
	}

	return parseGoMod(info[0].GoMod)
}

func parseGoMod(goMod string) (*modfile.File, error) {
	raw, err := os.ReadFile(filepath.Clean(goMod))
	if err != nil {
		return nil, fmt.Errorf("reading go.mod file: %w", err)
	}

	return modfile.Parse("go.mod", raw, nil)
}
