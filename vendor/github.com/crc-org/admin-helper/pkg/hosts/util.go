package hosts

import (
	"github.com/areYouLazy/libhosty"
)

type HostAdder struct {
	hostsToAdd []string
}

func (adder *HostAdder) PrependHosts(hosts []string) {
	adder.hostsToAdd = append(hosts, adder.hostsToAdd...)
}

func (adder *HostAdder) AppendHosts(hosts ...string) {
	adder.hostsToAdd = append(adder.hostsToAdd, hosts...)
}

func (adder *HostAdder) AppendHost(host string) {
	adder.hostsToAdd = append(adder.hostsToAdd, host)
}

func (adder *HostAdder) Len() int {
	return len(adder.hostsToAdd)
}

func (adder *HostAdder) PopN(count int) []string {
	if adder.Len() == 0 {
		return []string{}
	}
	if adder.Len() < count {
		count = adder.Len()
	}
	hosts := adder.hostsToAdd[:count]
	adder.hostsToAdd = adder.hostsToAdd[count:]

	return hosts
}

func (adder *HostAdder) FillLine(line *libhosty.HostsFileLine) {
	if len(line.Hostnames) == maxHostsInLine {
		return
	}
	if len(line.Hostnames) > maxHostsInLine {
		adder.PrependHosts(line.Hostnames[maxHostsInLine:])
		line.Hostnames = line.Hostnames[0:maxHostsInLine]
		return
	}

	newHosts := adder.PopN(maxHostsInLine - len(line.Hostnames))
	line.Hostnames = append(line.Hostnames, newHosts...)
}
