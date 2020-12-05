package ssh

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

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
