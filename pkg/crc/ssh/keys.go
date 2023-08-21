package ssh

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	gossh "golang.org/x/crypto/ssh"
)

var (
	ErrKeyGeneration     = errors.New("Unable to generate key")
	ErrPrivateKey        = errors.New("Unable to marshal private key")
	ErrPublicKey         = errors.New("Unable to convert public key")
	ErrUnableToWriteFile = errors.New("Unable to write file")
)

type KeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
}

// NewKeyPair generates a new SSH keypair
// This will return a private & public key encoded as DER.
func NewKeyPair() (keyPair *KeyPair, err error) {

	priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, ErrKeyGeneration
	}

	privDer, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, ErrPrivateKey
	}

	pubSSH, err := gossh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, ErrPublicKey
	}

	return &KeyPair{
		PrivateKey: privDer,
		PublicKey:  gossh.MarshalAuthorizedKey(pubSSH),
	}, nil
}

// GenerateSSHKey generates SSH keypair based on path of the private key
// The public key would be generated to the same path with ".pub" added
func GenerateSSHKey(path string) error {
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Desired directory for SSH keys does not exist: %s", err)
		}

		kp, err := NewKeyPair()
		if err != nil {
			return fmt.Errorf("Error generating key pair: %s", err)
		}

		if err := kp.WriteToFile(path, fmt.Sprintf("%s.pub", path)); err != nil {
			return fmt.Errorf("Error writing keys to file(s): %s", err)
		}
	}

	return nil
}

func RemoveCRCHostEntriesFromKnownHosts() error {
	knownHostsPath := filepath.Join(constants.GetHomeDir(), ".ssh", "known_hosts")
	if _, err := os.Stat(knownHostsPath); err != nil {
		return nil
	}
	f, err := os.Open(knownHostsPath)
	if err != nil {
		return fmt.Errorf("Unable to open user's 'known_hosts' file: %w", err)
	}
	defer f.Close()

	tempHostsFile, err := os.CreateTemp(filepath.Join(constants.GetHomeDir(), ".ssh"), "crc")
	if err != nil {
		return fmt.Errorf("Unable to create temp file: %w", err)
	}
	defer func() {
		tempHostsFile.Close()
		os.Remove(tempHostsFile.Name())
	}()

	if err := tempHostsFile.Chmod(0600); err != nil {
		return fmt.Errorf("Error trying to change permissions for temp file: %w", err)
	}

	// return each line along with the newline '\n' marker
	var splitFunc = func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			return i + 1, data[0 : i+1], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	}

	var foundCRCEntries bool
	scanner := bufio.NewScanner(f)
	scanner.Split(splitFunc)
	writer := bufio.NewWriter(tempHostsFile)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "[127.0.0.1]:2222") || strings.Contains(scanner.Text(), "192.168.130.11") {
			foundCRCEntries = true
			continue
		}
		if _, err := writer.WriteString(scanner.Text()); err != nil {
			return fmt.Errorf("Error while writing hostsfile content to temp file: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Error while reading content from known_hosts file: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("Error while flushing buffered content to temp file: %w", err)
	}

	if foundCRCEntries {
		if err := f.Close(); err != nil {
			return fmt.Errorf("Error closing known_hosts file: %w", err)
		}
		if err := tempHostsFile.Close(); err != nil {
			return fmt.Errorf("Error closing temp file: %w", err)
		}
		return os.Rename(tempHostsFile.Name(), knownHostsPath)
	}
	return nil
}
