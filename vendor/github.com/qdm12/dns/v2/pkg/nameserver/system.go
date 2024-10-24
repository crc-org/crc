package nameserver

import (
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strings"

	"github.com/qdm12/gosettings"
)

type SettingsSystemDNS struct {
	// IP is the IP address to use for the DNS.
	// It defaults to 127.0.0.1 if nil.
	IP netip.Addr
	// ResolvPath is the path to the resolv configuration file.
	// It defaults to /etc/resolv.conf.
	ResolvPath string
}

func (s *SettingsSystemDNS) SetDefaults() {
	s.IP = gosettings.DefaultValidator(s.IP, netip.AddrFrom4([4]byte{127, 0, 0, 1}))
	s.ResolvPath = gosettings.DefaultComparable(s.ResolvPath, "/etc/resolv.conf")
}

var (
	ErrResolvPathIsDirectory = errors.New("resolv path is a directory")
)

func (s *SettingsSystemDNS) Validate() (err error) {
	stat, err := os.Stat(s.ResolvPath)
	switch {
	case errors.Is(err, os.ErrNotExist): // it will be created
	case err != nil:
		return fmt.Errorf("stating resolv path: %w", err)
	case stat.IsDir():
		return fmt.Errorf("%w: %s", ErrResolvPathIsDirectory, s.ResolvPath)
	}
	return nil
}

// UseDNSSystemWide changes the nameserver to use for DNS system wide.
// If resolvConfPath is empty, it defaults to /etc/resolv.conf.
func UseDNSSystemWide(settings SettingsSystemDNS) (err error) {
	settings.SetDefaults()

	stat, err := os.Stat(settings.ResolvPath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return createResolvFile(settings.ResolvPath, settings.IP)
	case err != nil:
		return fmt.Errorf("stating resolv path: %w", err)
	case stat.IsDir():
		return fmt.Errorf("%w: %s", ErrResolvPathIsDirectory, settings.ResolvPath)
	}

	return patchResolvFile(settings.ResolvPath, settings.IP)
}

func createResolvFile(resolvPath string, ip netip.Addr) (err error) {
	parentDirectory := filepath.Dir(resolvPath)
	const defaultPerms os.FileMode = 0o755
	err = os.MkdirAll(parentDirectory, defaultPerms)
	if err != nil {
		return fmt.Errorf("creating resolv path parent directory: %w", err)
	}

	const filePermissions os.FileMode = 0600
	data := []byte("nameserver " + ip.String() + "\n")
	err = os.WriteFile(resolvPath, data, filePermissions)
	if err != nil {
		return fmt.Errorf("creating resolv file: %w", err)
	}
	return nil
}

func patchResolvFile(resolvPath string, ip netip.Addr) (err error) {
	data, err := os.ReadFile(resolvPath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	patchedLines := make([]string, 0, len(lines)+1)
	patchedLines = append(patchedLines, "nameserver "+ip.String())
	for _, line := range lines {
		if !strings.HasPrefix(line, "nameserver ") {
			patchedLines = append(patchedLines, line)
		}
	}

	patchedString := strings.Join(patchedLines, "\n")
	patchedString = strings.TrimRight(patchedString, "\n")
	hadTrailNewLine := patchedLines[len(patchedLines)-1] == ""
	if hadTrailNewLine {
		patchedString += "\n"
	}

	patchedData := []byte(patchedString)
	const permissions os.FileMode = 0600
	err = os.WriteFile(resolvPath, patchedData, permissions)
	if err != nil {
		return fmt.Errorf("writing resolv file: %w", err)
	}
	return nil
}
