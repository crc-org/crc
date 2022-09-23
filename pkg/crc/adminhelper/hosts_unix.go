//go:build !windows
// +build !windows

package adminhelper

import (
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/crc-org/admin-helper/pkg/types"
)

func execute(args ...string) error {
	_, _, err := crcos.RunWithDefaultLocale(BinPath, args...)
	return err
}

func instance() helper {
	return &cliHelper{}
}

type cliHelper struct {
}

func (c *cliHelper) Add(req *types.AddRequest) error {
	return execute(append([]string{"add", req.IP}, req.Hosts...)...)
}

func (c *cliHelper) Remove(req *types.RemoveRequest) error {
	return execute(append([]string{"rm"}, req.Hosts...)...)
}

func (c *cliHelper) Clean(req *types.CleanRequest) error {
	return execute(append([]string{"clean"}, req.Domains...)...)
}
