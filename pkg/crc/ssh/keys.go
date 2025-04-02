package ssh

import (
	"bufio"
	"bytes"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	gossh "golang.org/x/crypto/ssh"
)

var (
	ErrKeyGeneration     = errors.New("unable to generate key")
	ErrPrivateKey        = errors.New("unable to marshal private key")
	ErrPublicKey         = errors.New("unable to convert public key")
	ErrUnableToWriteFile = errors.New("unable to write file")
)

type KeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
}

// NewKeyPair generates a new SSH keypair
// This will return a private & public key encoded as DER.
func NewKeyPair() (keyPair *KeyPair, err error) {

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, ErrKeyGeneration
	}

	privMar, err := gossh.MarshalPrivateKey(crypto.PrivateKey(priv), "")
	if err != nil {
		return nil, ErrPrivateKey
	}

	pubSSH, err := gossh.NewPublicKey(pub)
	if err != nil {
		return nil, ErrPublicKey
	}

	return &KeyPair{
		PrivateKey: pem.EncodeToMemory(privMar),
		PublicKey:  gossh.MarshalAuthorizedKey(pubSSH),
	}, nil
}

// GenerateSSHKey generates SSH keypair based on path of the private key
// The public key would be generated to the same path with ".pub" added
func GenerateSSHKey(path string) error {
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("desired directory for SSH keys does not exist: %s", err)
		}

		kp, err := NewKeyPair()
		if err != nil {
			return fmt.Errorf("error generating key pair: %s", err)
		}

		if err := kp.WriteToFile(path, fmt.Sprintf("%s.pub", path)); err != nil {
			return fmt.Errorf("error writing keys to file(s): %s", err)
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
		return fmt.Errorf("unable to open user's 'known_hosts' file: %w", err)
	}
	defer f.Close()

	tempHostsFile, err := os.CreateTemp(filepath.Join(constants.GetHomeDir(), ".ssh"), "crc")
	if err != nil {
		return fmt.Errorf("unable to create temp file: %w", err)
	}
	defer func() {
		tempHostsFile.Close()
		os.Remove(tempHostsFile.Name())
	}()

	if err := tempHostsFile.Chmod(0600); err != nil {
		return fmt.Errorf("error trying to change permissions for temp file: %w", err)
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
			return fmt.Errorf("error while writing hostsfile content to temp file: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error while reading content from known_hosts file: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error while flushing buffered content to temp file: %w", err)
	}

	if foundCRCEntries {
		if err := f.Close(); err != nil {
			return fmt.Errorf("error closing known_hosts file: %w", err)
		}
		if err := tempHostsFile.Close(); err != nil {
			return fmt.Errorf("error closing temp file: %w", err)
		}
		return os.Rename(tempHostsFile.Name(), knownHostsPath)
	}
	return nil
}
