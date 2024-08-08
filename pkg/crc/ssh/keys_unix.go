//go:build !windows
// +build !windows

package ssh

import (
	"os"
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
			Value: kp.PrivateKey,
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

		if err := f.Chmod(0600); err != nil {
			return err
		}

	}

	return nil
}
