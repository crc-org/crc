package hostsfile

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"

	"github.com/dimchansky/utfbom"
)

type Hosts struct {
	Path  string
	Lines []HostsLine
}

// Return a new instance of ``Hosts`` using the default hosts file path.
func NewHosts() (Hosts, error) {
	osHostsFilePath := os.ExpandEnv(filepath.FromSlash(HostsFilePath))

	if env, isset := os.LookupEnv("HOSTS_PATH"); isset && len(env) > 0 {
		osHostsFilePath = os.ExpandEnv(filepath.FromSlash(env))
	}

	return NewCustomHosts(osHostsFilePath)
}

// Return a new instance of ``Hosts`` using a custom hosts file path.
func NewCustomHosts(osHostsFilePath string) (Hosts, error) {
	hosts := Hosts{
		Path: osHostsFilePath,
	}

	if err := hosts.Load(); err != nil {
		return hosts, err
	}

	return hosts, nil
}

// Return ```true``` if hosts file is writable.
func (h *Hosts) IsWritable() bool {
	_, err := os.OpenFile(h.Path, os.O_WRONLY, 0660)
	return err == nil
}

// Load the hosts file into ```l.Lines```.
// ```Load()``` is called by ```NewHosts()``` and ```Hosts.Flush()``` so you
// generally you won't need to call this yourself.
func (h *Hosts) Load() error {
	file, err := os.Open(h.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(utfbom.SkipOnly(file))
	for scanner.Scan() {
		h.Lines = append(h.Lines, NewHostsLine(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// Flush any changes made to hosts file.
func (h Hosts) Flush() error {
	file, err := os.Create(h.Path)
	if err != nil {
		return err
	}

	defer file.Close()

	w := bufio.NewWriter(file)

	for _, line := range h.Lines {
		if _, err := fmt.Fprintf(w, "%s%s", line.ToRaw(), eol); err != nil {
			return err
		}
	}

	err = w.Flush()
	if err != nil {
		return err
	}

	return h.Load()
}

// Add an entry to the hosts file.
func (h *Hosts) Add(ip string, hosts ...string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("%q is an invalid IP address", ip)
	}

	position := h.getIpPosition(ip)
	if position == -1 {
		// ip not already in hostsfile
		h.Lines = append(h.Lines, HostsLine{
			Raw:   buildRawLine(ip, hosts),
			IP:    ip,
			Hosts: hosts,
		})
	} else {
		// add new hosts to the correct position for the ip
		hostsCopy := h.Lines[position].Hosts
		for _, addHost := range hosts {
			if itemInSlice(addHost, hostsCopy) {
				continue // host exists for ip already
			}

			hostsCopy = append(hostsCopy, addHost)
		}
		h.Lines[position].Hosts = hostsCopy
		h.Lines[position].Raw = h.Lines[position].ToRaw() // reset raw
	}

	return nil
}

// merge dupelicate ips and hosts per ip
func (h *Hosts) Clean() {
	h.RemoveDuplicateIps()
	for pos, line := range h.Lines {
		line.RemoveDuplicateHosts()
		line.SortHosts()
		h.Lines[pos] = line
	}
	h.SortByIp()
	h.HostsPerLine(HostsPerLine)
}

// Return a bool if ip/host combo in hosts file.
func (h Hosts) Has(ip string, host string) bool {
	return h.getHostPosition(ip, host) != -1
}

// Return a bool if hostname in hosts file.
func (h Hosts) HasHostname(host string) bool {
	return h.getHostnamePosition(host) != -1
}

func (h Hosts) HasIp(ip string) bool {
	return h.getIpPosition(ip) != -1
}

// Remove an entry from the hosts file.
func (h *Hosts) Remove(ip string, hosts ...string) error {
	var outputLines []HostsLine
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("%q is an invalid IP address", ip)
	}

	for _, line := range h.Lines {
		// Bad lines or comments just get readded.
		if line.Err != nil || line.IsComment() || line.IP != ip {
			outputLines = append(outputLines, line)
			continue
		}

		var newHosts []string
		for _, checkHost := range line.Hosts {
			if !itemInSlice(checkHost, hosts) {
				newHosts = append(newHosts, checkHost)
			}
		}

		// If hosts is empty, skip the line completely.
		if len(newHosts) > 0 {
			newLineRaw := line.IP

			for _, host := range newHosts {
				newLineRaw = fmt.Sprintf("%s %s", newLineRaw, host)
			}
			newLine := NewHostsLine(newLineRaw)
			outputLines = append(outputLines, newLine)
		}
	}

	h.Lines = outputLines
	return nil
}

// Remove  entries by hostname from the hosts file.
func (h *Hosts) RemoveByHostname(host string) error {
	pos := h.getHostnamePosition(host)
	for pos > -1 {
		if len(h.Lines[pos].Hosts) > 1 {
			h.Lines[pos].Hosts = removeFromSlice(host, h.Lines[pos].Hosts)
			h.Lines[pos].RegenRaw()
		} else {
			h.removeByPosition(pos)
		}
		pos = h.getHostnamePosition(host)
	}

	return nil
}

func (h *Hosts) RemoveByIp(ip string) error {
	pos := h.getIpPosition(ip)
	for pos > -1 {
		h.removeByPosition(pos)
		pos = h.getIpPosition(ip)
	}

	return nil
}

func (h *Hosts) RemoveDuplicateIps() {
	ipCount := make(map[string]int)
	for _, line := range h.Lines {
		ipCount[line.IP]++
	}
	for ip, count := range ipCount {
		if count > 1 {
			h.combineIp(ip)
		}
	}
}

// convert to net.IP and byte.Compare
func (h *Hosts) SortByIp() {
	sortedIps := make([]net.IP, 0, len(h.Lines))
	for _, l := range h.Lines {
		sortedIps = append(sortedIps, net.ParseIP(l.IP))
	}
	sort.Slice(sortedIps, func(i, j int) bool {
		return bytes.Compare(sortedIps[i], sortedIps[j]) < 0
	})

	var sortedLines []HostsLine
	for _, ip := range sortedIps {
		for _, l := range h.Lines {
			if ip.String() == l.IP {
				sortedLines = append(sortedLines, l)
			}
		}
	}
	h.Lines = sortedLines
}

func (h *Hosts) HostsPerLine(count int) {
	if count <= 0 {
		return
	}
	var newLines []HostsLine
	for _, line := range h.Lines {
		if len(line.Hosts) <= count {
			newLines = append(newLines, line)
			continue
		}

		for i := 0; i < len(line.Hosts); i += count {
			lineCopy := line
			end := len(line.Hosts)
			if end > i+count {
				end = i + count
			}
			lineCopy.Hosts = line.Hosts[i:end]
			lineCopy.Raw = lineCopy.ToRaw()
			newLines = append(newLines, lineCopy)
		}
	}
	h.Lines = newLines
}

func (h *Hosts) combineIp(ip string) {
	newLine := HostsLine{
		IP: ip,
	}

	linesCopy := make([]HostsLine, len(h.Lines))
	copy(linesCopy, h.Lines)
	for _, line := range linesCopy {
		if line.IP == ip {
			newLine.Combine(line)
		}
	}
	newLine.SortHosts()
	h.removeIp(ip)
	h.Lines = append(h.Lines, newLine)
}

func (h *Hosts) removeByPosition(pos int) {
	if len(h.Lines) == 1 {
		h.Lines = []HostsLine{}
		return
	}
	if pos == len(h.Lines) {
		h.Lines = h.Lines[:pos-1]
		return
	}
	h.Lines = append(h.Lines[:pos], h.Lines[pos+1:]...)
}

func (h *Hosts) removeIp(ip string) {
	var newLines []HostsLine
	for _, line := range h.Lines {
		if line.IP != ip {
			newLines = append(newLines, line)
		}
	}

	h.Lines = newLines
}

func (h Hosts) getHostPosition(ip string, host string) int {
	for i := range h.Lines {
		line := h.Lines[i]
		if !line.IsComment() && line.Raw != "" {
			if ip == line.IP && itemInSlice(host, line.Hosts) {
				return i
			}
		}
	}

	return -1
}

func (h Hosts) getHostnamePosition(host string) int {
	for i := range h.Lines {
		line := h.Lines[i]
		if !line.IsComment() && line.Raw != "" {
			if itemInSlice(host, line.Hosts) {
				return i
			}
		}
	}

	return -1
}

func (h Hosts) getIpPosition(ip string) int {
	for i := range h.Lines {
		line := h.Lines[i]
		if !line.IsComment() && line.Raw != "" && line.IP == ip {
			return i
		}
	}

	return -1
}
