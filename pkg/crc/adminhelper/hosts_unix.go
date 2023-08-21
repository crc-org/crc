//go:build !windows
// +build !windows

package adminhelper

import (
	"github.com/crc-org/admin-helper/pkg/types"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	crcos "github.com/crc-org/crc/v2/pkg/os"
)

func execute(args ...string) error {
	_, _, err := crcos.RunWithDefaultLocale(constants.AdminHelperPath(), args...)
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
