package types

import (
	"net"
	"regexp"
)

type Configuration struct {
	Debug       bool
	CaptureFile string

	MTU int

	Subnet            string
	GatewayIP         string
	GatewayMacAddress string

	DNS []Zone

	Forwards map[string]string

	NAT map[string]string
}

type Zone struct {
	Name      string
	Records   []Record
	DefaultIP net.IP
}

type Record struct {
	Name   string
	IP     net.IP
	Regexp *regexp.Regexp
}
