package virtualbox

import (
	"bufio"
	"github.com/code-ready/machine/libmachine/mcnutils"
	"math/rand"
	"os"

	"time"

	"github.com/code-ready/machine/libmachine/ssh"
)

// CRCDiskCopier describes the interactions with crc disk image.
type CRCDiskCopier interface {
	CopyDiskToMachineDir(storePath, machineName, isoURL string) error
}

func NewCRCDiskCopier() CRCDiskCopier {
	return &crcDiskUtilsCopier{}
}

type crcDiskUtilsCopier struct{}

func (u *crcDiskUtilsCopier) CopyDiskToMachineDir(storePath, machineName, isoURL string) error {
	return mcnutils.NewB2dUtils(storePath).CopyDiskToMachineDir(isoURL, machineName)
}

// SSHKeyGenerator describes the generation of ssh keys.
type SSHKeyGenerator interface {
	Generate(path string) error
}

func NewSSHKeyGenerator() SSHKeyGenerator {
	return &defaultSSHKeyGenerator{}
}

type defaultSSHKeyGenerator struct{}

func (g *defaultSSHKeyGenerator) Generate(path string) error {
	return ssh.GenerateSSHKey(path)
}

// LogsReader describes the reading of VBox.log
type LogsReader interface {
	Read(path string) ([]string, error)
}

func NewLogsReader() LogsReader {
	return &vBoxLogsReader{}
}

type vBoxLogsReader struct{}

func (c *vBoxLogsReader) Read(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return []string{}, err
	}

	defer file.Close()

	lines := []string{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, nil
}

// RandomInter returns random int values.
type RandomInter interface {
	RandomInt(n int) int
}

func NewRandomInter() RandomInter {
	src := rand.NewSource(time.Now().UnixNano())

	return &defaultRandomInter{
		rand: rand.New(src),
	}
}

type defaultRandomInter struct {
	rand *rand.Rand
}

func (d *defaultRandomInter) RandomInt(n int) int {
	return d.rand.Intn(n)
}

// Sleeper sleeps for given duration.
type Sleeper interface {
	Sleep(d time.Duration)
}

func NewSleeper() Sleeper {
	return &defaultSleeper{}
}

type defaultSleeper struct{}

func (s *defaultSleeper) Sleep(d time.Duration) {
	time.Sleep(d)
}
