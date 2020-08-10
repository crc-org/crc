package main

import (
	"fmt"
	"os"

	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/code-ready/crc/pkg/drivers/libvirt"
	"github.com/code-ready/machine/libmachine/drivers/plugin"
)

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "version" {
			fmt.Printf("Driver version: %s\n", version.GetCommitSha())
			os.Exit(0)
		}
	}
	plugin.RegisterDriver(libvirt.NewDriver("default", "path"))
}
