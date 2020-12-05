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

		if _, err := f.Write(v.Value); err != nil {
			return ErrUnableToWriteFile
		}

		if err := windowsChmod(v.File, 0600); err != nil {
			return err
		}

	}

	return nil
}

func windowsChmod(filePath string, fileMode os.FileMode) error {
	if err := acl.Chmod(filePath, fileMode); err != nil {
		return err
	}

	return nil
}
