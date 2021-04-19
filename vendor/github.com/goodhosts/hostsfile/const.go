// +build !windows

package hostsfile

const (
	HostsPerLine  = -1 // unlimited
	HostsFilePath = "/etc/hosts"
	eol           = "\n"
)
