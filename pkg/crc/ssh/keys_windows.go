package ssh

import (
	"encoding/pem"
	"os"

	"github.com/hectane/go-acl"
)

// WriteToFile writes keypair to files
func (kp *KeyPair) WriteToFile(privateKeyPath string, publicKeyPath string) error {
	files := []struct {
		File  string
		Type  string
		Value []byte
	}{
		{
			File:  privateKeyPath,
			Value: pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Headers: nil, Bytes: kp.PrivateKey}),
		},
		{
			File:  publicKeyPath,
			Value: kp.PublicKey,
		},
	}

	for _, v := range files {
		f, err := os.Create(v.File)
		if err != nil {
			return ErrUnableToWriteFile
		}
		defer f.Close()

		if _, err := f.Write(v.Value); err != nil {
			return ErrUnableToWriteFile
		}

		if err := acl.Chmod(v.File, 0600); err != nil {
			return err
		}
	}

	return nil
}
