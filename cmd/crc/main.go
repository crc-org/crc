package main

import (
	"github.com/code-ready/crc/cmd/crc/cmd"
	"github.com/code-ready/crc/pkg/crc/machine/client"
)

func main() {
	client.StartDriver()
	cmd.Execute()
}
