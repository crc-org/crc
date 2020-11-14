// +build linux

package linux

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/code-ready/crc/pkg/crc/logging"
)

const releaseFile = "/etc/os-release"

// Taken from https://github.com/docker/machine/blob/master/libmachine/provision/os_release.go
// The /etc/os-release file contains operating system identification data
// See http://www.freedesktop.org/software/systemd/man/os-release.html for more details

// OsRelease reflects values in /etc/os-release
// Values in this struct must always be string
// or the reflection will not work properly.
type OsRelease struct {
	AnsiColor    string `osr:"ANSI_COLOR"`
	Name         string `osr:"NAME"`
	Version      string `osr:"VERSION"`
	Variant      string `osr:"VARIANT"`
	VariantID    string `osr:"VARIANT_ID"`
	ID           OsType `osr:"ID"`
	IDLike       string `osr:"ID_LIKE"`
	PrettyName   string `osr:"PRETTY_NAME"`
	VersionID    string `osr:"VERSION_ID"`
	HomeURL      string `osr:"HOME_URL"`
	SupportURL   string `osr:"SUPPORT_URL"`
	BugReportURL string `osr:"BUG_REPORT_URL"`
}

// Enum the Distro ID
type OsType string

const (
	RHEL   OsType = "rhel"
	Fedora OsType = "fedora"
	CentOS OsType = "centos"
	Ubuntu OsType = "ubuntu"
)

func stripQuotes(val string) string {
	if len(val) > 0 && val[0] == '"' {
		return val[1 : len(val)-1]
	}
	return val
}

func (osr *OsRelease) setIfPossible(key, val string) error {
	v := reflect.ValueOf(osr).Elem()
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := v.Type().Field(i)
		originalName := fieldType.Tag.Get("osr")
		if key == originalName && fieldValue.Kind() == reflect.String {
			fieldValue.SetString(val)
			return nil
		}
	}
	return fmt.Errorf("Couldn't set key %s, no corresponding struct field found", key)
}

func parseLine(osrLine string) (string, string, error) {
	if osrLine == "" {
		return "", "", nil
	}

	vals := strings.Split(osrLine, "=")
	if len(vals) != 2 {
		return "", "", fmt.Errorf("Expected %s to split by '=' char into two strings, instead got %d strings", osrLine, len(vals))
	}
	key := vals[0]
	val := stripQuotes(vals[1])
	return key, val, nil
}

func UnmarshalOsRelease(osReleaseContents []byte, release *OsRelease) error {
	r := bytes.NewReader(osReleaseContents)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		key, val, err := parseLine(scanner.Text())
		if err != nil {
			logging.Warnf("Warning: got an invalid line error parsing /etc/os-release: %s", err)
			continue
		}
		if err := release.setIfPossible(key, val); err != nil {
			logging.Debug(err)
		}
	}
	return nil
}

func GetOsRelease() (*OsRelease, error) {
	// Check if release file exist
	if _, err := os.Stat(releaseFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s not exist", releaseFile)
	}
	content, err := ioutil.ReadFile(releaseFile)
	if err != nil {
		return nil, err
	}
	var osr OsRelease
	if err := UnmarshalOsRelease(content, &osr); err != nil {
		return nil, err
	}
	return &osr, nil
}

func (osr *OsRelease) GetIDLike() []OsType {
	var idLike []OsType

	if osr == nil {
		return idLike
	}

	for _, id := range strings.Split(osr.IDLike, " ") {
		idLike = append(idLike, OsType(id))
	}

	return idLike
}
