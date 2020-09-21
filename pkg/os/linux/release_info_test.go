// +build linux

package linux

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOsRelease(t *testing.T) {
	// These example osr files stolen shamelessly from
	// https://github.com/docker/docker/blob/master/pkg/parsers/operatingsystem/operatingsystem_test.go
	// cheers @tiborvass
	tcs := []struct {
		contents []byte
		expected OsRelease
	}{
		{
			contents: []byte(`NAME="Ubuntu"
VERSION="14.04, Trusty Tahr"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 14.04 LTS"
VERSION_ID="14.04"
HOME_URL="http://www.ubuntu.com/"
SUPPORT_URL="http://help.ubuntu.com/"
BUG_REPORT_URL="http://bugs.launchpad.net/ubuntu/"
`),
			expected: OsRelease{
				AnsiColor:    "",
				Name:         "Ubuntu",
				Version:      "14.04, Trusty Tahr",
				ID:           "ubuntu",
				IDLike:       "debian",
				PrettyName:   "Ubuntu 14.04 LTS",
				VersionID:    "14.04",
				HomeURL:      "http://www.ubuntu.com/",
				SupportURL:   "http://help.ubuntu.com/",
				BugReportURL: "http://bugs.launchpad.net/ubuntu/",
			},
		}, {
			contents: []byte(`NAME=Gentoo
ID=gentoo
PRETTY_NAME="Gentoo/Linux"
ANSI_COLOR="1;32"
HOME_URL="http://www.gentoo.org/"
SUPPORT_URL="http://www.gentoo.org/main/en/support.xml"
BUG_REPORT_URL="https://bugs.gentoo.org/"
`),
			expected: OsRelease{
				AnsiColor:    "1;32",
				Name:         "Gentoo",
				Version:      "",
				ID:           "gentoo",
				IDLike:       "",
				PrettyName:   "Gentoo/Linux",
				VersionID:    "",
				HomeURL:      "http://www.gentoo.org/",
				SupportURL:   "http://www.gentoo.org/main/en/support.xml",
				BugReportURL: "https://bugs.gentoo.org/",
			},
		}, {
			contents: []byte(`NAME="Ubuntu"
VERSION="14.04, Trusty Tahr"
ID=ubuntu
ID_LIKE=debian
VERSION_ID="14.04"
HOME_URL="http://www.ubuntu.com/"
SUPPORT_URL="http://help.ubuntu.com/"
BUG_REPORT_URL="http://bugs.launchpad.net/ubuntu/"
`),
			expected: OsRelease{
				AnsiColor:    "",
				Name:         "Ubuntu",
				Version:      "14.04, Trusty Tahr",
				ID:           "ubuntu",
				IDLike:       "debian",
				PrettyName:   "",
				VersionID:    "14.04",
				HomeURL:      "http://www.ubuntu.com/",
				SupportURL:   "http://help.ubuntu.com/",
				BugReportURL: "http://bugs.launchpad.net/ubuntu/",
			},
		}, {
			contents: []byte(`NAME="CentOS Linux"
VERSION="7 (Core)"
ID="centos"
ID_LIKE="rhel fedora"
VERSION_ID="7"
PRETTY_NAME="CentOS Linux 7 (Core)"
ANSI_COLOR="0;31"
HOME_URL="https://www.centos.org/"
BUG_REPORT_URL="https://bugs.centos.org/"
`),
			expected: OsRelease{
				Name:         "CentOS Linux",
				Version:      "7 (Core)",
				ID:           "centos",
				IDLike:       "rhel fedora",
				PrettyName:   "CentOS Linux 7 (Core)",
				AnsiColor:    "0;31",
				VersionID:    "7",
				HomeURL:      "https://www.centos.org/",
				BugReportURL: "https://bugs.centos.org/",
			},
		},
		{
			contents: []byte(`NAME=Fedora
VERSION="23 (Twenty Three)"
ID=fedora
VERSION_ID=23
VARIANT="Server Edition"
VARIANT_ID=server
PRETTY_NAME="Fedora 23 (Twenty Three)"
ANSI_COLOR="0;34"
HOME_URL="https://fedoraproject.org/"
BUG_REPORT_URL="https://bugzilla.redhat.com/"
`),
			expected: OsRelease{
				Name:         "Fedora",
				Version:      "23 (Twenty Three)",
				ID:           "fedora",
				PrettyName:   "Fedora 23 (Twenty Three)",
				Variant:      "Server Edition",
				VariantID:    "server",
				AnsiColor:    "0;34",
				VersionID:    "23",
				HomeURL:      "https://fedoraproject.org/",
				BugReportURL: "https://bugzilla.redhat.com/",
			},
		},
		{
			contents: []byte(`NAME="Red Hat Enterprise Linux"
VERSION="8.2 (Ootpa)"
ID="rhel"
ID_LIKE="fedora"
VERSION_ID="8.2"
PRETTY_NAME="Red Hat Enterprise Linux 8.2 (Ootpa)"
ANSI_COLOR="0;31"
HOME_URL="https://www.redhat.com/"
BUG_REPORT_URL="https://bugzilla.redhat.com/"
`),
			expected: OsRelease{
				Name:         "Red Hat Enterprise Linux",
				Version:      "8.2 (Ootpa)",
				ID:           "rhel",
				IDLike:       "fedora",
				PrettyName:   "Red Hat Enterprise Linux 8.2 (Ootpa)",
				AnsiColor:    "0;31",
				VersionID:    "8.2",
				HomeURL:      "https://www.redhat.com/",
				BugReportURL: "https://bugzilla.redhat.com/",
			},
		},
	}
	for _, tc := range tcs {
		var osr OsRelease
		assert.NoError(t, UnmarshalOsRelease(tc.contents, &osr))
		assert.Equal(t, tc.expected, osr)
	}
}

func TestParseLine(t *testing.T) {
	var (
		withQuotes    = "ID=\"ubuntu\""
		withoutQuotes = "ID=gentoo"
		wtf           = "LOTS=OF=EQUALS"
		blank         = ""
	)

	key, val, err := parseLine(withQuotes)
	assert.NoError(t, err)
	assert.Equal(t, "ID", key)
	assert.Equal(t, "ubuntu", val)

	key, val, err = parseLine(withoutQuotes)
	assert.NoError(t, err)
	assert.Equal(t, "ID", key)
	assert.Equal(t, "gentoo", val)

	_, _, err = parseLine(wtf)
	assert.EqualError(t, err, "Expected LOTS=OF=EQUALS to split by '=' char into two strings, instead got 3 strings")

	key, val, err = parseLine(blank)
	assert.NoError(t, err)
	assert.Equal(t, "", key)
	assert.Equal(t, "", val)
}
